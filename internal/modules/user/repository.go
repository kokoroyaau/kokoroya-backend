package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// User is a row in the users table.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        *string   `json:"email"`
	PasswordHash *string   `json:"-"`
	Role         string    `json:"role"`
	Phone        *string   `json:"phone"`
	TFN          *string   `json:"tfn"`
	PIN          *string   `json:"pin,omitempty"`
	IsActive     bool      `json:"is_active"`
	Permissions  []string  `json:"permissions"`
	BranchIDs    []int64   `json:"branch_ids,omitempty"`
	RateWeekday  *float64  `json:"rate_weekday"`
	RateWeekend  *float64  `json:"rate_weekend"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Filter struct {
	ID    *int64
	Email *string
	PIN   *string
}

type UpdateFields struct {
	Name        *string
	Email       *string
	Phone       *string
	TFN         *string
	PIN         *string
	Role        *string
	IsActive    *bool
	RateWeekday *float64
	RateWeekend *float64
}

type Repository interface {
	FindBy(ctx context.Context, filter Filter) (*User, error)
	Create(ctx context.Context, u *User, branchIDs []int64) error
	Update(ctx context.Context, id int64, fields UpdateFields) (*User, error)
	Delete(ctx context.Context, id int64) error
	SetPermissions(ctx context.Context, id int64, permissions []string) error
	SetBranches(ctx context.Context, userID int64, branchIDs []int64) error
	List(ctx context.Context) ([]*User, error)
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new user Repository.
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const userColumns = `id, name, email, password_hash, role, phone, tfn, pin, is_active, permissions, rate_weekday, rate_weekend, created_at, updated_at`

func scanUser(row *sql.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Phone, &u.TFN, &u.PIN, &u.IsActive, pq.Array(&u.Permissions), &u.RateWeekday, &u.RateWeekend, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindBy looks up a user by whichever field is set on filter (ID takes
// precedence if both are somehow set).
func (r *repository) FindBy(ctx context.Context, filter Filter) (*User, error) {
	switch {
	case filter.ID != nil:
		row := r.db.QueryRowContext(ctx, `select `+userColumns+` from users where id = $1`, *filter.ID)
		return scanUser(row)
	case filter.Email != nil:
		row := r.db.QueryRowContext(ctx, `select `+userColumns+` from users where email = $1`, *filter.Email)
		return scanUser(row)
	case filter.PIN != nil:
		row := r.db.QueryRowContext(ctx, `select `+userColumns+` from users where pin = $1`, *filter.PIN)
		return scanUser(row)
	default:
		return nil, errors.New("user.FindBy: no filter field set")
	}
}

func (r *repository) Create(ctx context.Context, u *User, branchIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := tx.QueryRowContext(ctx, `
		insert into users (name, email, password_hash, role, phone, tfn, pin, is_active, permissions, rate_weekday, rate_weekend)
		values ($1, $2, $3, $4, $5, $6, $7, true, $8, $9, $10)
		returning id, created_at, updated_at
	`, u.Name, u.Email, u.PasswordHash, u.Role, u.Phone, u.TFN, u.PIN, pq.Array(u.Permissions), u.RateWeekday, u.RateWeekend,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return err
	}

	if err := insertUserBranches(ctx, tx, u.ID, branchIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) SetPermissions(ctx context.Context, id int64, permissions []string) error {
	_, err := r.db.ExecContext(ctx, `update users set permissions = $1, updated_at = now() where id = $2`, pq.Array(permissions), id)
	return err
}

func (r *repository) SetBranches(ctx context.Context, userID int64, branchIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `delete from user_branches where user_id = $1`, userID); err != nil {
		return err
	}

	if err := insertUserBranches(ctx, tx, userID, branchIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func insertUserBranches(ctx context.Context, tx *sql.Tx, userID int64, branchIDs []int64) error {
	for _, branchID := range branchIDs {
		if _, err := tx.ExecContext(ctx, `insert into user_branches (user_id, branch_id) values ($1, $2)`, userID, branchID); err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) Update(ctx context.Context, id int64, fields UpdateFields) (*User, error) {
	if fields.Name != nil {
		if _, err := r.db.ExecContext(ctx, `update users set name = $1, updated_at = now() where id = $2`, *fields.Name, id); err != nil {
			return nil, err
		}
	}
	if fields.Email != nil {
		if _, err := r.db.ExecContext(ctx, `update users set email = $1, updated_at = now() where id = $2`, *fields.Email, id); err != nil {
			return nil, err
		}
	}
	if fields.Phone != nil {
		if _, err := r.db.ExecContext(ctx, `update users set phone = $1, updated_at = now() where id = $2`, *fields.Phone, id); err != nil {
			return nil, err
		}
	}
	if fields.TFN != nil {
		if _, err := r.db.ExecContext(ctx, `update users set tfn = $1, updated_at = now() where id = $2`, *fields.TFN, id); err != nil {
			return nil, err
		}
	}
	if fields.PIN != nil {
		if _, err := r.db.ExecContext(ctx, `update users set pin = $1, updated_at = now() where id = $2`, *fields.PIN, id); err != nil {
			return nil, err
		}
	}
	if fields.Role != nil {
		if _, err := r.db.ExecContext(ctx, `update users set role = $1, updated_at = now() where id = $2`, *fields.Role, id); err != nil {
			return nil, err
		}
	}
	if fields.IsActive != nil {
		if _, err := r.db.ExecContext(ctx, `update users set is_active = $1, updated_at = now() where id = $2`, *fields.IsActive, id); err != nil {
			return nil, err
		}
	}
	if fields.RateWeekday != nil {
		if _, err := r.db.ExecContext(ctx, `update users set rate_weekday = $1, updated_at = now() where id = $2`, *fields.RateWeekday, id); err != nil {
			return nil, err
		}
	}
	if fields.RateWeekend != nil {
		if _, err := r.db.ExecContext(ctx, `update users set rate_weekend = $1, updated_at = now() where id = $2`, *fields.RateWeekend, id); err != nil {
			return nil, err
		}
	}

	row := r.db.QueryRowContext(ctx, `select `+userColumns+` from users where id = $1`, id)
	return scanUser(row)
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `update users set is_active = false, updated_at = now() where id = $1`, id)
	return err
}

func (r *repository) List(ctx context.Context) ([]*User, error) {
	rows, err := r.db.QueryContext(ctx, `select `+userColumns+` from users order by id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Phone, &u.TFN, &u.PIN, &u.IsActive, pq.Array(&u.Permissions), &u.RateWeekday, &u.RateWeekend, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	branchRows, err := r.db.QueryContext(ctx, `select user_id, array_agg(branch_id) from user_branches group by user_id`)
	if err != nil {
		return nil, err
	}
	defer branchRows.Close()

	branchesByUser := make(map[int64][]int64)
	for branchRows.Next() {
		var userID int64
		var branchIDs []int64
		if err := branchRows.Scan(&userID, pq.Array(&branchIDs)); err != nil {
			return nil, err
		}
		branchesByUser[userID] = branchIDs
	}
	if err := branchRows.Err(); err != nil {
		return nil, err
	}

	for _, u := range users {
		u.BranchIDs = branchesByUser[u.ID]
	}
	return users, nil
}
