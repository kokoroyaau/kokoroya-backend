package clock

import (
	"context"
	"errors"
	"time"

	"kokoroya-backend/internal/dateutil"
	"kokoroya-backend/internal/modules/labour"
	"kokoroya-backend/internal/modules/user"
)

var ErrInvalidPin = errors.New("invalid pin")
var ErrNotFound = errors.New("time entry not found")

const maxShiftDuration = 16 * time.Hour
const quarterHour = 15 * time.Minute

func roundedHours(in, out time.Time) float64 {
	d := out.Sub(in)
	rounded := (d + quarterHour/2) / quarterHour * quarterHour
	return rounded.Hours()
}

type PunchResult struct {
	Name   string
	Action string
	At     time.Time
	Hours  *float64
}

type Service interface {
	Punch(ctx context.Context, pin string, branchID int64) (*PunchResult, error)
	UpdateEntry(ctx context.Context, id, branchID int64, clockInAt time.Time, clockOutAt *time.Time) (*TimeEntry, error)
}

type service struct {
	repo       Repository
	userRepo   user.Repository
	labourRepo labour.Repository
}

func NewService(repo Repository, userRepo user.Repository, labourRepo labour.Repository) Service {
	return &service{repo: repo, userRepo: userRepo, labourRepo: labourRepo}
}

func (s *service) Punch(ctx context.Context, pin string, branchID int64) (*PunchResult, error) {
	u, err := s.userRepo.FindBy(ctx, user.Filter{PIN: &pin})
	if err != nil {
		return nil, ErrInvalidPin
	}

	open, err := s.repo.FindOpenByUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	if open != nil && time.Since(open.ClockInAt) > maxShiftDuration {
		if _, err := s.repo.Close(ctx, open.ID); err != nil {
			return nil, err
		}
		open = nil
	}

	if open != nil {
		closed, err := s.repo.Close(ctx, open.ID)
		if err != nil {
			return nil, err
		}

		hours := roundedHours(closed.ClockInAt, *closed.ClockOutAt)
		date := dateutil.DayOf(closed.ClockInAt)
		if err := s.labourRepo.AddHours(ctx, closed.BranchID, closed.UserID, date, hours); err != nil {
			return nil, err
		}

		return &PunchResult{Name: u.Name, Action: "out", At: *closed.ClockOutAt, Hours: &hours}, nil
	}

	opened, err := s.repo.Open(ctx, u.ID, branchID)
	if err != nil {
		return nil, err
	}
	return &PunchResult{Name: u.Name, Action: "in", At: opened.ClockInAt}, nil
}

// UpdateEntry corrects a time entry's clock-in/out and recomputes that
// employee's labour hours for every day touched (both the old and new date,
// when the edit moves the entry across days) from the raw time entries, so
// the stored total always matches what's actually on the clock rather than
// drifting via delta adjustments.
func (s *service) UpdateEntry(ctx context.Context, id, branchID int64, clockInAt time.Time, clockOutAt *time.Time) (*TimeEntry, error) {
	old, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if old == nil || old.BranchID != branchID {
		return nil, ErrNotFound
	}

	updated, err := s.repo.Update(ctx, id, clockInAt, clockOutAt)
	if err != nil {
		return nil, err
	}

	oldDate := dateutil.DayOf(old.ClockInAt)
	newDate := dateutil.DayOf(updated.ClockInAt)
	if err := s.recomputeDay(ctx, branchID, updated.UserID, oldDate); err != nil {
		return nil, err
	}
	if !newDate.Equal(oldDate) {
		if err := s.recomputeDay(ctx, branchID, updated.UserID, newDate); err != nil {
			return nil, err
		}
	}

	return updated, nil
}

// recomputeDay re-sums the employee's rounded hours for date from the raw
// time entries and overwrites the stored total, rather than adding a delta,
// so an edit can't leave the total out of sync with the clock.
func (s *service) recomputeDay(ctx context.Context, branchID, userID int64, date time.Time) error {
	shifts, err := s.labourRepo.ListShiftEntries(ctx, branchID, date, date)
	if err != nil {
		return err
	}

	var total float64
	for _, sh := range shifts {
		if sh.UserID != userID || sh.ClockOutAt == nil {
			continue
		}
		total += roundedHours(sh.ClockInAt, *sh.ClockOutAt)
	}

	return s.labourRepo.UpsertHourEntry(ctx, branchID, userID, date, total)
}
