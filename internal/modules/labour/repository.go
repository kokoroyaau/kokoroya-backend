package labour

import (
	"context"
	"database/sql"
	"time"
)

type HourEntry struct {
	UserID     int64
	EntryDate  time.Time
	TotalHours float64
}

type ShiftEntry struct {
	UserID     int64
	ClockInAt  time.Time
	ClockOutAt *time.Time
}

type Repository interface {
	ListHourEntries(ctx context.Context, branchID int64, from, to time.Time) ([]*HourEntry, error)
	ListShiftEntries(ctx context.Context, branchID int64, from, to time.Time) ([]*ShiftEntry, error)
	UpsertHourEntry(ctx context.Context, branchID, userID int64, date time.Time, hours float64) error
	AddHours(ctx context.Context, branchID, userID int64, date time.Time, deltaHours float64) error
	FindWeeklyRate(ctx context.Context, branchID int64, weekStart time.Time) (weekdayRate, weekendRate float64, ok bool, err error)
	UpsertWeeklyRate(ctx context.Context, branchID int64, weekStart time.Time, weekdayRate, weekendRate float64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListHourEntries(ctx context.Context, branchID int64, from, to time.Time) ([]*HourEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		select user_id, entry_date, total_hours from labour_hour_entries
		where branch_id = $1 and entry_date >= $2 and entry_date <= $3
	`, branchID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*HourEntry
	for rows.Next() {
		var e HourEntry
		if err := rows.Scan(&e.UserID, &e.EntryDate, &e.TotalHours); err != nil {
			return nil, err
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}

func (r *repository) ListShiftEntries(ctx context.Context, branchID int64, from, to time.Time) ([]*ShiftEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		select user_id, clock_in_at, clock_out_at from time_entries
		where branch_id = $1 and clock_in_at >= $2 and clock_in_at < $3
		order by clock_in_at
	`, branchID, from, to.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*ShiftEntry
	for rows.Next() {
		var e ShiftEntry
		if err := rows.Scan(&e.UserID, &e.ClockInAt, &e.ClockOutAt); err != nil {
			return nil, err
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}

func (r *repository) UpsertHourEntry(ctx context.Context, branchID, userID int64, date time.Time, hours float64) error {
	_, err := r.db.ExecContext(ctx, `
		insert into labour_hour_entries (branch_id, user_id, entry_date, total_hours)
		values ($1, $2, $3, $4)
		on conflict (branch_id, user_id, entry_date) do update set total_hours = excluded.total_hours, updated_at = now()
	`, branchID, userID, date, hours)
	return err
}

func (r *repository) AddHours(ctx context.Context, branchID, userID int64, date time.Time, deltaHours float64) error {
	_, err := r.db.ExecContext(ctx, `
		insert into labour_hour_entries (branch_id, user_id, entry_date, total_hours)
		values ($1, $2, $3, $4)
		on conflict (branch_id, user_id, entry_date) do update set total_hours = labour_hour_entries.total_hours + excluded.total_hours, updated_at = now()
	`, branchID, userID, date, deltaHours)
	return err
}

func (r *repository) FindWeeklyRate(ctx context.Context, branchID int64, weekStart time.Time) (float64, float64, bool, error) {
	var weekdayRate, weekendRate float64
	err := r.db.QueryRowContext(ctx, `
		select weekday_rate, weekend_rate from weekly_labour_rates
		where branch_id = $1 and week_start_date <= $2
		order by week_start_date desc limit 1
	`, branchID, weekStart).Scan(&weekdayRate, &weekendRate)
	if err == sql.ErrNoRows {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	return weekdayRate, weekendRate, true, nil
}

func (r *repository) UpsertWeeklyRate(ctx context.Context, branchID int64, weekStart time.Time, weekdayRate, weekendRate float64) error {
	_, err := r.db.ExecContext(ctx, `
		insert into weekly_labour_rates (branch_id, week_start_date, weekday_rate, weekend_rate)
		values ($1, $2, $3, $4)
		on conflict (branch_id, week_start_date) do update set weekday_rate = excluded.weekday_rate, weekend_rate = excluded.weekend_rate, updated_at = now()
	`, branchID, weekStart, weekdayRate, weekendRate)
	return err
}
