package schema

type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
	// Email/Password are optional — a PIN-only employee (no email/password)
	// can clock in/out but never logs in.
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=8"`
	Role     string `json:"role" binding:"required"`
	Phone    string `json:"phone"`
	TFN      string `json:"tfn"`
	// EmployerName/EmployerABN are the legal entity shown on this employee's
	// payslip — kept per-employee, not per-branch, since staff at the same
	// branch can be employed under different entities.
	EmployerName string   `json:"employer_name"`
	EmployerABN  string   `json:"employer_abn"`
	Pin          *string  `json:"pin" binding:"omitempty,len=4,numeric"`
	RateWeekday  *float64 `json:"rate_weekday" binding:"omitempty,min=0"`
	RateWeekend  *float64 `json:"rate_weekend" binding:"omitempty,min=0"`
	Permissions  []string `json:"permissions"`
	BranchIDs    []int64  `json:"branch_ids"`
}

type UpdateUserRequest struct {
	Name           *string  `json:"name"`
	Email          *string  `json:"email"`
	Phone          *string  `json:"phone"`
	TFN            *string  `json:"tfn"`
	EmployerName   *string  `json:"employer_name"`
	EmployerABN    *string  `json:"employer_abn"`
	Pin            *string  `json:"pin" binding:"omitempty,len=4,numeric"`
	Role           *string  `json:"role"`
	IsActive       *bool    `json:"is_active"`
	RateWeekday    *float64 `json:"rate_weekday" binding:"omitempty,min=0"`
	RateWeekend    *float64 `json:"rate_weekend" binding:"omitempty,min=0"`
	HourCapWeekday *float64 `json:"hour_cap_weekday" binding:"omitempty,min=0"`
	HourCapWeekend *float64 `json:"hour_cap_weekend" binding:"omitempty,min=0"`
}

type SetPermissionsRequest struct {
	Permissions []string `json:"permissions" binding:"required"`
}

type SetBranchesRequest struct {
	BranchIDs []int64 `json:"branch_ids" binding:"required"`
}

type MeResponse struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type PermissionsResponse struct {
	Pages []string `json:"pages"`
}
