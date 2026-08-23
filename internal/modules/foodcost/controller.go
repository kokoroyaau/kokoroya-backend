package foodcost

import (
	"strconv"
	"time"

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

func parseIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func (ctrl *Controller) GetReport(c *gin.Context) {
	start, end, ok := parseDateRange(c)
	if !ok {
		return
	}

	report, err := ctrl.service.GetReport(c.Request.Context(), c.GetInt64("branchID"), start, end)
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, report)
}

func parseDateRange(c *gin.Context) (start, end time.Time, ok bool) {
	start, err := time.Parse("2006-01-02", c.Query("start_date"))
	if err != nil {
		response.Err(c, 400, "start_date must be in YYYY-MM-DD format")
		return start, end, false
	}
	end, err = time.Parse("2006-01-02", c.Query("end_date"))
	if err != nil {
		response.Err(c, 400, "end_date must be in YYYY-MM-DD format")
		return start, end, false
	}
	if end.Before(start) {
		response.Err(c, 400, "end_date must not be before start_date")
		return start, end, false
	}
	return start, end, true
}

func (ctrl *Controller) UpsertPurchaseEntry(c *gin.Context) {
	var req schema.UpsertPurchaseEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	date, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		response.Err(c, 400, "purchase_date must be in YYYY-MM-DD format")
		return
	}

	if err := ctrl.service.UpsertPurchaseEntry(c.Request.Context(), c.GetInt64("branchID"), req.SupplierID, date, req.Amount); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) UpsertGrossSales(c *gin.Context) {
	var req schema.UpsertGrossSalesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	date, err := time.Parse("2006-01-02", req.SalesDate)
	if err != nil {
		response.Err(c, 400, "sales_date must be in YYYY-MM-DD format")
		return
	}

	if err := ctrl.service.UpsertGrossSales(c.Request.Context(), c.GetInt64("branchID"), date, req.Amount); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) UpsertNetSalesRate(c *gin.Context) {
	var req schema.UpsertNetSalesRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}
	weekStart, err := time.Parse("2006-01-02", req.WeekStartDate)
	if err != nil || weekStart.Weekday() != time.Monday {
		response.Err(c, 400, "week_start_date must be in YYYY-MM-DD format and a Monday")
		return
	}

	if err := ctrl.service.UpsertNetSalesRate(c.Request.Context(), c.GetInt64("branchID"), weekStart, req.Rate); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}

func (ctrl *Controller) ListSuppliers(c *gin.Context) {
	suppliers, err := ctrl.service.ListSuppliers(c.Request.Context(), c.GetInt64("branchID"))
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, suppliers)
}

func (ctrl *Controller) CreateSupplier(c *gin.Context) {
	var req schema.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	supplier, err := ctrl.service.CreateSupplier(c.Request.Context(), c.GetInt64("branchID"), req.Name)
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 201, supplier)
}

func (ctrl *Controller) UpdateSupplier(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	var req schema.UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, 400, err.Error())
		return
	}

	supplier, err := ctrl.service.UpdateSupplier(c.Request.Context(), id, SupplierUpdateFields{
		Name:      req.Name,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.OK(c, 200, supplier)
}

func (ctrl *Controller) DeleteSupplier(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.Err(c, 400, "invalid id")
		return
	}

	if err := ctrl.service.DeleteSupplier(c.Request.Context(), id); err != nil {
		response.Err(c, 500, "internal server error")
		return
	}
	response.NoContent(c)
}
