package controllers

import (
	"strconv"
	"time"

	"github.com/Joko206/UAS_PWEB1/database"
	"github.com/Joko206/UAS_PWEB1/models"
	"github.com/gofiber/fiber/v2"
)

// GetAuditLogs retrieves audit logs with pagination, search, and filters (admin only)
func GetAuditLogs(c *fiber.Ctx) error {
	// Parse query parameters
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")
	username := c.Query("username")
	action := c.Query("action")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")
	ip := c.Query("ip")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query
	query := database.DB.Model(&models.AuditLog{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if action != "" {
		query = query.Where("action = ?", action)
	}

	if ip != "" {
		query = query.Where("ip_address = ?", ip)
	}

	if dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			return sendResponse(c, fiber.StatusBadRequest, false, "Invalid date_from format. Use YYYY-MM-DD", nil)
		}
		query = query.Where("created_at >= ?", dateFrom)
	}

	if dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			return sendResponse(c, fiber.StatusBadRequest, false, "Invalid date_to format. Use YYYY-MM-DD", nil)
		}
		// Set to end of day
		dateTo = dateTo.Add(24*time.Hour - time.Second)
		query = query.Where("created_at <= ?", dateTo)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return handleError(c, err, "Failed to count audit logs")
	}

	// Get paginated results, ordered by created_at desc
	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return handleError(c, err, "Failed to retrieve audit logs")
	}

	// Calculate pagination info
	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := fiber.Map{
		"logs":        logs,
		"pagination": fiber.Map{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	return sendResponse(c, fiber.StatusOK, true, "Audit logs retrieved successfully", response)
}