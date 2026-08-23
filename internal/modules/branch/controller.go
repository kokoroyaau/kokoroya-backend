package branch

import (
	"github.com/gin-gonic/gin"

	"kokoroya-backend/internal/middleware"
	"kokoroya-backend/internal/response"
	"kokoroya-backend/internal/schema"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (ctrl *Controller) ListMine(c *gin.Context) {
	var (
		branches []*Branch
		err      error
	)

	if c.GetString("role") == middleware.RoleOwner {
		branches, err = ctrl.service.List(c.Request.Context())
	} else {
		branches, err = ctrl.service.ListForUser(c.Request.Context(), c.GetInt64("userID"))
	}

	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, branches)
}

func (ctrl *Controller) List(c *gin.Context) {
	branches, err := ctrl.service.List(c.Request.Context())
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, branches)
}

func (ctrl *Controller) Create(c *gin.Context) {
	var req schema.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	b, err := ctrl.service.Create(c.Request.Context(), req.Name)
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 201, b)
}

func (ctrl *Controller) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	var req schema.UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	b, err := ctrl.service.Update(c.Request.Context(), id, req.Name, req.IsActive)
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, b)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	if err := ctrl.service.Delete(c.Request.Context(), id); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) Employees(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	employees, err := ctrl.service.ListEmployees(c.Request.Context(), id)
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, employees)
}
