package branch

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"kokoroya-backend/internal/middleware"
)

func parseIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func RegisterRoutes(rg *gin.RouterGroup, controller *Controller, authMW, requireEmployee gin.HandlerFunc) {
	rg.GET("/me/branches", authMW, controller.ListMine)

	branchesRead := rg.Group("/branches", authMW, requireEmployee)
	branchesRead.GET("", controller.List)
	branchesRead.GET("/:id/employees", controller.Employees)

	branchesManage := rg.Group("/branches", authMW, middleware.RequireRole(middleware.RoleOwner))
	branchesManage.POST("", controller.Create)
	branchesManage.PATCH("/:id", controller.Update)
	branchesManage.DELETE("/:id", controller.Delete)
}
