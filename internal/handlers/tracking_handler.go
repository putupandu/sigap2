package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/models"
	"github.com/sigap2/sigap2/internal/services"
)

// ============================================================
// TRACKING PAGES (Admin)
// ============================================================

// ShowTrackingPage renders the live tracking map page for admin
func ShowTrackingPage(c *fiber.Ctx) error {
	deliveries, _ := services.GetActiveDeliveries()
	return c.Render("tracking/index", fiber.Map{
		"Deliveries": deliveries,
	}, "layouts/base")
}

// ShowTrackingDetail renders the tracking detail for a specific delivery
func ShowTrackingDetail(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	dist, err := services.GetDistributionByID(uint(id))
	if err != nil {
		return c.Status(404).SendString("Distribution not found")
	}

	return c.Render("tracking/detail", fiber.Map{
		"Distribution": dist,
	}, "layouts/base")
}

// ============================================================
// TRACKING API ENDPOINTS
// ============================================================

// APIUpdateLocation updates the volunteer's GPS position (called by relawan's browser)
func APIUpdateLocation(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var body struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	err = services.UpdateVolunteerLocation(uint(id), body.Lat, body.Lng)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// APIGetActiveDeliveries returns all active deliveries as JSON (for admin tracking map)
func APIGetActiveDeliveries(c *fiber.Ctx) error {
	deliveries, err := services.GetActiveDeliveries()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var result []fiber.Map
	for _, d := range deliveries {
		volunteerName := "Unknown"
		if d.Volunteer.ID != 0 {
			volunteerName = d.Volunteer.Name
		}
		recipientName := "Anonim"
		if d.Report.ReporterName != "" {
			recipientName = d.Report.ReporterName
		} else if d.Report.User.ID != 0 {
			recipientName = d.Report.User.Name
		}

		result = append(result, fiber.Map{
			"id":              d.ID,
			"volunteer_name":  volunteerName,
			"volunteer_phone": d.Volunteer.Phone,
			"volunteer_lat":   d.VolunteerLat,
			"volunteer_lng":   d.VolunteerLng,
			"dest_lat":        d.Report.Latitude,
			"dest_lng":        d.Report.Longitude,
			"recipient_name":  recipientName,
			"item_name":       d.Logistic.ItemName,
			"quantity_sent":   d.QuantitySent,
			"unit":            d.Logistic.Unit,
			"report_id":       d.ReportID,
		})
	}

	if result == nil {
		result = []fiber.Map{}
	}

	return c.JSON(result)
}

// APIGetDeliveryLocation returns the current location for a specific delivery
func APIGetDeliveryLocation(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	dist, err := services.GetDistributionByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	volunteerName := "Unknown"
	if dist.Volunteer.ID != 0 {
		volunteerName = dist.Volunteer.Name
	}

	return c.JSON(fiber.Map{
		"id":              dist.ID,
		"status":          dist.Status,
		"volunteer_name":  volunteerName,
		"volunteer_lat":   dist.VolunteerLat,
		"volunteer_lng":   dist.VolunteerLng,
		"dest_lat":        dist.Report.Latitude,
		"dest_lng":        dist.Report.Longitude,
	})
}

// ============================================================
// DELIVERY CONFIRMATION (Relawan)
// ============================================================

// ShowDeliverForm renders the form for a volunteer to mark delivery + upload photo for all items to a victim
func ShowDeliverForm(c *fiber.Ctx) error {
	reportID, err := strconv.ParseUint(c.Params("report_id"), 10, 32)
	if err != nil {
		return c.Status(400).SendString("Invalid Report ID")
	}
	user := c.Locals("user").(models.User)

	dists, err := services.GetDistributionsByReportAndVolunteer(uint(reportID), user.ID)
	if err != nil || len(dists) == 0 {
		return c.Status(404).SendString("Active distributions not found for this report")
	}

	return c.Render("deliver/form", fiber.Map{
		"ReportID": reportID,
		"Report": dists[0].Report,
		"Distributions": dists,
	}, "layouts/base")
}

// ProcessDeliver processes the delivery confirmation with photo upload for a report group
func ProcessDeliver(c *fiber.Ctx) error {
	reportID, err := strconv.ParseUint(c.Params("report_id"), 10, 32)
	if err != nil {
		return c.Status(400).SendString("Invalid Report ID")
	}
	user := c.Locals("user").(models.User)

	// Handle file upload
	file, err := c.FormFile("proof_photo")
	if err != nil {
		return c.Status(400).SendString("Error: Foto bukti harus dilampirkan")
	}

	// Create uploads directory if not exists
	uploadsDir := "./web/static/uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return c.Status(500).SendString("Error creating upload directory")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("proof_report_%d_%d%s", reportID, time.Now().UnixMilli(), ext)
	savePath := filepath.Join(uploadsDir, filename)

	// Save file
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).SendString("Error saving file: " + err.Error())
	}

	// Update distribution status
	photoURL := "/static/uploads/" + filename
	err = services.MarkDeliveriesDeliveredByReport(uint(reportID), user.ID, photoURL)
	if err != nil {
		return c.Status(500).SendString("Error: " + err.Error())
	}

	return c.Redirect("/distributions")
}

// ============================================================
// ACTIVE DELIVERIES PAGE (Relawan)
// ============================================================

// DeliveryGroup groups multiple logistic items delivered to the same victim
type DeliveryGroup struct {
	ReportID      uint
	Report        models.Report
	Distributions []models.Distribution
}

// ShowMyDeliveries shows a relawan's active deliveries with GPS tracking
func ShowMyDeliveries(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	deliveries, _ := services.GetVolunteerActiveDeliveries(user.ID)
	
	// Group deliveries by ReportID
	groupMap := make(map[uint]*DeliveryGroup)
	var groups []DeliveryGroup
	
	for _, d := range deliveries {
		if group, exists := groupMap[d.ReportID]; exists {
			group.Distributions = append(group.Distributions, d)
		} else {
			newGroup := &DeliveryGroup{
				ReportID:      d.ReportID,
				Report:        d.Report,
				Distributions: []models.Distribution{d},
			}
			groupMap[d.ReportID] = newGroup
		}
	}
	
	// Preserve order (descending by timestamp of the first distribution)
	for _, d := range deliveries {
		if group, exists := groupMap[d.ReportID]; exists {
			groups = append(groups, *group)
			delete(groupMap, d.ReportID)
		}
	}

	return c.Render("deliver/my_deliveries", fiber.Map{
		"DeliveryGroups": groups,
	}, "layouts/base")
}

// ============================================================
// VERIFICATIONS (Admin)
// ============================================================

// ShowVerifications renders the pending verifications page for admin
func ShowVerifications(c *fiber.Ctx) error {
	pending, _ := services.GetPendingVerifications()
	return c.Render("verifications/index", fiber.Map{
		"PendingVerifications": pending,
	}, "layouts/base")
}

// ProcessVerification processes admin's verification decision
func ProcessVerification(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	action := c.FormValue("action") // "approve" or "reject"
	notes := c.FormValue("notes")

	approved := action == "approve"

	err = services.VerifyDistribution(uint(id), user.ID, approved, notes)
	if err != nil {
		return c.Status(500).SendString("Error: " + err.Error())
	}

	return c.Redirect("/verifications")
}
