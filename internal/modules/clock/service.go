package clock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"kokoroya-backend/internal/dateutil"
	"kokoroya-backend/internal/email"
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

func snapClockOut(t, clockInAt time.Time) time.Time {
	floored := t.Truncate(quarterHour)
	if floored.Before(clockInAt) {
		return clockInAt
	}
	return floored
}

type PunchResult struct {
	Name   string
	Action string
	At     time.Time
	Hours  *float64
}

type Service interface {
	Punch(ctx context.Context, pin string, branchID int64) (*PunchResult, error)
	UpdateEntry(ctx context.Context, id, branchID, actorUserID int64, clockInAt time.Time, clockOutAt *time.Time) (*TimeEntry, error)
	CreateEntry(ctx context.Context, branchID, userID, actorUserID int64, clockInAt time.Time, clockOutAt *time.Time) (*TimeEntry, error)
	DeleteEntry(ctx context.Context, id, branchID, actorUserID int64) error
}

type service struct {
	repo         Repository
	userRepo     user.Repository
	labourRepo   labour.Repository
	emailService email.Service
	notifyEmail  string
	log          *logrus.Logger
}

func NewService(repo Repository, userRepo user.Repository, labourRepo labour.Repository, emailService email.Service, notifyEmail string, log *logrus.Logger) Service {
	return &service{
		repo:         repo,
		userRepo:     userRepo,
		labourRepo:   labourRepo,
		emailService: emailService,
		notifyEmail:  notifyEmail,
		log:          log,
	}
}

func (s *service) describeUser(ctx context.Context, userID int64, preferEmail bool) string {
	u, err := s.userRepo.FindBy(ctx, user.Filter{ID: &userID})
	if err != nil || u == nil {
		return fmt.Sprintf("user #%d", userID)
	}
	if preferEmail && u.Email != nil && *u.Email != "" {
		return *u.Email
	}
	return u.Name
}

func (s *service) notifyEdit(ctx context.Context, actorUserID, employeeUserID int64, action, detail string) {
	if s.notifyEmail == "" {
		return
	}
	actor := s.describeUser(ctx, actorUserID, true)
	employee := s.describeUser(ctx, employeeUserID, false)

	subject := fmt.Sprintf("Clock entry %s - %s", action, employee)
	body := fmt.Sprintf(
		"<p><strong>%s</strong> is editing %s's clock entry (%s).</p><p>%s</p>",
		actor, employee, action, detail,
	)
	go func() {
		if err := s.emailService.Send(context.Background(), s.notifyEmail, subject, body); err != nil {
			s.log.WithError(err).Warn("clock: failed to send owner notification email")
		}
	}()
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
		if _, err := s.repo.Close(ctx, open.ID, snapClockOut(time.Now(), open.ClockInAt)); err != nil {
			return nil, err
		}
		open = nil
	}

	if open != nil && open.BranchID != branchID {
		return nil, ErrOpenAtOtherBranch
	}

	if open != nil {
		closed, err := s.repo.Close(ctx, open.ID, snapClockOut(time.Now(), open.ClockInAt))
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

func (s *service) UpdateEntry(ctx context.Context, id, branchID, actorUserID int64, clockInAt time.Time, clockOutAt *time.Time) (*TimeEntry, error) {
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

	s.notifyEdit(ctx, actorUserID, updated.UserID, "entry updated", fmt.Sprintf(
		"Clock in %s, clock out %s",
		updated.ClockInAt.Format(time.RFC1123),
		formatOptionalTime(updated.ClockOutAt),
	))
	return updated, nil
}

func (s *service) CreateEntry(ctx context.Context, branchID, userID, actorUserID int64, clockInAt time.Time, clockOutAt *time.Time) (*TimeEntry, error) {
	created, err := s.repo.Create(ctx, userID, branchID, clockInAt, clockOutAt)
	if err != nil {
		return nil, err
	}

	if err := s.recomputeDay(ctx, branchID, userID, dateutil.DayOf(created.ClockInAt)); err != nil {
		return nil, err
	}

	s.notifyEdit(ctx, actorUserID, userID, "entry added", fmt.Sprintf(
		"Clock in %s, clock out %s",
		created.ClockInAt.Format(time.RFC1123),
		formatOptionalTime(created.ClockOutAt),
	))
	return created, nil
}

func (s *service) DeleteEntry(ctx context.Context, id, branchID, actorUserID int64) error {
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

	if err := s.recomputeDay(ctx, branchID, old.UserID, dateutil.DayOf(old.ClockInAt)); err != nil {
		return err
	}

	s.notifyEdit(ctx, actorUserID, old.UserID, "entry deleted", fmt.Sprintf(
		"Clock in %s, clock out %s",
		old.ClockInAt.Format(time.RFC1123),
		formatOptionalTime(old.ClockOutAt),
	))
	return nil
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return "(open)"
	}
	return t.Format(time.RFC1123)
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
