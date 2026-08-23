package schedule

import (
	"context"
	"database/sql"
	"time"
)

type Section struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SectionUpdateFields struct {
	Name      *string
	SortOrder *int
	IsActive  *bool
}

type Shift struct {
	SectionID int64
	UserID    int64
	ShiftDate time.Time
	StartTime *string
	Code      *string
}

type Repository interface {
	ListSections(ctx context.Context, branchID int64) ([]*Section, error)
	CreateSection(ctx context.Context, branchID int64, name string) (*Section, error)
	UpdateSection(ctx context.Context, id int64, fields SectionUpdateFields) (*Section, error)
	DeleteSection(ctx context.Context, id int64) error

	ListShifts(ctx context.Context, branchID int64, from, to time.Time) ([]*Shift, error)
	UpsertShift(ctx context.Context, branchID, sectionID, userID int64, date time.Time, startTime, code *string) error

	FindNotes(ctx context.Context, branchID int64, weekStart time.Time) (string, error)
	UpsertNotes(ctx context.Context, branchID int64, weekStart time.Time, notes string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListSections(ctx context.Context, branchID int64) ([]*Section, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, branch_id, name, sort_order, is_active, created_at, updated_at
		from schedule_sections where branch_id = $1 and is_active = true
		order by sort_order, name
	`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []*Section
	for rows.Next() {
		var s Section
		if err := rows.Scan(&s.ID, &s.BranchID, &s.Name, &s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sections = append(sections, &s)
	}
	return sections, rows.Err()
}

func (r *repository) CreateSection(ctx context.Context, branchID int64, name string) (*Section, error) {
	s := &Section{BranchID: branchID, Name: name, IsActive: true}
	err := r.db.QueryRowContext(ctx, `
		insert into schedule_sections (branch_id, name) values ($1, $2)
		returning id, sort_order, is_active, created_at, updated_at
	`, branchID, name).Scan(&s.ID, &s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *repository) UpdateSection(ctx context.Context, id int64, fields SectionUpdateFields) (*Section, error) {
	if fields.Name != nil {
		if _, err := r.db.ExecContext(ctx, `update schedule_sections set name = $1, updated_at = now() where id = $2`, *fields.Name, id); err != nil {
			return nil, err
		}
	}
	if fields.SortOrder != nil {
		if _, err := r.db.ExecContext(ctx, `update schedule_sections set sort_order = $1, updated_at = now() where id = $2`, *fields.SortOrder, id); err != nil {
			return nil, err
		}
	}
	if fields.IsActive != nil {
		if _, err := r.db.ExecContext(ctx, `update schedule_sections set is_active = $1, updated_at = now() where id = $2`, *fields.IsActive, id); err != nil {
			return nil, err
		}
	}

	var s Section
	row := r.db.QueryRowContext(ctx, `
		select id, branch_id, name, sort_order, is_active, created_at, updated_at from schedule_sections where id = $1
	`, id)
	if err := row.Scan(&s.ID, &s.BranchID, &s.Name, &s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) DeleteSection(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `update schedule_sections set is_active = false, updated_at = now() where id = $1`, id)
	return err
}

func (r *repository) ListShifts(ctx context.Context, branchID int64, from, to time.Time) ([]*Shift, error) {
	rows, err := r.db.QueryContext(ctx, `
		select section_id, user_id, shift_date, start_time, code from schedule_shifts
		where branch_id = $1 and shift_date >= $2 and shift_date <= $3
	`, branchID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*Shift
	for rows.Next() {
		var sh Shift
		if err := rows.Scan(&sh.SectionID, &sh.UserID, &sh.ShiftDate, &sh.StartTime, &sh.Code); err != nil {
			return nil, err
		}
		shifts = append(shifts, &sh)
	}
	return shifts, rows.Err()
}

// UpsertShift deletes the cell when both startTime and code are empty, so
// clearing a shift in the grid removes the row instead of leaving a blank one.
func (r *repository) UpsertShift(ctx context.Context, branchID, sectionID, userID int64, date time.Time, startTime, code *string) error {
	if (startTime == nil || *startTime == "") && (code == nil || *code == "") {
		_, err := r.db.ExecContext(ctx, `
			delete from schedule_shifts where section_id = $1 and user_id = $2 and shift_date = $3
		`, sectionID, userID, date)
		return err
	}

	_, err := r.db.ExecContext(ctx, `
		insert into schedule_shifts (branch_id, section_id, user_id, shift_date, start_time, code)
		values ($1, $2, $3, $4, $5, $6)
		on conflict (section_id, user_id, shift_date)
		do update set start_time = excluded.start_time, code = excluded.code, updated_at = now()
	`, branchID, sectionID, userID, date, startTime, code)
	return err
}

func (r *repository) FindNotes(ctx context.Context, branchID int64, weekStart time.Time) (string, error) {
	var notes string
	err := r.db.QueryRowContext(ctx, `
		select notes from schedule_notes where branch_id = $1 and week_start_date = $2
	`, branchID, weekStart).Scan(&notes)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return notes, nil
}

func (r *repository) UpsertNotes(ctx context.Context, branchID int64, weekStart time.Time, notes string) error {
	_, err := r.db.ExecContext(ctx, `
		insert into schedule_notes (branch_id, week_start_date, notes)
		values ($1, $2, $3)
		on conflict (branch_id, week_start_date) do update set notes = excluded.notes, updated_at = now()
	`, branchID, weekStart, notes)
	return err
}
