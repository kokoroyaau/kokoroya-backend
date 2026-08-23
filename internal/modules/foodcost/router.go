package foodcost

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, controller *Controller, authMW, requireBranch, requirePerm gin.HandlerFunc) {
	foodCost := rg.Group("/food-cost", authMW, requireBranch, requirePerm)
	foodCost.GET("/report", controller.GetReport)
	foodCost.PUT("/purchase-entry", controller.UpsertPurchaseEntry)
	foodCost.PUT("/gross-sales", controller.UpsertGrossSales)
	foodCost.PUT("/net-sales-rate", controller.UpsertNetSalesRate)
	foodCost.GET("/suppliers", controller.ListSuppliers)
	foodCost.POST("/suppliers", controller.CreateSupplier)
	foodCost.PUT("/suppliers/:id", controller.UpdateSupplier)
	foodCost.DELETE("/suppliers/:id", controller.DeleteSupplier)
}
