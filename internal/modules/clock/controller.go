package clock

import (
	"strconv"

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

func (ctrl *Controller) Punch(c *gin.Context) {
	var req schema.PunchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	result, err := ctrl.service.Punch(c.Request.Context(), req.Pin, c.GetInt64("branchID"))
	if err != nil {
		response.Err(c, 404, "invalid pin")
		return
	}

	response.OK(c, 200, schema.PunchResponse{
		Action: result.Action,
		Name:   result.Name,
		At:     result.At,
	})
}

func (ctrl *Controller) UpdateEntry(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	var req schema.UpdateClockEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	if req.ClockOutAt != nil && req.ClockOutAt.Before(req.ClockInAt) {
		response.Err(c, 400, "clock_out_at must not be before clock_in_at")
		return
	}

	entry, err := ctrl.service.UpdateEntry(c.Request.Context(), id, c.GetInt64("branchID"), req.ClockInAt, req.ClockOutAt)
	if err == ErrNotFound {
		response.Err(c, 404, "time entry not found")
		return
	}
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, entry)
}
