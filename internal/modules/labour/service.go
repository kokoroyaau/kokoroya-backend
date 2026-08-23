package labour

import (
	"context"
	"time"

	"kokoroya-backend/internal/dateutil"
	"kokoroya-backend/internal/modules/branch"
)

const dateLayout = "2006-01-02"

// Fallback gross rate (already includes superannuation) used when a branch
// has never set its own weekly rate, matching the reference project's
// defaults (33.05 base wage x 1.12 super / 39.66 base wage x 1.12 super).
const (
	fallbackWeekdayRate = 37.02
	fallbackWeekendRate = 44.42
)

type ShiftEntryInfo struct {
	ClockInAt  time.Time  `json:"clock_in_at"`
	ClockOutAt *time.Time `json:"clock_out_at"`
}

type EmployeeWeekRow struct {
	UserID          int64                       `json:"user_id"`
	Name            string                      `json:"name"`
	DailyHours      map[string]float64          `json:"daily_hours"`
	DailyShifts     map[string][]ShiftEntryInfo `json:"daily_shifts"`
	TotalHours      float64                     `json:"total_hours"`
	PercentageOfAll float64                     `json:"percentage_of_all"`
	GrossPay        float64                     `json:"gross_pay"`
}

type LabourDayInfo struct {
	StaffCount int     `json:"staff_count"`
	TotalHours float64 `json:"total_hours"`
	LabourCost float64 `json:"labour_cost"`
	IsWeekend  bool    `json:"is_weekend"`
}

type Report struct {
	StartDate   string                   `json:"start_date"`
	EndDate     string                   `json:"end_date"`
	Employees   []EmployeeWeekRow        `json:"employees"`
	LabourDaily map[string]LabourDayInfo `json:"labour_daily"`
	LabourTotal float64                  `json:"labour_total"`
	WeekdayRate float64                  `json:"weekday_rate"`
	WeekendRate float64                  `json:"weekend_rate"`
}

type Service interface {
	GetReport(ctx context.Context, branchID int64, start, end time.Time) (*Report, error)
	UpsertHourEntry(ctx context.Context, branchID, userID int64, date time.Time, hours float64) error
	UpsertWeeklyRate(ctx context.Context, branchID int64, weekStart time.Time, weekdayRate, weekendRate float64) error
}

type service struct {
	repo       Repository
	branchRepo branch.Repository
}

func NewService(repo Repository, branchRepo branch.Repository) Service {
	return &service{repo: repo, branchRepo: branchRepo}
}

// rateForDay resolves the rate to charge for hours worked on date: the
// employee's own rate if they have one set, otherwise the branch's weekly
// gross rate, otherwise the hardcoded fallback.
func rateForDay(date time.Time, weekdayRate, weekendRate *float64, branchWeekdayRate, branchWeekendRate float64) float64 {
	isWeekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday

	if !isWeekend && weekdayRate != nil {
		return *weekdayRate
	}
	if isWeekend && weekendRate != nil {
		return *weekendRate
	}
	if isWeekend {
		return branchWeekendRate
	}
	return branchWeekdayRate
}

// branchRateResolver resolves the carry-forward weekly gross rate for any
// date in the report range, caching one lookup per week touched.
func (s *service) branchRateResolver(ctx context.Context, branchID int64) func(time.Time) (weekday, weekend float64, err error) {
	type rate struct{ weekday, weekend float64 }
	cache := make(map[string]rate)
	return func(date time.Time) (float64, float64, error) {
		monday := dateutil.MondayOf(date)
		key := monday.Format(dateLayout)
		if r, ok := cache[key]; ok {
			return r.weekday, r.weekend, nil
		}
		r := rate{fallbackWeekdayRate, fallbackWeekendRate}
		if wd, we, ok, err := s.repo.FindWeeklyRate(ctx, branchID, monday); err != nil {
			return 0, 0, err
		} else if ok {
			r = rate{wd, we}
		}
		cache[key] = r
		return r.weekday, r.weekend, nil
	}
}

