package foodcost

import (
	"context"
	"time"

	"kokoroya-backend/internal/dateutil"
)

const dateLayout = "2006-01-02"

// fallbackNetSalesRate divides out 10% GST (net = gross / 1.10) when a
// branch never set its own weekly net sales rate.
const fallbackNetSalesRate = 1 / 1.10

type SupplierWeekRow struct {
	SupplierID      int64              `json:"supplier_id"`
	SupplierName    string             `json:"supplier_name"`
	DailyAmounts    map[string]float64 `json:"daily_amounts"`
	Total           float64            `json:"total"`
	PercentageOfAll float64            `json:"percentage_of_all"`
}

type Report struct {
	StartDate          string             `json:"start_date"`
	EndDate            string             `json:"end_date"`
	Suppliers          []SupplierWeekRow  `json:"suppliers"`
	GrandTotalPurchase float64            `json:"grand_total_purchase"`
	GrossSalesDaily    map[string]float64 `json:"gross_sales_daily"`
	GrossSalesTotal    float64            `json:"gross_sales_total"`
	NetSales           float64            `json:"net_sales"`
	NetSalesRate       float64            `json:"net_sales_rate"`
	PurchaseRatioPct   float64            `json:"purchase_ratio_pct"`
}

type Service interface {
	GetReport(ctx context.Context, branchID int64, start, end time.Time) (*Report, error)
	UpsertPurchaseEntry(ctx context.Context, branchID, supplierID int64, date time.Time, amount float64) error
	UpsertGrossSales(ctx context.Context, branchID int64, date time.Time, amount float64) error
	UpsertNetSalesRate(ctx context.Context, branchID int64, weekStart time.Time, rate float64) error
	ListSuppliers(ctx context.Context, branchID int64) ([]*Supplier, error)
	CreateSupplier(ctx context.Context, branchID int64, name string) (*Supplier, error)
	UpdateSupplier(ctx context.Context, id int64, fields SupplierUpdateFields) (*Supplier, error)
	DeleteSupplier(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// netSalesRateResolver resolves the carry-forward weekly net sales rate for
// any date in the report range, caching one lookup per week touched.
func (s *service) netSalesRateResolver(ctx context.Context, branchID int64) func(time.Time) (float64, error) {
	cache := make(map[string]float64)
	return func(date time.Time) (float64, error) {
		monday := dateutil.MondayOf(date)
		key := monday.Format(dateLayout)
		if rate, ok := cache[key]; ok {
			return rate, nil
		}
		rate := fallbackNetSalesRate
		if r, ok, err := s.repo.FindNetSalesRate(ctx, branchID, monday); err != nil {
			return 0, err
		} else if ok {
			rate = r
		}
		cache[key] = rate
		return rate, nil
	}
}

func (s *service) GetReport(ctx context.Context, branchID int64, start, end time.Time) (*Report, error) {
	dates := make([]string, 0)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(dateLayout))
	}

	suppliers, err := s.repo.ListSuppliers(ctx, branchID)
	if err != nil {
		return nil, err
	}

	entries, err := s.repo.ListPurchaseEntries(ctx, branchID, start, end)
	if err != nil {
		return nil, err
	}
	entriesBySupplier := make(map[int64]map[string]float64)
	for _, e := range entries {
		if entriesBySupplier[e.SupplierID] == nil {
			entriesBySupplier[e.SupplierID] = make(map[string]float64)
		}
		entriesBySupplier[e.SupplierID][e.PurchaseDate.Format(dateLayout)] = e.Amount
	}

	grossEntries, err := s.repo.ListGrossSales(ctx, branchID, start, end)
	if err != nil {
		return nil, err
	}
	grossByDate := make(map[string]float64)
	for _, g := range grossEntries {
		grossByDate[g.SalesDate.Format(dateLayout)] = g.Amount
	}

	rateForDate := s.netSalesRateResolver(ctx, branchID)
	displayRate, err := rateForDate(start)
	if err != nil {
		return nil, err
	}

	rows := make([]SupplierWeekRow, 0, len(suppliers))
	var grandTotal float64
	for _, supplier := range suppliers {
		daily := make(map[string]float64, len(dates))
		var total float64
		for _, d := range dates {
			amount := entriesBySupplier[supplier.ID][d]
			daily[d] = amount
			total += amount
		}
		grandTotal += total
		rows = append(rows, SupplierWeekRow{
			SupplierID:   supplier.ID,
			SupplierName: supplier.Name,
			DailyAmounts: daily,
			Total:        total,
		})
	}
	for i := range rows {
		if grandTotal > 0 {
			rows[i].PercentageOfAll = rows[i].Total / grandTotal * 100
		}
	}

	grossFilled := make(map[string]float64, len(dates))
	var grossTotal, netSales float64
	for _, d := range dates {
		amount := grossByDate[d]
		grossFilled[d] = amount
		grossTotal += amount

		date, err := time.Parse(dateLayout, d)
		if err != nil {
			return nil, err
		}
		rate, err := rateForDate(date)
		if err != nil {
			return nil, err
		}
		netSales += amount * rate
	}

	var purchaseRatioPct float64
	if netSales > 0 {
		purchaseRatioPct = grandTotal / netSales * 100
	}

	return &Report{
		StartDate:          start.Format(dateLayout),
		EndDate:            end.Format(dateLayout),
		Suppliers:          rows,
		GrandTotalPurchase: grandTotal,
		GrossSalesDaily:    grossFilled,
		GrossSalesTotal:    grossTotal,
		NetSales:           netSales,
		NetSalesRate:       displayRate,
		PurchaseRatioPct:   purchaseRatioPct,
	}, nil
}

func (s *service) UpsertPurchaseEntry(ctx context.Context, branchID, supplierID int64, date time.Time, amount float64) error {
	return s.repo.UpsertPurchaseEntry(ctx, branchID, supplierID, date, amount)
}

func (s *service) UpsertGrossSales(ctx context.Context, branchID int64, date time.Time, amount float64) error {
	return s.repo.UpsertGrossSales(ctx, branchID, date, amount)
}

func (s *service) UpsertNetSalesRate(ctx context.Context, branchID int64, weekStart time.Time, rate float64) error {
	return s.repo.UpsertNetSalesRate(ctx, branchID, weekStart, rate)
}

func (s *service) ListSuppliers(ctx context.Context, branchID int64) ([]*Supplier, error) {
	return s.repo.ListSuppliers(ctx, branchID)
}

func (s *service) CreateSupplier(ctx context.Context, branchID int64, name string) (*Supplier, error) {
	return s.repo.CreateSupplier(ctx, branchID, name)
}

func (s *service) UpdateSupplier(ctx context.Context, id int64, fields SupplierUpdateFields) (*Supplier, error) {
	return s.repo.UpdateSupplier(ctx, id, fields)
}

func (s *service) DeleteSupplier(ctx context.Context, id int64) error {
	return s.repo.DeleteSupplier(ctx, id)
}
