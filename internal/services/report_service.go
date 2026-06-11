package services

import (
	"fmt"
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
)

func CreateReport(reporterName, needs string, lat, lng float64, desc string) error {
	// Step 1: Local ML Classification
	combinedText := desc + " " + needs
	if CheckIfIrrelevantML(combinedText) {
		return fmt.Errorf("Laporan ditolak: Sistem kami mendeteksi teks ini bukan format laporan keadaan darurat.")
	}

	urgency, extractedData, errNLP := AnalyzeReport(desc, needs)
	if errNLP != nil {
		fmt.Println("Warning: NLP analysis failed:", errNLP)
		urgency = models.UrgencyMedium
		extractedData = "[]"
	}

	if urgency == "irrelevant" {
		return fmt.Errorf("Laporan ditolak: Keterangan tidak berhubungan dengan bencana atau keadaan darurat.")
	}

	report := models.Report{
		ReporterName:  reporterName,
		Needs:         needs,
		Latitude:      lat,
		Longitude:     lng,
		Description:   desc,
		Status:        models.StatusPending,
		UrgencyLevel:  urgency,
		ExtractedData: extractedData,
	}

	err := database.DB.Create(&report).Error
	if err == nil {
		// Notify relawan asynchronously
		NotifyNewReport(report.ID)
	}
	return err
}

func GetReports() ([]models.Report, error) {
	var reports []models.Report
	err := database.DB.Preload("User").Order("created_at desc").Find(&reports).Error
	return reports, err
}

func GetPendingReports() ([]models.Report, error) {
	var reports []models.Report
	err := database.DB.Preload("User").
		Where("status = ?", models.StatusPending).
		Order("created_at desc").
		Limit(20).
		Find(&reports).Error
	return reports, err
}

func UpdateReportStatus(reportID uint, status string) error {
	err := database.DB.Model(&models.Report{}).Where("id = ?", reportID).Update("status", status).Error
	if err == nil {
		// Notify korban asynchronously
		NotifyReportStatusChange(reportID, status)
	}
	return err
}
