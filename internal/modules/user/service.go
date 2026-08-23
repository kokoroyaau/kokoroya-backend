package user

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"kokoroya-backend/internal/jwtauth"
	"kokoroya-backend/internal/session"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailExists = errors.New("email already exists")

type Service interface {
	Login(ctx context.Context, email, password string) (token string, expiresAt time.Time, role string, err error)
	Logout(ctx context.Context, jti string) error
	// CreateUser: email and password are optional — a PIN-only employee
	// (empty email/password) can clock in/out but never logs in.
	CreateUser(ctx context.Context, name, email, password, role, phone, tfn, pin string, rateWeekday, rateWeekend *float64, permissions []string, branchIDs []int64) (*User, error)
	UpdateUser(ctx context.Context, id int64, fields UpdateFields) (*User, error)
	DeleteUser(ctx context.Context, id int64) error
	SetPermissions(ctx context.Context, userID int64, permissions []string) error
	SetBranches(ctx context.Context, userID int64, branchIDs []int64) error
	List(ctx context.Context) ([]*User, error)
	Me(ctx context.Context, userID int64) (*User, error)
}

type service struct {
	repo           Repository
	jwtManager     *jwtauth.Manager
	sessionManager *session.Manager
	log            *logrus.Logger
}

func NewService(repo Repository, jwtManager *jwtauth.Manager, sessionManager *session.Manager, log *logrus.Logger) Service {
	return &service{repo: repo, jwtManager: jwtManager, sessionManager: sessionManager, log: log}
}

func (s *service) Login(ctx context.Context, email, password string) (string, time.Time, string, error) {
	u, err := s.repo.FindBy(ctx, Filter{Email: &email})
	if err != nil {
		s.log.WithError(err).WithField("email", email).Warn("user.Login: FindBy failed")
		return "", time.Time{}, "", ErrInvalidCredentials
	}

	if u.PasswordHash == nil {
		s.log.WithField("email", email).Warn("user.Login: account has no password (PIN-only user)")
		return "", time.Time{}, "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
		s.log.WithField("email", email).Warn("user.Login: wrong password")
		return "", time.Time{}, "", ErrInvalidCredentials
	}

	token, jti, expiresAt, err := s.jwtManager.Generate(u.ID, u.Role)
	if err != nil {
		s.log.WithError(err).WithField("user_id", u.ID).Error("user.Login: jwt generate failed")
		return "", time.Time{}, "", err
	}

	if err := s.sessionManager.Create(ctx, u.ID, jti, time.Until(expiresAt)); err != nil {
		s.log.WithError(err).WithField("user_id", u.ID).Error("user.Login: session create failed")
		return "", time.Time{}, "", err
	}

	return token, expiresAt, u.Role, nil
}

func (s *service) Logout(ctx context.Context, jti string) error {
	if err := s.sessionManager.Revoke(ctx, jti); err != nil {
		s.log.WithError(err).WithField("jti", jti).Error("user.Logout: session revoke failed")
		return err
	}
	return nil
}

func (s *service) CreateUser(ctx context.Context, name, email, password, role, phone, tfn, pin string, rateWeekday, rateWeekend *float64, permissions []string, branchIDs []int64) (*User, error) {
	u := &User{
		Name:        name,
		Role:        role,
		Permissions: permissions,
	}

	if email != "" {
		existing, err := s.repo.FindBy(ctx, Filter{Email: &email})
		if err == nil && existing != nil {
			s.log.WithField("email", email).Warn("user.CreateUser: email already exists")
			return nil, ErrEmailExists
		}
		u.Email = &email
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			s.log.WithError(err).Error("user.CreateUser: bcrypt hash failed")
			return nil, err
		}
		hashStr := string(hash)
		u.PasswordHash = &hashStr
	}
	if phone != "" {
		u.Phone = &phone
	}
	if tfn != "" {
		u.TFN = &tfn
	}
	if pin != "" {
		u.PIN = &pin
	}
	u.RateWeekday = rateWeekday
	u.RateWeekend = rateWeekend

	if err := s.repo.Create(ctx, u, branchIDs); err != nil {
		s.log.WithError(err).WithField("email", email).Error("user.CreateUser: repo create failed")
		return nil, err
	}
	return u, nil
}

func (s *service) UpdateUser(ctx context.Context, id int64, fields UpdateFields) (*User, error) {
	u, err := s.repo.Update(ctx, id, fields)
	if err != nil {
		s.log.WithError(err).WithField("user_id", id).Error("user.UpdateUser: repo update failed")
		return nil, err
	}
	return u, nil
}

func (s *service) DeleteUser(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.WithError(err).WithField("user_id", id).Error("user.DeleteUser: repo delete failed")
		return err
	}
	return nil
}

func (s *service) SetPermissions(ctx context.Context, userID int64, permissions []string) error {
	if err := s.repo.SetPermissions(ctx, userID, permissions); err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("user.SetPermissions: repo update failed")
		return err
	}
	return nil
}

func (s *service) SetBranches(ctx context.Context, userID int64, branchIDs []int64) error {
	if err := s.repo.SetBranches(ctx, userID, branchIDs); err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("user.SetBranches: repo update failed")
		return err
	}
	return nil
}

func (s *service) List(ctx context.Context) ([]*User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		s.log.WithError(err).Error("user.List: repo query failed")
		return nil, err
	}
	return users, nil
}

func (s *service) Me(ctx context.Context, userID int64) (*User, error) {
	u, err := s.repo.FindBy(ctx, Filter{ID: &userID})
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("user.Me: repo query failed")
		return nil, err
	}
	return u, nil
}
