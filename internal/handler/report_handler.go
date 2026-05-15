package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"support-ticket.com/internal/service"
)

type ReportHandler struct {
	reportSvc service.ReportService
}

func NewReportHandler(reportSvc service.ReportService) *ReportHandler {
	return &ReportHandler{reportSvc: reportSvc}
}

// GetDaily godoc
// @Summary Get daily report
// @Description Get daily ticket SLA report by date. If date is not provided, today will be used.
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date query string false "Report date in YYYY-MM-DD format" example(2026-05-05)
// @Success 200 {object} map[string]interface{} "Get daily report successfully"
// @Failure 400 {object} map[string]interface{} "Invalid date format"
// @Failure 404 {object} map[string]interface{} "Report not found"
// @Router /reports/daily [get]
func (h *ReportHandler) GetDaily(c *gin.Context) {
	// Parse date từ query param, default là hôm nay
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid date format, expected YYYY-MM-DD",
		})
		return
	}

	report, err := h.reportSvc.GetReport(date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}
