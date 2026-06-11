package services

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
)

// preValidateReport performs basic sanity checks BEFORE any ML/AI processing.
// This is Gate 0 — catches obviously invalid input instantly without burning API credits.
func preValidateReport(description, needs string) error {
	desc := strings.TrimSpace(description)

	// 1. Description WAJIB diisi
	if desc == "" {
		return fmt.Errorf("VALIDASI: Keterangan kondisi wajib diisi. Jelaskan situasi bencana yang Anda alami.")
	}

	// 2. Minimum 15 karakter
	if len([]rune(desc)) < 15 {
		return fmt.Errorf("VALIDASI: Keterangan terlalu pendek (minimal 15 karakter). Jelaskan situasi bencana secara detail.")
	}

	// 3. Minimum 3 kata
	words := strings.Fields(desc)
	if len(words) < 3 {
		return fmt.Errorf("VALIDASI: Keterangan terlalu singkat (minimal 3 kata). Jelaskan kondisi, lokasi, dan kebutuhan Anda.")
	}

	// 4. Block teks yang 100% angka (tidak ada huruf alfabet sama sekali)
	hasLetter := false
	for _, r := range desc {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return fmt.Errorf("VALIDASI: Keterangan tidak boleh hanya berisi angka atau simbol. Jelaskan situasi bencana dengan kata-kata.")
	}

	// 5. Block karakter berulang > 3x berturut-turut (e.g. "aaaa", "1111", "!!!!")
	repeatingRegex := regexp.MustCompile(`(.)\1{3,}`)
	cleaned := repeatingRegex.ReplaceAllString(desc, "")
	// Jika setelah dibersihkan panjangnya < 50% dari aslinya, itu mostly repetisi
	if len([]rune(cleaned)) < len([]rune(desc))/2 {
		return fmt.Errorf("VALIDASI: Keterangan tidak valid (karakter berulang). Jelaskan situasi bencana yang sebenarnya.")
	}

	// 6. Block teks yang hanya terdiri dari 1 kata unik yang diulang-ulang
	// e.g. "halo halo halo halo halo"
	if len(words) >= 3 {
		uniqueWords := make(map[string]bool)
		for _, w := range words {
			uniqueWords[strings.ToLower(w)] = true
		}
		if len(uniqueWords) == 1 {
			return fmt.Errorf("VALIDASI: Keterangan tidak valid (kata berulang). Jelaskan situasi bencana secara detail.")
		}
	}

	// 7. Needs juga harus diisi
	if strings.TrimSpace(needs) == "" {
		return fmt.Errorf("VALIDASI: Pilih minimal satu kebutuhan (Makanan, Air Bersih, Evakuasi, dll).")
	}

	return nil
}

func CreateReport(reporterName, needs string, lat, lng float64, desc string) error {
	// Step 0: Pre-validation (basic sanity checks)
	if err := preValidateReport(desc, needs); err != nil {
		return err
	}

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
