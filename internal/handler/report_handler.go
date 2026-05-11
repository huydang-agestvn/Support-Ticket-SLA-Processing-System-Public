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
// GET /reports/daily?date=2026-05-05
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
