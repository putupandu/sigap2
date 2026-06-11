package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/models"
	"github.com/sigap2/sigap2/internal/services"
	"github.com/sigap2/sigap2/internal/database"
)

func ShowSOS(c *fiber.Ctx) error {
	return c.Render("reports/create", fiber.Map{}, "layouts/base")
}

func CreateReport(c *fiber.Ctx) error {
	lat, _ := strconv.ParseFloat(c.FormValue("latitude"), 64)
	lng, _ := strconv.ParseFloat(c.FormValue("longitude"), 64)
	desc := c.FormValue("description")
	reporterName := c.FormValue("reporter_name")
	needs := c.FormValue("needs")

	err := services.CreateReport(reporterName, needs, lat, lng, desc)
	if err != nil {
		errMsg := err.Error()
		// Jika ini adalah error validasi (pre-validation, ML, atau AI)
		if len(errMsg) > 9 && (errMsg[:9] == "VALIDASI:" || errMsg[:15] == "Laporan ditolak") {
			if c.Accepts("json") == "json" || c.Get("X-Requested-With") == "XMLHttpRequest" {
				return c.Status(400).JSON(fiber.Map{"error": errMsg})
			}
			return c.Redirect("/sos?error=invalid_report")
		}
		// Error internal server lainnya
		if c.Accepts("json") == "json" || c.Get("X-Requested-With") == "XMLHttpRequest" {
			return c.Status(500).JSON(fiber.Map{"error": "Terjadi kesalahan pada server. Silakan coba lagi."})
		}
		return c.Status(500).SendString("Terjadi kesalahan pada server: " + errMsg)
	}

	// If API call, return JSON
	if c.Accepts("json") == "json" || c.Get("X-Requested-With") == "XMLHttpRequest" {
		return c.JSON(fiber.Map{"message": "SOS Sent!"})
	}

	return c.Redirect("/sos?success=true")
}

func ListReports(c *fiber.Ctx) error {
	reports, _ := services.GetReports()
	return c.Render("reports/index", fiber.Map{
		"Reports": reports,
	}, "layouts/base")
}

func MyReports(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	var reports []models.Report
	database.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&reports)
	
	return c.Render("reports/index", fiber.Map{
		"Reports": reports,
		"IsMyReports": true,
	}, "layouts/base")
}

func ShowReportDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	var report models.Report
	if err := database.DB.Preload("User").First(&report, id).Error; err != nil {
		return c.Status(404).SendString("Report not found")
	}

	var distributions []models.Distribution
	database.DB.Where("report_id = ?", report.ID).Preload("Logistic").Preload("Volunteer").Find(&distributions)

	isEvacuationOnly := false
	if report.UrgencyLevel == models.UrgencyHigh || report.UrgencyLevel == models.UrgencyCritical {
		if report.ExtractedData == "[]" || report.ExtractedData == "" || report.ExtractedData == "null" {
			isEvacuationOnly = true
		}
	}

	return c.Render("reports/detail", fiber.Map{
		"Report":           report,
		"Distributions":    distributions,
		"IsEvacuationOnly": isEvacuationOnly,
	}, "layouts/base")
}

func CompleteReport(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	user := c.Locals("user").(models.User)

	var report models.Report
	if err := database.DB.First(&report, id).Error; err != nil {
		return c.Status(404).SendString("Report not found")
	}

	// Pastikan hanya pembuat laporan (korban) atau relawan/admin yang bisa menutup
	if user.Role == models.RoleKorban && (report.UserID == nil || *report.UserID != user.ID) {
		return c.Status(403).SendString("Forbidden")
	}

	err := services.UpdateReportStatus(uint(id), models.StatusCompleted)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Redirect("/reports/" + c.Params("id"))
}

// ExportReportsCSV mengunduh seluruh laporan dalam bentuk CSV
func ExportReportsCSV(c *fiber.Ctx) error {
	var reports []models.Report
	database.DB.Preload("User").Order("created_at desc").Find(&reports)

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=laporan_sigap.csv")

	// Tulis header CSV
	csvStr := "ID Laporan,Waktu,Nama Pelapor,No WhatsApp,Kebutuhan,Keterangan,Status\n"

	for _, r := range reports {
		reporterName := r.ReporterName
		if reporterName == "" {
			if r.User.ID != 0 {
				reporterName = r.User.Name
			} else {
				reporterName = "Anonim"
			}
		}

		phone := ""
		if r.User.ID != 0 {
			phone = r.User.Phone
		}

		// Escape commas and quotes for CSV
		needs := `"` + r.Needs + `"`
		desc := `"` + r.Description + `"`

		csvStr += strconv.Itoa(int(r.ID)) + "," +
			r.CreatedAt.Format("2006-01-02 15:04:05") + "," +
			`"` + reporterName + `",` +
			`"` + phone + `",` +
			needs + "," +
			desc + "," +
			r.Status + "\n"
	}

	return c.SendString(csvStr)
}

// DeleteAllReports menghapus semua laporan dan distribusinya (Admin Only)
func DeleteAllReports(c *fiber.Ctx) error {
	// Matikan foreign key checks sementara agar TRUNCATE bisa berjalan
	database.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

	// TRUNCATE otomatis reset AUTO_INCREMENT ke 1 (lebih handal dari DELETE + ALTER TABLE)
	database.DB.Exec("TRUNCATE TABLE distributions")
	database.DB.Exec("TRUNCATE TABLE reports")

	// Aktifkan kembali foreign key checks
	database.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

	return c.Redirect("/reports?success=deleted")
}

