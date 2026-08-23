package labour

import (
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

func (ctrl *Controller) GetReport(c *gin.Context) {
	start, err := time.Parse("2006-01-02", c.Query("start_date"))
	if err != nil {
		response.Err(c, 400, "start_date must be in YYYY-MM-DD format")
		return
	}
	end, err := time.Parse("2006-01-02", c.Query("end_date"))
	if err != nil {
		response.Err(c, 400, "end_date must be in YYYY-MM-DD format")
		return
	}
	if end.Before(start) {
		response.Err(c, 400, "end_date must not be before start_date")
		return
	}

	report, err := ctrl.service.GetReport(c.Request.Context(), c.GetInt64("branchID"), start, end)
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, report)
}

func (ctrl *Controller) UpsertHourEntry(c *gin.Context) {
	var req schema.UpsertHourEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	date, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		response.Err(c, 400, "entry_date must be in YYYY-MM-DD format")
		return
	}

	if err := ctrl.service.UpsertHourEntry(c.Request.Context(), c.GetInt64("branchID"), req.UserID, date, req.TotalHours); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) UpsertWeeklyRate(c *gin.Context) {
	var req schema.UpsertLabourRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	weekStart, err := time.Parse("2006-01-02", req.WeekStartDate)
	if err != nil || weekStart.Weekday() != time.Monday {
		response.Err(c, 400, "week_start_date must be in YYYY-MM-DD format and a Monday")
		return
	}

	if err := ctrl.service.UpsertWeeklyRate(c.Request.Context(), c.GetInt64("branchID"), weekStart, req.WeekdayRate, req.WeekendRate); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}
