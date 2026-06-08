package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/services"
)

func APIReportMarkers(c *fiber.Ctx) error {
	reports, err := services.GetReports()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// We only want to send limited data to the map for performance
	type Marker struct {
		ID     uint    `json:"id"`
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
		Status string  `json:"status"`
		Name   string  `json:"name"`
		Needs  string  `json:"needs"`
	}

	var markers []Marker
	for _, r := range reports {
		name := "Anonim"
		if r.ReporterName != "" {
			name = r.ReporterName
		} else if r.User.ID != 0 {
			name = r.User.Name
		}

		markers = append(markers, Marker{
			ID:     r.ID,
			Lat:    r.Latitude,
			Lng:    r.Longitude,
			Status: r.Status,
			Name:   name,
			Needs:  r.Needs,
		})
	}

	return c.JSON(markers)
}

// APIGetNotifications returns pending report notifications for admin/relawan
func APIGetNotifications(c *fiber.Ctx) error {
	reports, err := services.GetPendingReports()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	type NotifItem struct {
		ID           uint      `json:"id"`
		ReporterName string    `json:"reporter_name"`
		Needs        string    `json:"needs"`
		Status       string    `json:"status"`
		UrgencyLevel string    `json:"urgency_level"`
		CreatedAt    time.Time `json:"created_at"`
	}

	var items []NotifItem
	for _, r := range reports {
		name := "Anonim"
		if r.ReporterName != "" {
			name = r.ReporterName
		} else if r.User.ID != 0 {
			name = r.User.Name
		}
		items = append(items, NotifItem{
			ID:           r.ID,
			ReporterName: name,
			Needs:        r.Needs,
			Status:       r.Status,
			UrgencyLevel: r.UrgencyLevel,
			CreatedAt:    r.CreatedAt,
		})
	}

	if items == nil {
		items = []NotifItem{}
	}

	return c.JSON(fiber.Map{
		"count": len(items),
		"items": items,
	})
}
