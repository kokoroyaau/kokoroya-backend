package user

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func RegisterRoutes(rg *gin.RouterGroup, controller *Controller, authMW, requireEmployee gin.HandlerFunc) {
	auth := rg.Group("/auth")
	auth.POST("/login", controller.Login)
	auth.POST("/logout", authMW, controller.Logout)

	rg.GET("/me", authMW, controller.Me)
	rg.GET("/permissions", authMW, controller.Permissions)

	users := rg.Group("/users", authMW, requireEmployee)
	users.GET("", controller.List)
	users.POST("", controller.CreateUser)
	users.PATCH("/:id", controller.UpdateUser)
	users.DELETE("/:id", controller.DeleteUser)
	users.PATCH("/:id/permissions", controller.SetPermissions)
	users.PATCH("/:id/branches", controller.SetBranches)
}
