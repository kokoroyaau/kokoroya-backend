package clock

import (
	"context"
	"database/sql"
	"time"
)

type TimeEntry struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	BranchID   int64      `json:"branch_id"`
	ClockInAt  time.Time  `json:"clock_in_at"`
	ClockOutAt *time.Time `json:"clock_out_at"`
}

type Repository interface {
	FindOpenByUser(ctx context.Context, userID int64) (*TimeEntry, error)
	Open(ctx context.Context, userID, branchID int64) (*TimeEntry, error)
	Close(ctx context.Context, id int64) (*TimeEntry, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const timeEntryColumns = `id, user_id, branch_id, clock_in_at, clock_out_at`

func scanTimeEntry(row *sql.Row) (*TimeEntry, error) {
	var e TimeEntry
	if err := row.Scan(&e.ID, &e.UserID, &e.BranchID, &e.ClockInAt, &e.ClockOutAt); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repository) FindOpenByUser(ctx context.Context, userID int64) (*TimeEntry, error) {
	row := r.db.QueryRowContext(ctx, `
		select `+timeEntryColumns+` from time_entries
		where user_id = $1 and clock_out_at is null
	`, userID)
	e, err := scanTimeEntry(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *repository) Open(ctx context.Context, userID, branchID int64) (*TimeEntry, error) {
	row := r.db.QueryRowContext(ctx, `
		insert into time_entries (user_id, branch_id, clock_in_at)
		values ($1, $2, now())
		returning `+timeEntryColumns+`
	`, userID, branchID)
	return scanTimeEntry(row)
}

func (r *repository) Close(ctx context.Context, id int64) (*TimeEntry, error) {
	row := r.db.QueryRowContext(ctx, `
		update time_entries set clock_out_at = now() where id = $1
		returning `+timeEntryColumns+`
	`, id)
	return scanTimeEntry(row)
}
