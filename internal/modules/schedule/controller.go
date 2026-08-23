package schedule

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"kokoroya-backend/internal/response"
	"kokoroya-backend/internal/schema"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func parseIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func (ctrl *Controller) ListSections(c *gin.Context) {
	sections, err := ctrl.service.ListSections(c.Request.Context(), c.GetInt64("branchID"))
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, sections)
}

func (ctrl *Controller) CreateSection(c *gin.Context) {
	var req schema.CreateScheduleSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	section, err := ctrl.service.CreateSection(c.Request.Context(), c.GetInt64("branchID"), req.Name)
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 201, section)
}

func (ctrl *Controller) UpdateSection(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	var req schema.UpdateScheduleSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	section, err := ctrl.service.UpdateSection(c.Request.Context(), id, SectionUpdateFields{
		Name:      req.Name,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, section)
}

func (ctrl *Controller) DeleteSection(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	if err := ctrl.service.DeleteSection(c.Request.Context(), id); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) GetWeeklyReport(c *gin.Context) {
	weekStart, err := time.Parse("2006-01-02", c.Query("week_start_date"))
	if err != nil {
		response.Err(c, 400, "week_start_date must be in YYYY-MM-DD format")
		return
	}
	if weekStart.Weekday() != time.Monday {
		response.Err(c, 400, "week_start_date must be a Monday")
		return
	}

	report, err := ctrl.service.GetWeeklyReport(c.Request.Context(), c.GetInt64("branchID"), weekStart)
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, report)
}

func (ctrl *Controller) UpsertShift(c *gin.Context) {
	var req schema.UpsertShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	date, err := time.Parse("2006-01-02", req.ShiftDate)
	if err != nil {
		response.Err(c, 400, "shift_date must be in YYYY-MM-DD format")
		return
	}

	if err := ctrl.service.UpsertShift(c.Request.Context(), c.GetInt64("branchID"), req.SectionID, req.UserID, date, req.StartTime, req.Code); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) UpsertNotes(c *gin.Context) {
	var req schema.UpsertScheduleNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	weekStart, err := time.Parse("2006-01-02", req.WeekStartDate)
	if err != nil || weekStart.Weekday() != time.Monday {
		response.Err(c, 400, "week_start_date must be in YYYY-MM-DD format and a Monday")
		return
	}

	if err := ctrl.service.UpsertNotes(c.Request.Context(), c.GetInt64("branchID"), weekStart, req.Notes); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}
