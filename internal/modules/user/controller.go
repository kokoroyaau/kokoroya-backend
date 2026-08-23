package user

import (
	"errors"

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

func (ctrl *Controller) Login(c *gin.Context) {
	var req schema.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	token, expiresAt, role, err := ctrl.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Err(c, 401, err.Error())
			return
		}
		response.Err(c, 500, "internal server error")
		return
	}

	data := schema.LoginResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt.Unix(),
		Role:        role,
	}

	response.OK(c, 200, data)
}

func (ctrl *Controller) Logout(c *gin.Context) {
	jti := c.GetString("jti")
	if err := ctrl.service.Logout(c.Request.Context(), jti); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) CreateUser(c *gin.Context) {
	var req schema.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	var pin string
	if req.Pin != nil {
		pin = *req.Pin
	}
	u, err := ctrl.service.CreateUser(c.Request.Context(), req.Name, req.Email, req.Password, req.Role, req.Phone, req.TFN, pin, req.RateWeekday, req.RateWeekend, req.Permissions, req.BranchIDs)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			response.Err(c, 409, err.Error())
			return
		}
		response.DBErr(c, err)
		return
	}

	response.OK(c, 201, u)
}

func (ctrl *Controller) UpdateUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	var req schema.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	u, err := ctrl.service.UpdateUser(c.Request.Context(), id, UpdateFields{
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		TFN:         req.TFN,
		PIN:         req.Pin,
		Role:        req.Role,
		IsActive:    req.IsActive,
		RateWeekday: req.RateWeekday,
		RateWeekend: req.RateWeekend,
	})
	if err != nil {
		response.DBErr(c, err)
		return
	}
	response.OK(c, 200, u)
}

func (ctrl *Controller) DeleteUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	if err := ctrl.service.DeleteUser(c.Request.Context(), id); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) SetBranches(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	var req schema.SetBranchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	if err := ctrl.service.SetBranches(c.Request.Context(), id, req.BranchIDs); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) SetPermissions(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	var req schema.SetPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	if err := ctrl.service.SetPermissions(c.Request.Context(), id, req.Permissions); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) List(c *gin.Context) {
	users, err := ctrl.service.List(c.Request.Context())
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, users)
}

func (ctrl *Controller) Permissions(c *gin.Context) {
	response.OK(c, 200, schema.PermissionsResponse{Pages: middleware.Pages})
}

func (ctrl *Controller) Me(c *gin.Context) {
	u, err := ctrl.service.Me(c.Request.Context(), c.GetInt64("userID"))
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}

	var email string
	if u.Email != nil {
		email = *u.Email
	}
	data := schema.MeResponse{
		ID:          u.ID,
		Name:        u.Name,
		Email:       email,
		Role:        u.Role,
		Permissions: u.Permissions,
	}
	response.OK(c, 200, data)
}
