package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
	"github.com/sigap2/sigap2/internal/services"
)

type GroupedDist struct {
	ReportID   uint
	Report     models.Report
	Timestamp  string
	Status     string
	TrackingID uint
	Items      []models.Distribution
}

func ListDistributions(c *fiber.Ctx) error {
	dists, _ := services.GetDistributions()

	groupedMap := make(map[uint]*GroupedDist)
	var orderedKeys []uint

	for _, d := range dists {
		if _, exists := groupedMap[d.ReportID]; !exists {
			groupedMap[d.ReportID] = &GroupedDist{
				ReportID:   d.ReportID,
				Report:     d.Report,
				Timestamp:  d.Timestamp.Format("02 Jan 2006 15:04"),
				Status:     d.Status,
				TrackingID: d.ID,
			}
			orderedKeys = append(orderedKeys, d.ReportID)
		}
		groupedMap[d.ReportID].Items = append(groupedMap[d.ReportID].Items, d)
	}

	var groupedDists []GroupedDist
	for _, k := range orderedKeys {
		groupedDists = append(groupedDists, *groupedMap[k])
	}

	return c.Render("distributions/index", fiber.Map{
		"GroupedDistributions": groupedDists,
	}, "layouts/base")
}

func ShowDistributionForm(c *fiber.Ctx) error {
	selectedReportID := c.Query("report_id")
	evacuationOnly := c.Query("evacuation_only") == "true"
	user := c.Locals("user").(models.User)

	var reports []models.Report
	database.DB.Preload("User").Where("status != ?", models.StatusCompleted).Find(&reports)

	logistics, _ := services.GetLogistics()
	
	var evacItemID uint
	if evacuationOnly {
		for _, l := range logistics {
			if l.ItemName == "Tim Evakuasi Darurat" {
				evacItemID = l.ID
				break
			}
		}
	}

	var volunteers []models.User
	if user.Role == models.RoleAdmin {
		database.DB.Where("role = ?", models.RoleRelawan).Find(&volunteers)
	}

	return c.Render("distributions/form", fiber.Map{
		"Reports":          reports,
		"Logistics":        logistics,
		"SelectedReportID": selectedReportID,
		"EvacuationOnly":   evacuationOnly,
		"EvacItemID":       evacItemID,
		"Volunteers":       volunteers,
		"User":             user,
	}, "layouts/base")
}

func CreateDistribution(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)

	// 1. Ambil Report ID
	reportIDStr := c.FormValue("report_id")
	if reportIDStr == "" {
		reportIDStr = c.Query("report_id") // Fallback
	}

	reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
	if err != nil || reportID == 0 {
		return c.Status(400).SendString("Error: Pilih laporan SOS yang valid")
	}

	// Tentukan Volunteer ID
	assignedVolunteerID := user.ID // Default to self
	if user.Role == models.RoleAdmin {
		formVolIDStr := c.FormValue("volunteer_id")
		if formVolIDStr != "" {
			if parsedID, err := strconv.ParseUint(formVolIDStr, 10, 32); err == nil && parsedID > 0 {
				assignedVolunteerID = uint(parsedID)
			}
		}
	}

	// 2. Ambil Array Logistik menggunakan PeekMulti (paling handal di Fiber untuk form arrays)
	logisticIDsRaw := c.Request().PostArgs().PeekMulti("logistic_id[]")
	quantitiesRaw := c.Request().PostArgs().PeekMulti("quantity[]")

	// Jika tidak pakai [], coba tanpa [] (hanya jaga-jaga)
	if len(logisticIDsRaw) == 0 {
		logisticIDsRaw = c.Request().PostArgs().PeekMulti("logistic_id")
		quantitiesRaw = c.Request().PostArgs().PeekMulti("quantity")
	}

	var finalLogisticIDs []uint
	var finalQuantities []int

	for i := 0; i < len(logisticIDsRaw); i++ {
		id, _ := strconv.ParseUint(string(logisticIDsRaw[i]), 10, 32)
		
		qty := 0
		if i < len(quantitiesRaw) {
			qty, _ = strconv.Atoi(string(quantitiesRaw[i]))
		}
		
		if id > 0 && qty > 0 {
			finalLogisticIDs = append(finalLogisticIDs, uint(id))
			finalQuantities = append(finalQuantities, qty)
		}
	}

	if len(finalLogisticIDs) == 0 {
		return c.Status(400).SendString("Error: Pilih setidaknya satu barang dengan jumlah yang valid")
	}

	// 3. Panggil Service (with assigned volunteer ID)
	err = services.CreateDistribution(uint(reportID), finalLogisticIDs, finalQuantities, assignedVolunteerID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Redirect("/distributions")
}
