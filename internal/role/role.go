package role

const (
	Owner    = "owner"
	Manager  = "manager"
	Employee = "employee"
)

func IsPrivileged(r string) bool {
	return r == Owner || r == Manager
}
