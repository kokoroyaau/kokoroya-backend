package clock

import (
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
