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
var ErrOpenAtOtherBranch = errors.New("already clocked in at another branch")

const maxShiftDuration = 16 * time.Hour
const quarterHour = 15 * time.Minute
const graceLateness = 3 * time.Minute

func roundedHours(in, out time.Time) float64 {
	d := out.Sub(in)
	rounded := (d + quarterHour/2) / quarterHour * quarterHour
	return rounded.Hours()
}

// snapClockIn rounds a clock-in time to the 15-minute grid: within the first
// graceLateness minutes of a block it snaps down (on time), otherwise it
// snaps up to the next block (counted as late).
func snapClockIn(t time.Time) time.Time {
	floor := t.Truncate(quarterHour)
	if t.Sub(floor) <= graceLateness {
		return floor
	}
	return floor.Add(quarterHour)
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
	CreateEntry(ctx context.Context, branchID, userID int64, clockInAt time.Time, clockOutAt *time.Time) (*TimeEntry, error)
	DeleteEntry(ctx context.Context, id, branchID int64) error
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

	if open != nil && open.BranchID != branchID {
		return nil, ErrOpenAtOtherBranch
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

	opened, err := s.repo.Open(ctx, u.ID, branchID, snapClockIn(time.Now()))
	if err != nil {
		return nil, err
	}
	return &PunchResult{Name: u.Name, Action: "in", At: opened.ClockInAt}, nil
}

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

func (s *service) CreateEntry(ctx context.Context, branchID, userID int64, clockInAt time.Time, clockOutAt *time.Time) (*TimeEntry, error) {
	created, err := s.repo.Create(ctx, userID, branchID, clockInAt, clockOutAt)
	if err != nil {
		return nil, err
	}

	if err := s.recomputeDay(ctx, branchID, userID, dateutil.DayOf(created.ClockInAt)); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *service) DeleteEntry(ctx context.Context, id, branchID int64) error {
	old, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if old == nil || old.BranchID != branchID {
		return ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return s.recomputeDay(ctx, branchID, old.UserID, dateutil.DayOf(old.ClockInAt))
}

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
