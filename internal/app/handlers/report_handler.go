package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"reports-system/internal/domain/entities"
	"reports-system/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type ReportHandler struct {
	service *usecase.ReportService
}

func NewReportHandler(service *usecase.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) GetReport(c fiber.Ctx) error {
	reportID := c.Params("report_id")
	format := c.Query("format", "json")

	println("Report ID:", reportID)

	// Extrair parâmetros da query string
	params := make(map[string]interface{})
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := string(key)
		if keyStr != "format" {
			valueStr := string(value)
			// Tentar converter para número se possível
			if num, err := strconv.ParseFloat(valueStr, 64); err == nil {
				params[keyStr] = num
			} else {
				params[keyStr] = valueStr
			}
		}
	})

	report, err := h.service.GetReport(reportID, params, format)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	switch strings.ToLower(format) {
	case "csv":
		return h.renderCSV(c, report)
	case "xlsx":
		return h.renderXLSX(c, report)
	default:
		return c.JSON(report)
	}
}

func (h *ReportHandler) PostReport(c fiber.Ctx) error {
	reportID := c.Params("report_id")

	var requestBody struct {
		Params map[string]interface{} `json:"params"`
		Format string                 `json:"format"`
	}

	if err := c.Bind().JSON(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON body",
		})
	}

	if requestBody.Format == "" {
		requestBody.Format = "json"
	}

	report, err := h.service.GetReport(reportID, requestBody.Params, requestBody.Format)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(report)
}

func (h *ReportHandler) GetAvailableReports(c fiber.Ctx) error {
	reports := h.service.GetAvailableReports()
	return c.JSON(fiber.Map{
		"reports": reports,
	})
}

func (h *ReportHandler) renderCSV(c fiber.Ctx, report *entities.ReportResponse) error {
	// Implementação básica CSV
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", report.Metadata.Report))

	// Converter dados para CSV (implementação simplificada)
	return c.SendString("CSV export not fully implemented yet")
}

func (h *ReportHandler) renderXLSX(c fiber.Ctx, report *entities.ReportResponse) error {
	// Implementação básica XLSX
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", report.Metadata.Report))

	// Converter dados para XLSX (implementação simplificada)
	return c.SendString("XLSX export not fully implemented yet")
}
