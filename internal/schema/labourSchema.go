package schema

type UpsertHourEntryRequest struct {
	UserID     int64   `json:"user_id" binding:"required"`
	EntryDate  string  `json:"entry_date" binding:"required"`
	TotalHours float64 `json:"total_hours" binding:"min=0"`
}

type UpsertLabourRateRequest struct {
	WeekStartDate string  `json:"week_start_date" binding:"required"`
	WeekdayRate   float64 `json:"weekday_rate" binding:"min=0"`
	WeekendRate   float64 `json:"weekend_rate" binding:"min=0"`
}
