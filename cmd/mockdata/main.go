// Command mockdata seeds local test data on top of the owner account: two
// branches, a staff user with access to both (for the all-branches
// dashboard) and one scoped to a single branch, plus this week's clock
// entries, labour hours, and food-cost numbers so the dashboard and
// clock-entries pages have something to show. Re-running it wipes and
// re-inserts everything it owns, so it's safe to run repeatedly.
package main

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"kokoroya-backend/config"
	"kokoroya-backend/internal/database"
	"kokoroya-backend/pkg/logger"
)

const mockPassword = "password123"

// Prefixed distinctly so this never matches a real branch already sitting
// in the DB — seedBranches looks entries up by exact name before creating
// them.
var branchNames = []string{"[Mock] Branch A", "[Mock] Branch B"}

type mockUser struct {
	name         string
	email        string
	pin          string
	permissions  string // postgres array literal
	branches     []int  // indexes into branchNames
	employerName string
	employerABN  string
}

var mockUsers = []mockUser{
	{
		name:         "Multi Branch Staff",
		email:        "multi-staff@kokoroya.test",
		pin:          "9911",
		permissions:  "{dashboard,labour,food-cost,clock-in,salary}",
		branches:     []int{0, 1},
		employerName: "Backdoor Pty Ltd",
		employerABN:  "68 678 033 273",
	},
	{
		name:         "Branch A Staff",
		email:        "branch-a-staff@kokoroya.test",
		pin:          "9922",
		permissions:  "{dashboard,labour,food-cost,clock-in,salary}",
		branches:     []int{0},
		employerName: "Backdoor Pty Ltd",
		employerABN:  "68 678 033 273",
	},
}

const quarterHour = 15 * time.Minute

func roundedHours(in, out time.Time) float64 {
	d := out.Sub(in)
	rounded := (d + quarterHour/2) / quarterHour * quarterHour
	return rounded.Hours()
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg)

	db, err := database.NewPostgresConnection(cfg, log)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	branchIDs := seedBranches(db, log)
	userIDs := seedUsers(db, log)
	clearMockData(db, branchIDs)
	seedClockEntries(db, log, branchIDs, userIDs)
	seedFoodCost(db, log, branchIDs)

	log.Infof("mock data ready: branches=%v users=%v", branchIDs, userIDs)
}

func seedBranches(db *sql.DB, log interface{ Infof(string, ...any) }) map[string]int64 {
	ids := make(map[string]int64, len(branchNames))
	for _, name := range branchNames {
		var id int64
		err := db.QueryRow(`select id from branches where name = $1`, name).Scan(&id)
		if err == sql.ErrNoRows {
			err = db.QueryRow(`insert into branches (name) values ($1) returning id`, name).Scan(&id)
		}
		if err != nil {
			panic(fmt.Errorf("branch %s: %w", name, err))
		}
		ids[name] = id
	}
	log.Infof("branches: %v", ids)
	return ids
}

