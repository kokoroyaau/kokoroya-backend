package labour

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, controller *Controller, authMW, requireBranch, requirePerm gin.HandlerFunc) {
	labour := rg.Group("/labour", authMW, requireBranch, requirePerm)
	labour.GET("/report", controller.GetReport)
	labour.PUT("/hour-entry", controller.UpsertHourEntry)
	labour.PUT("/rate", controller.UpsertWeeklyRate)
}

// RegisterSalaryRoutes exposes the same report read-only under /salary,
// gated by its own permission key instead of "labour".
func RegisterSalaryRoutes(rg *gin.RouterGroup, controller *Controller, authMW, requireBranch, requirePerm gin.HandlerFunc) {
	rg.GET("/salary/report", authMW, requireBranch, requirePerm, controller.GetReport)
}
