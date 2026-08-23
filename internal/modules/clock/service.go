package clock

import (
	"context"
	"errors"
	"time"

	"kokoroya-backend/internal/modules/labour"
	"kokoroya-backend/internal/modules/user"
)

var ErrInvalidPin = errors.New("invalid pin")

const maxShiftDuration = 16 * time.Hour

type PunchResult struct {
	Name   string
	Action string
	At     time.Time
}

type Service interface {
	Punch(ctx context.Context, pin string, branchID int64) (*PunchResult, error)
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

		hours := closed.ClockOutAt.Sub(closed.ClockInAt).Hours()
		date := time.Date(closed.ClockInAt.Year(), closed.ClockInAt.Month(), closed.ClockInAt.Day(), 0, 0, 0, 0, closed.ClockInAt.Location())
		if err := s.labourRepo.AddHours(ctx, closed.BranchID, closed.UserID, date, hours); err != nil {
			return nil, err
		}

		return &PunchResult{Name: u.Name, Action: "out", At: *closed.ClockOutAt}, nil
	}

	opened, err := s.repo.Open(ctx, u.ID, branchID)
	if err != nil {
		return nil, err
	}
	return &PunchResult{Name: u.Name, Action: "in", At: opened.ClockInAt}, nil
}
