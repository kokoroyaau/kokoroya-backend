package clock

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, controller *Controller, authMW, requireBranch, requireLabour gin.HandlerFunc) {
	clock := rg.Group("/clock", authMW, requireBranch)
	clock.POST("/punch", controller.Punch)
	clock.PUT("/entries/:id", requireLabour, controller.UpdateEntry)
}