func (s *service) GetReport(ctx context.Context, branchID int64, start, end time.Time) (*Report, error) {
	dates := make([]string, 0)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(dateLayout))
	}

	allEmployees, err := s.branchRepo.ListEmployees(ctx, branchID)
	if err != nil {
		return nil, err
	}
	employees := make([]*branch.Employee, 0, len(allEmployees))
	for _, e := range allEmployees {
		if e.Role == "owner" {
			continue
		}
		employees = append(employees, e)
	}

	entries, err := s.repo.ListHourEntries(ctx, branchID, start, end)
	if err != nil {
		return nil, err
	}
	hoursByUser := make(map[int64]map[string]float64)
	for _, e := range entries {
		if hoursByUser[e.UserID] == nil {
			hoursByUser[e.UserID] = make(map[string]float64)
		}
		hoursByUser[e.UserID][e.EntryDate.Format(dateLayout)] = e.TotalHours
	}

	shifts, err := s.repo.ListShiftEntries(ctx, branchID, start, end)
	if err != nil {
		return nil, err
	}
	shiftsByUser := make(map[int64]map[string][]ShiftEntryInfo)
	for _, sh := range shifts {
		if shiftsByUser[sh.UserID] == nil {
			shiftsByUser[sh.UserID] = make(map[string][]ShiftEntryInfo)
		}
		d := sh.ClockInAt.Format(dateLayout)
		shiftsByUser[sh.UserID][d] = append(shiftsByUser[sh.UserID][d], ShiftEntryInfo{
			ClockInAt:  sh.ClockInAt,
			ClockOutAt: sh.ClockOutAt,
		})
	}

	branchRate := s.branchRateResolver(ctx, branchID)
	displayWeekdayRate, displayWeekendRate, err := branchRate(start)
	if err != nil {
		return nil, err
	}

	rows := make([]EmployeeWeekRow, 0, len(employees))
	var weekTotalHours float64
	for _, employee := range employees {
		daily := make(map[string]float64, len(dates))
		dailyShifts := make(map[string][]ShiftEntryInfo, len(dates))
		var total, grossPay float64
		for _, d := range dates {
			hours := hoursByUser[employee.ID][d]
			daily[d] = hours
			total += hours
			shiftsForDay := shiftsByUser[employee.ID][d]
			if shiftsForDay == nil {
				shiftsForDay = []ShiftEntryInfo{}
			}
			dailyShifts[d] = shiftsForDay

			if hours > 0 {
				date, err := time.Parse(dateLayout, d)
				if err != nil {
					return nil, err
				}
				branchWeekday, branchWeekend, err := branchRate(date)
				if err != nil {
					return nil, err
				}
				grossPay += hours * rateForDay(date, employee.RateWeekday, employee.RateWeekend, branchWeekday, branchWeekend)
			}
		}
		weekTotalHours += total
		rows = append(rows, EmployeeWeekRow{
			UserID:      employee.ID,
			Name:        employee.Name,
			DailyHours:  daily,
			DailyShifts: dailyShifts,
			TotalHours:  total,
			GrossPay:    grossPay,
		})
	}
	for i := range rows {
		if weekTotalHours > 0 {
			rows[i].PercentageOfAll = rows[i].TotalHours / weekTotalHours * 100
		}
	}

	labourDaily := make(map[string]LabourDayInfo, len(dates))
	var labourTotal float64
	for _, d := range dates {
		date, err := time.Parse(dateLayout, d)
		if err != nil {
			return nil, err
		}
		branchWeekday, branchWeekend, err := branchRate(date)
		if err != nil {
			return nil, err
		}
		var dayHours float64
		var dayCost float64
		var staffCount int
		for _, employee := range employees {
			hours := hoursByUser[employee.ID][d]
			if hours <= 0 {
				continue
			}
			dayHours += hours
			staffCount++
			dayCost += hours * rateForDay(date, employee.RateWeekday, employee.RateWeekend, branchWeekday, branchWeekend)
		}
		labourDaily[d] = LabourDayInfo{
			StaffCount: staffCount,
			TotalHours: dayHours,
			LabourCost: dayCost,
			IsWeekend:  date.Weekday() == time.Saturday || date.Weekday() == time.Sunday,
		}
		labourTotal += dayCost
	}

	return &Report{
		StartDate:   start.Format(dateLayout),
		EndDate:     end.Format(dateLayout),
		Employees:   rows,
		LabourDaily: labourDaily,
		LabourTotal: labourTotal,
		WeekdayRate: displayWeekdayRate,
		WeekendRate: displayWeekendRate,
	}, nil
}

func (s *service) UpsertHourEntry(ctx context.Context, branchID, userID int64, date time.Time, hours float64) error {
	return s.repo.UpsertHourEntry(ctx, branchID, userID, date, hours)
}

func (s *service) UpsertWeeklyRate(ctx context.Context, branchID int64, weekStart time.Time, weekdayRate, weekendRate float64) error {
	return s.repo.UpsertWeeklyRate(ctx, branchID, weekStart, weekdayRate, weekendRate)
}
