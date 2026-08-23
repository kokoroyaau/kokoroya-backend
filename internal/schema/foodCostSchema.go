package schema

type CreateSupplierRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateSupplierRequest struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

type UpsertPurchaseEntryRequest struct {
	SupplierID   int64   `json:"supplier_id" binding:"required"`
	PurchaseDate string  `json:"purchase_date" binding:"required"`
	Amount       float64 `json:"amount" binding:"min=0"`
}

type UpsertGrossSalesRequest struct {
	SalesDate string  `json:"sales_date" binding:"required"`
	Amount    float64 `json:"amount" binding:"min=0"`
}

type UpsertNetSalesRateRequest struct {
	WeekStartDate string  `json:"week_start_date" binding:"required"`
	Rate          float64 `json:"rate" binding:"min=0"`
}
