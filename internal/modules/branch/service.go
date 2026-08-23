package branch

import "context"

type Service interface {
	List(ctx context.Context) ([]*Branch, error)
	ListForUser(ctx context.Context, userID int64) ([]*Branch, error)
	Create(ctx context.Context, name string) (*Branch, error)
	Update(ctx context.Context, id int64, name *string, isActive *bool) (*Branch, error)
	Delete(ctx context.Context, id int64) error
	ListEmployees(ctx context.Context, branchID int64) ([]*Employee, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) ([]*Branch, error) {
	return s.repo.List(ctx)
}

func (s *service) ListForUser(ctx context.Context, userID int64) ([]*Branch, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *service) Create(ctx context.Context, name string) (*Branch, error) {
	b := &Branch{Name: name}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *service) Update(ctx context.Context, id int64, name *string, isActive *bool) (*Branch, error) {
	return s.repo.Update(ctx, id, name, isActive)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) ListEmployees(ctx context.Context, branchID int64) ([]*Employee, error) {
	return s.repo.ListEmployees(ctx, branchID)
}
