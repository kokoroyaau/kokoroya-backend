package foodcost

import (
	"context"
	"database/sql"
	"time"
)

type Supplier struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PurchaseEntry struct {
	SupplierID   int64
	PurchaseDate time.Time
	Amount       float64
}

type GrossSalesEntry struct {
	SalesDate time.Time
	Amount    float64
}

type SupplierUpdateFields struct {
	Name      *string
	SortOrder *int
	IsActive  *bool
}

type Repository interface {
	ListSuppliers(ctx context.Context, branchID int64) ([]*Supplier, error)
	CreateSupplier(ctx context.Context, branchID int64, name string) (*Supplier, error)
	UpdateSupplier(ctx context.Context, id int64, fields SupplierUpdateFields) (*Supplier, error)
	DeleteSupplier(ctx context.Context, id int64) error

	ListPurchaseEntries(ctx context.Context, branchID int64, from, to time.Time) ([]*PurchaseEntry, error)
	UpsertPurchaseEntry(ctx context.Context, branchID, supplierID int64, date time.Time, amount float64) error

	ListGrossSales(ctx context.Context, branchID int64, from, to time.Time) ([]*GrossSalesEntry, error)
	UpsertGrossSales(ctx context.Context, branchID int64, date time.Time, amount float64) error

	FindNetSalesRate(ctx context.Context, branchID int64, weekStart time.Time) (float64, bool, error)
	UpsertNetSalesRate(ctx context.Context, branchID int64, weekStart time.Time, rate float64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListSuppliers(ctx context.Context, branchID int64) ([]*Supplier, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, branch_id, name, sort_order, is_active, created_at, updated_at
		from suppliers where branch_id = $1 and is_active = true
		order by sort_order, name
	`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []*Supplier
	for rows.Next() {
		var s Supplier
		if err := rows.Scan(&s.ID, &s.BranchID, &s.Name, &s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, &s)
	}
	return suppliers, rows.Err()
}

func (r *repository) CreateSupplier(ctx context.Context, branchID int64, name string) (*Supplier, error) {
	s := &Supplier{BranchID: branchID, Name: name, IsActive: true}
	err := r.db.QueryRowContext(ctx, `
		insert into suppliers (branch_id, name) values ($1, $2)
		returning id, sort_order, is_active, created_at, updated_at
	`, branchID, name).Scan(&s.ID, &s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *repository) UpdateSupplier(ctx context.Context, id int64, fields SupplierUpdateFields) (*Supplier, error) {
	if fields.Name != nil {
		if _, err := r.db.ExecContext(ctx, `update suppliers set name = $1, updated_at = now() where id = $2`, *fields.Name, id); err != nil {
			return nil, err
		}
	}
	if fields.SortOrder != nil {
		if _, err := r.db.ExecContext(ctx, `update suppliers set sort_order = $1, updated_at = now() where id = $2`, *fields.SortOrder, id); err != nil {
			return nil, err
		}
	}
	if fields.IsActive != nil {
		if _, err := r.db.ExecContext(ctx, `update suppliers set is_active = $1, updated_at = now() where id = $2`, *fields.IsActive, id); err != nil {
			return nil, err
		}
	}

	var s Supplier
	row := r.db.QueryRowContext(ctx, `
		select id, branch_id, name, sort_order, is_active, created_at, updated_at from suppliers where id = $1
	`, id)
	if err := row.Scan(&s.ID, &s.BranchID, &s.Name, &s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) DeleteSupplier(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `update suppliers set is_active = false, updated_at = now() where id = $1`, id)
	return err
}

func (r *repository) ListPurchaseEntries(ctx context.Context, branchID int64, from, to time.Time) ([]*PurchaseEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		select supplier_id, purchase_date, amount from purchase_entries
		where branch_id = $1 and purchase_date >= $2 and purchase_date <= $3
	`, branchID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*PurchaseEntry
	for rows.Next() {
		var e PurchaseEntry
		if err := rows.Scan(&e.SupplierID, &e.PurchaseDate, &e.Amount); err != nil {
			return nil, err
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}

func (r *repository) UpsertPurchaseEntry(ctx context.Context, branchID, supplierID int64, date time.Time, amount float64) error {
	_, err := r.db.ExecContext(ctx, `
		insert into purchase_entries (branch_id, supplier_id, purchase_date, amount)
		values ($1, $2, $3, $4)
		on conflict (supplier_id, purchase_date) do update set amount = excluded.amount, updated_at = now()
	`, branchID, supplierID, date, amount)
	return err
}

func (r *repository) ListGrossSales(ctx context.Context, branchID int64, from, to time.Time) ([]*GrossSalesEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		select sales_date, amount from gross_sales_entries
		where branch_id = $1 and sales_date >= $2 and sales_date <= $3
	`, branchID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*GrossSalesEntry
	for rows.Next() {
		var e GrossSalesEntry
		if err := rows.Scan(&e.SalesDate, &e.Amount); err != nil {
			return nil, err
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}

func (r *repository) UpsertGrossSales(ctx context.Context, branchID int64, date time.Time, amount float64) error {
	_, err := r.db.ExecContext(ctx, `
		insert into gross_sales_entries (branch_id, sales_date, amount)
		values ($1, $2, $3)
		on conflict (branch_id, sales_date) do update set amount = excluded.amount, updated_at = now()
	`, branchID, date, amount)
	return err
}

func (r *repository) FindNetSalesRate(ctx context.Context, branchID int64, weekStart time.Time) (float64, bool, error) {
	var rate float64
	err := r.db.QueryRowContext(ctx, `
		select rate from weekly_net_sales_rates
		where branch_id = $1 and week_start_date <= $2
		order by week_start_date desc limit 1
	`, branchID, weekStart).Scan(&rate)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return rate, true, nil
}

func (r *repository) UpsertNetSalesRate(ctx context.Context, branchID int64, weekStart time.Time, rate float64) error {
	_, err := r.db.ExecContext(ctx, `
		insert into weekly_net_sales_rates (branch_id, week_start_date, rate)
		values ($1, $2, $3)
		on conflict (branch_id, week_start_date) do update set rate = excluded.rate, updated_at = now()
	`, branchID, weekStart, rate)
	return err
}
