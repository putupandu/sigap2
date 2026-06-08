package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
)

func Dashboard(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)

	if user.Role == models.RoleKorban {
		return c.Redirect("/reports/my")
	}

	var totalReports, pendingReports, completedReports int64
	database.DB.Model(&models.Report{}).Count(&totalReports)
	database.DB.Model(&models.Report{}).Where("status = ?", models.StatusPending).Count(&pendingReports)
	database.DB.Model(&models.Report{}).Where("status = ?", models.StatusCompleted).Count(&completedReports)

	var reports []models.Report
	database.DB.Preload("User").Order("created_at desc").Limit(5).Find(&reports)

	return c.Render("dashboard/admin", fiber.Map{
		"TotalReports":     totalReports,
		"PendingReports":   pendingReports,
		"CompletedReports": completedReports,
		"RecentReports":    reports,
	}, "layouts/base")
}
