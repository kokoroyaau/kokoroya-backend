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
	ID         int64      `json:"id"`
	ClockInAt  time.Time  `json:"clock_in_at"`
	ClockOutAt *time.Time `json:"clock_out_at"`
}

// PayBreakdown is one payslip line: hours worked in a category (weekday,
// Saturday, or Sunday), the effective rate for those hours, and the total.
// Rate is derived as Total/Hours rather than looked up separately, so it
// stays exact even if the resolved rate varied within the period (e.g. a
// weekly rate change mid-fortnight) — Total always equals Hours * Rate.
type PayBreakdown struct {
	Hours float64 `json:"hours"`
	Rate  float64 `json:"rate"`
	Total float64 `json:"total"`
}

type EmployeeWeekRow struct {
	UserID          int64                       `json:"user_id"`
	Name            string                      `json:"name"`
	EmployerName    *string                     `json:"employer_name"`
	EmployerABN     *string                     `json:"employer_abn"`
	DailyHours      map[string]float64          `json:"daily_hours"`
	DailyShifts     map[string][]ShiftEntryInfo `json:"daily_shifts"`
	Weekday         PayBreakdown                `json:"weekday"`
	Saturday        PayBreakdown                `json:"saturday"`
	Sunday          PayBreakdown                `json:"sunday"`
	TotalHours      float64                     `json:"total_hours"`
	PercentageOfAll float64                     `json:"percentage_of_all"`
	GrossPay        float64                     `json:"gross_pay"`
	CashHours       float64                     `json:"cash_hours"`
	CashAmount      float64                     `json:"cash_amount"`
	HourCapWeekday  *float64                    `json:"hour_cap_weekday"`
	HourCapWeekend  *float64                    `json:"hour_cap_weekend"`
	PaySplitWeekday *float64                    `json:"pay_split_weekday"`
	PaySplitWeekend *float64                    `json:"pay_split_weekend"`
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
	UpsertPaySplit(ctx context.Context, branchID, userID int64, weekStart time.Time, weekdayHours, weekendHours float64) error
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
		d := dateutil.DayOf(sh.ClockInAt).Format(dateLayout)
		shiftsByUser[sh.UserID][d] = append(shiftsByUser[sh.UserID][d], ShiftEntryInfo{
			ID:         sh.ID,
			ClockInAt:  sh.ClockInAt,
			ClockOutAt: sh.ClockOutAt,
		})
	}

	splits, err := s.repo.ListPaySplits(ctx, branchID, dateutil.MondayOf(start), end)
	if err != nil {
		return nil, err
	}
	splitsByUser := make(map[int64]map[string]*PaySplit)
	for _, sp := range splits {
		if splitsByUser[sp.UserID] == nil {
			splitsByUser[sp.UserID] = make(map[string]*PaySplit)
		}
		splitsByUser[sp.UserID][sp.WeekStartDate.Format(dateLayout)] = sp
	}

	branchRate := s.branchRateResolver(ctx, branchID)
	displayWeekdayRate, displayWeekendRate, err := branchRate(start)
	if err != nil {
		return nil, err
	}

	type weekActual struct{ weekdayHours, satHours, sunHours float64 }

	rows := make([]EmployeeWeekRow, 0, len(employees))
	var weekTotalHours float64
	for _, employee := range employees {
		daily := make(map[string]float64, len(dates))
		dailyShifts := make(map[string][]ShiftEntryInfo, len(dates))
		weekActuals := make(map[string]*weekActual)
		var total float64
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
				weekKey := dateutil.MondayOf(date).Format(dateLayout)
				wa := weekActuals[weekKey]
				if wa == nil {
					wa = &weekActual{}
					weekActuals[weekKey] = wa
				}
				switch date.Weekday() {
				case time.Saturday:
					wa.satHours += hours
				case time.Sunday:
					wa.sunHours += hours
				default:
					wa.weekdayHours += hours
				}
			}
		}

		var grossPay, cashHours, cashAmount float64
		var weekdayPay, saturdayPay, sundayPay PayBreakdown
		for weekKey, wa := range weekActuals {
			weekStart, err := time.Parse(dateLayout, weekKey)
			if err != nil {
				return nil, err
			}
			branchWeekday, branchWeekend, err := branchRate(weekStart)
			if err != nil {
				return nil, err
			}
			weekdayRate := rateForDay(weekStart, employee.RateWeekday, employee.RateWeekend, branchWeekday, branchWeekend)
			weekendRate := rateForDay(weekStart.AddDate(0, 0, 5), employee.RateWeekday, employee.RateWeekend, branchWeekday, branchWeekend)

			split := splitsByUser[employee.ID][weekKey]
			if split == nil {
				weekdayPay.Hours += wa.weekdayHours
				weekdayPay.Total += wa.weekdayHours * weekdayRate
				saturdayPay.Hours += wa.satHours
				saturdayPay.Total += wa.satHours * weekendRate
				sundayPay.Hours += wa.sunHours
				sundayPay.Total += wa.sunHours * weekendRate
				grossPay += wa.weekdayHours*weekdayRate + (wa.satHours+wa.sunHours)*weekendRate
				continue
			}

			actualTotal := wa.weekdayHours + wa.satHours + wa.sunHours
			paidWeekday, paidWeekend := split.WeekdayHours, split.WeekendHours
			if paidWeekday < 0 {
				paidWeekday = 0
			}
			if paidWeekend < 0 {
				paidWeekend = 0
			}
			if excess := paidWeekday + paidWeekend - actualTotal; excess > 0 {
				if paidWeekend >= excess {
					paidWeekend -= excess
				} else {
					excess -= paidWeekend
					paidWeekend = 0
					paidWeekday -= excess
					if paidWeekday < 0 {
						paidWeekday = 0
					}
				}
			}
			cash := actualTotal - paidWeekday - paidWeekend

			weekdayPay.Hours += paidWeekday
			weekdayPay.Total += paidWeekday * weekdayRate
			// The manual split collapses Saturday/Sunday into one payable
			// "weekend" bucket, so it's carried on the Saturday breakdown —
			// Sunday stays untouched (0) and its payslip line hides itself.
			saturdayPay.Hours += paidWeekend
			saturdayPay.Total += paidWeekend * weekendRate
			grossPay += paidWeekday*weekdayRate + paidWeekend*weekendRate
			cashHours += cash
			// ponytail: cash hours are valued at the weekday rate as a
			// stand-in — revisit if cash should follow whichever day it
			// was actually worked on.
			cashAmount += cash * weekdayRate
		}

		if weekdayPay.Hours > 0 {
			weekdayPay.Rate = weekdayPay.Total / weekdayPay.Hours
		}
		if saturdayPay.Hours > 0 {
			saturdayPay.Rate = saturdayPay.Total / saturdayPay.Hours
		}
		if sundayPay.Hours > 0 {
			sundayPay.Rate = sundayPay.Total / sundayPay.Hours
		}

		var paySplitWeekday, paySplitWeekend *float64
		if sp := splitsByUser[employee.ID][dateutil.MondayOf(start).Format(dateLayout)]; sp != nil {
			paySplitWeekday, paySplitWeekend = &sp.WeekdayHours, &sp.WeekendHours
		}

		weekTotalHours += total
		rows = append(rows, EmployeeWeekRow{
			UserID:          employee.ID,
			Name:            employee.Name,
			EmployerName:    employee.EmployerName,
			EmployerABN:     employee.EmployerABN,
			DailyHours:      daily,
			DailyShifts:     dailyShifts,
			Weekday:         weekdayPay,
			Saturday:        saturdayPay,
			Sunday:          sundayPay,
			TotalHours:      total,
			GrossPay:        grossPay,
			CashHours:       cashHours,
			CashAmount:      cashAmount,
			HourCapWeekday:  employee.HourCapWeekday,
			HourCapWeekend:  employee.HourCapWeekend,
			PaySplitWeekday: paySplitWeekday,
			PaySplitWeekend: paySplitWeekend,
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

func (s *service) UpsertPaySplit(ctx context.Context, branchID, userID int64, weekStart time.Time, weekdayHours, weekendHours float64) error {
	return s.repo.UpsertPaySplit(ctx, branchID, userID, weekStart, weekdayHours, weekendHours)
}
