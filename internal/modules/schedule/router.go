package schedule

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, controller *Controller, authMW, requireBranch, requirePerm gin.HandlerFunc) {
	schedule := rg.Group("/schedule", authMW, requireBranch, requirePerm)
	schedule.GET("/report", controller.GetWeeklyReport)
	schedule.PUT("/shift", controller.UpsertShift)
	schedule.PUT("/notes", controller.UpsertNotes)
	schedule.GET("/sections", controller.ListSections)
	schedule.POST("/sections", controller.CreateSection)
	schedule.PUT("/sections/:id", controller.UpdateSection)
	schedule.DELETE("/sections/:id", controller.DeleteSection)
}