func seedUsers(db *sql.DB, log interface{ Infof(string, ...any) }) map[string]int64 {
	hash, err := bcrypt.GenerateFromPassword([]byte(mockPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	ids := make(map[string]int64, len(mockUsers))
	for _, u := range mockUsers {
		var id int64
		err := db.QueryRow(`
			insert into users (name, email, password_hash, role, is_active, permissions, pin, employer_name, employer_abn)
			values ($1, $2, $3, 'employee', true, $4, $5, $6, $7)
			on conflict (email) do update set
				permissions = excluded.permissions,
				pin = excluded.pin,
				employer_name = excluded.employer_name,
				employer_abn = excluded.employer_abn
			returning id
		`, u.name, u.email, string(hash), u.permissions, u.pin, u.employerName, u.employerABN).Scan(&id)
		if err != nil {
			panic(fmt.Errorf("user %s: %w", u.email, err))
		}
		ids[u.email] = id

		if _, err := db.Exec(`delete from user_branches where user_id = $1`, id); err != nil {
			panic(err)
		}
		for _, branchIdx := range u.branches {
			branchName := branchNames[branchIdx]
			if _, err := db.Exec(`
				insert into user_branches (user_id, branch_id)
				select $1, id from branches where name = $2
			`, id, branchName); err != nil {
				panic(err)
			}
		}
	}
	log.Infof("users: %v (password: %s)", ids, mockPassword)
	return ids
}

func clearMockData(db *sql.DB, branchIDs map[string]int64) {
	for _, id := range branchIDs {
		for _, table := range []string{"time_entries", "labour_hour_entries", "purchase_entries", "gross_sales_entries", "suppliers"} {
			if _, err := db.Exec(`delete from `+table+` where branch_id = $1`, id); err != nil {
				panic(fmt.Errorf("clear %s: %w", table, err))
			}
		}
	}
}

// seedClockEntries inserts this week's shifts: a closed shift each weekday
// for both staff at their branch(es), plus one still-open shift today, so
// the clock-entries edit UI has both a normal row and an "open" row to
// exercise.
func seedClockEntries(db *sql.DB, log interface{ Infof(string, ...any) }, branchIDs map[string]int64, userIDs map[string]int64) {
	multiStaffID := userIDs["multi-staff@kokoroya.test"]
	branchAStaffID := userIDs["branch-a-staff@kokoroya.test"]
	branchAID := branchIDs[branchNames[0]]
	branchBID := branchIDs[branchNames[1]]

	now := time.Now()
	monday := now.AddDate(0, 0, -int(now.Weekday()+6)%7)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())

	type shift struct {
		userID, branchID int64
		day              int // offset from monday
	}
	shifts := []shift{
		{multiStaffID, branchAID, 0},
		{multiStaffID, branchAID, 1},
		{multiStaffID, branchBID, 2},
		{branchAStaffID, branchAID, 0},
		{branchAStaffID, branchAID, 1},
		{branchAStaffID, branchAID, 3},
	}

	hoursByUserDate := make(map[string]float64)
	for _, s := range shifts {
		if s.day > int(now.Weekday()+6)%7 {
			continue // don't seed shifts in the future relative to "today"
		}
		day := monday.AddDate(0, 0, s.day)
		clockIn := day.Add(9 * time.Hour)
		clockOut := day.Add(17*time.Hour + 15*time.Minute)
		if _, err := db.Exec(`
			insert into time_entries (user_id, branch_id, clock_in_at, clock_out_at)
			values ($1, $2, $3, $4)
		`, s.userID, s.branchID, clockIn, clockOut); err != nil {
			panic(err)
		}
		key := fmt.Sprintf("%d|%d|%s", s.branchID, s.userID, day.Format("2006-01-02"))
		hoursByUserDate[key] += roundedHours(clockIn, clockOut)
	}

	// one open shift today, for the multi-branch staff at Branch A
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	if _, err := db.Exec(`
		insert into time_entries (user_id, branch_id, clock_in_at, clock_out_at)
		values ($1, $2, $3, null)
	`, multiStaffID, branchAID, todayStart); err != nil {
		panic(err)
	}

	for key, hours := range hoursByUserDate {
		var branchID, userID int64
		var dateStr string
		fmt.Sscanf(key, "%d|%d|%s", &branchID, &userID, &dateStr)
		if _, err := db.Exec(`
			insert into labour_hour_entries (branch_id, user_id, entry_date, total_hours)
			values ($1, $2, $3, $4)
		`, branchID, userID, dateStr, hours); err != nil {
			panic(err)
		}
	}

	log.Infof("seeded %d clock entries + 1 open shift", len(shifts))
}

// seedFoodCost gives both branches a supplier, this week's purchases, and
// daily gross sales so the dashboard/all-branches page has non-zero
// numbers.
func seedFoodCost(db *sql.DB, log interface{ Infof(string, ...any) }, branchIDs map[string]int64) {
	now := time.Now()
	monday := now.AddDate(0, 0, -int(now.Weekday()+6)%7)

	for name, branchID := range branchIDs {
		var supplierID int64
		if err := db.QueryRow(`
			insert into suppliers (branch_id, name) values ($1, 'Main Supplier') returning id
		`, branchID).Scan(&supplierID); err != nil {
			panic(err)
		}

		for d := range 7 {
			date := monday.AddDate(0, 0, d)
			if date.After(now) {
				break
			}
			if _, err := db.Exec(`
				insert into gross_sales_entries (branch_id, sales_date, amount) values ($1, $2, $3)
			`, branchID, date, 1200.0+float64(d)*50); err != nil {
				panic(err)
			}
			if _, err := db.Exec(`
				insert into purchase_entries (branch_id, supplier_id, purchase_date, amount) values ($1, $2, $3, $4)
			`, branchID, supplierID, date, 300.0+float64(d)*10); err != nil {
				panic(err)
			}
		}
		log.Infof("seeded food-cost data for %s", name)
	}
}
