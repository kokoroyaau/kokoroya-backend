package schedule

import (
	"context"
	"time"

	"kokoroya-backend/internal/modules/branch"
)

const dateLayout = "2006-01-02"

type ShiftCell struct {
	StartTime *string `json:"start_time"`
	Code      *string `json:"code"`
}

type EmployeeWeekRow struct {
	UserID int64                `json:"user_id"`
	Name   string               `json:"name"`
	Shifts map[string]ShiftCell `json:"shifts"`
}

type SectionWeekRow struct {
	SectionID   int64             `json:"section_id"`
	SectionName string            `json:"section_name"`
	Employees   []EmployeeWeekRow `json:"employees"`
}

type WeeklyReport struct {
	WeekStartDate string           `json:"week_start_date"`
	WeekEndDate   string           `json:"week_end_date"`
	Sections      []SectionWeekRow `json:"sections"`
	Notes         string           `json:"notes"`
}

type Service interface {
	ListSections(ctx context.Context, branchID int64) ([]*Section, error)
	CreateSection(ctx context.Context, branchID int64, name string) (*Section, error)
	UpdateSection(ctx context.Context, id int64, fields SectionUpdateFields) (*Section, error)
	DeleteSection(ctx context.Context, id int64) error

	GetWeeklyReport(ctx context.Context, branchID int64, weekStart time.Time) (*WeeklyReport, error)
	UpsertShift(ctx context.Context, branchID, sectionID, userID int64, date time.Time, startTime, code *string) error
	UpsertNotes(ctx context.Context, branchID int64, weekStart time.Time, notes string) error
}

type service struct {
	repo       Repository
	branchRepo branch.Repository
}

func NewService(repo Repository, branchRepo branch.Repository) Service {
	return &service{repo: repo, branchRepo: branchRepo}
}

func (s *service) ListSections(ctx context.Context, branchID int64) ([]*Section, error) {
	return s.repo.ListSections(ctx, branchID)
}

func (s *service) CreateSection(ctx context.Context, branchID int64, name string) (*Section, error) {
	return s.repo.CreateSection(ctx, branchID, name)
}

func (s *service) UpdateSection(ctx context.Context, id int64, fields SectionUpdateFields) (*Section, error) {
	return s.repo.UpdateSection(ctx, id, fields)
}

func (s *service) DeleteSection(ctx context.Context, id int64) error {
	return s.repo.DeleteSection(ctx, id)
}

func (s *service) GetWeeklyReport(ctx context.Context, branchID int64, weekStart time.Time) (*WeeklyReport, error) {
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekDates := make([]string, 7)
	for i := range weekDates {
		weekDates[i] = weekStart.AddDate(0, 0, i).Format(dateLayout)
	}

	sections, err := s.repo.ListSections(ctx, branchID)
	if err != nil {
		return nil, err
	}
	employees, err := s.branchRepo.ListEmployees(ctx, branchID)
	if err != nil {
		return nil, err
	}
	shifts, err := s.repo.ListShifts(ctx, branchID, weekStart, weekEnd)
	if err != nil {
		return nil, err
	}

	cellBySection := make(map[int64]map[int64]map[string]ShiftCell)
	for _, sh := range shifts {
		if cellBySection[sh.SectionID] == nil {
			cellBySection[sh.SectionID] = make(map[int64]map[string]ShiftCell)
		}
		if cellBySection[sh.SectionID][sh.UserID] == nil {
			cellBySection[sh.SectionID][sh.UserID] = make(map[string]ShiftCell)
		}
		cellBySection[sh.SectionID][sh.UserID][sh.ShiftDate.Format(dateLayout)] = ShiftCell{
			StartTime: sh.StartTime,
			Code:      sh.Code,
		}
	}

	sectionRows := make([]SectionWeekRow, 0, len(sections))
	for _, sec := range sections {
		employeeRows := make([]EmployeeWeekRow, 0, len(employees))
		for _, employee := range employees {
			shiftsForUser := make(map[string]ShiftCell, 7)
			for _, d := range weekDates {
				shiftsForUser[d] = cellBySection[sec.ID][employee.ID][d]
			}
			employeeRows = append(employeeRows, EmployeeWeekRow{
				UserID: employee.ID,
				Name:   employee.Name,
				Shifts: shiftsForUser,
			})
		}
		sectionRows = append(sectionRows, SectionWeekRow{
			SectionID:   sec.ID,
			SectionName: sec.Name,
			Employees:   employeeRows,
		})
	}

	notes, err := s.repo.FindNotes(ctx, branchID, weekStart)
	if err != nil {
		return nil, err
	}

	return &WeeklyReport{
		WeekStartDate: weekStart.Format(dateLayout),
		WeekEndDate:   weekEnd.Format(dateLayout),
		Sections:      sectionRows,
		Notes:         notes,
	}, nil
}

func (s *service) UpsertShift(ctx context.Context, branchID, sectionID, userID int64, date time.Time, startTime, code *string) error {
	return s.repo.UpsertShift(ctx, branchID, sectionID, userID, date, startTime, code)
}

func (s *service) UpsertNotes(ctx context.Context, branchID int64, weekStart time.Time, notes string) error {
	return s.repo.UpsertNotes(ctx, branchID, weekStart, notes)
}
