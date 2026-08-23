package schema

type CreateScheduleSectionRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateScheduleSectionRequest struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

type UpsertShiftRequest struct {
	SectionID int64   `json:"section_id" binding:"required"`
	UserID    int64   `json:"user_id" binding:"required"`
	ShiftDate string  `json:"shift_date" binding:"required"`
	StartTime *string `json:"start_time"`
	Code      *string `json:"code" binding:"omitempty,oneof=C F S FS B TOILET"`
}

type UpsertScheduleNotesRequest struct {
	WeekStartDate string `json:"week_start_date" binding:"required"`
	Notes         string `json:"notes"`
}
