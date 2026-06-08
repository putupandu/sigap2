package services

import (
	"errors"
	"time"

	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
	"gorm.io/gorm"
)

func CreateDistribution(reportID uint, logisticIDs []uint, quantities []int, volunteerID uint) error {
	if len(logisticIDs) != len(quantities) || len(logisticIDs) == 0 {
		return errors.New("invalid logistic data")
	}

	// Start transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(logisticIDs); i++ {
			logID := logisticIDs[i]
			qty := quantities[i]

			if qty <= 0 {
				continue // Skip invalid quantities
			}

			var logistic models.Logistic
			if err := tx.First(&logistic, logID).Error; err != nil {
				return err
			}

			if logistic.Quantity < qty {
				return errors.New("insufficient stock for " + logistic.ItemName)
			}

			// Deduct stock
			logistic.Quantity -= qty
			if err := tx.Save(&logistic).Error; err != nil {
				return err
			}

			// Create distribution record with volunteer tracking
			dist := models.Distribution{
				ReportID:     reportID,
				LogisticID:   logID,
				QuantitySent: qty,
				VolunteerID:  &volunteerID,
				Status:       models.DistStatusDelivering,
			}
			if err := tx.Create(&dist).Error; err != nil {
				return err
			}

			// Notify
			NotifyDistribution(reportID, logistic.ItemName, qty)
		}

		// Update report status if pending
		var report models.Report
		if err := tx.First(&report, reportID).Error; err == nil {
			if report.Status == models.StatusPending {
				report.Status = models.StatusProcess
				tx.Save(&report)
				NotifyReportStatusChange(report.ID, models.StatusProcess)
			}
		}

		return nil
	})
}

func GetDistributions() ([]models.Distribution, error) {
	var dists []models.Distribution
	err := database.DB.Preload("Report").Preload("Report.User").Preload("Logistic").Preload("Volunteer").Preload("VerifiedBy").Order("timestamp desc").Find(&dists).Error
	return dists, err
}

func GetDistributionByID(id uint) (models.Distribution, error) {
	var dist models.Distribution
	err := database.DB.Preload("Report").Preload("Report.User").Preload("Logistic").Preload("Volunteer").Preload("VerifiedBy").First(&dist, id).Error
	return dist, err
}

// UpdateVolunteerLocation updates the GPS position of a volunteer for a specific delivery
func UpdateVolunteerLocation(distributionID uint, lat, lng float64) error {
	return database.DB.Model(&models.Distribution{}).Where("id = ? AND status = ?", distributionID, models.DistStatusDelivering).
		Updates(map[string]interface{}{
			"volunteer_lat": lat,
			"volunteer_lng": lng,
		}).Error
}

// MarkDelivered marks a distribution as delivered with a proof photo
func MarkDelivered(distributionID uint, photoPath string) error {
	return database.DB.Model(&models.Distribution{}).Where("id = ?", distributionID).
		Updates(map[string]interface{}{
			"status":           models.DistStatusDelivered,
			"proof_photo_path": photoPath,
		}).Error
}

// VerifyDistribution allows admin to verify or reject a delivered distribution
func VerifyDistribution(distributionID, adminID uint, approved bool, notes string) error {
	now := time.Now()
	status := models.DistStatusVerified
	if !approved {
		status = models.DistStatusRejected
	}

	err := database.DB.Model(&models.Distribution{}).Where("id = ?", distributionID).
		Updates(map[string]interface{}{
			"status":         status,
			"verified_by_id": adminID,
			"verified_at":    &now,
			"admin_notes":    notes,
		}).Error

	if err != nil {
		return err
	}

	// If verified, check if all distributions for this report are verified → mark report as completed
	if approved {
		var dist models.Distribution
		database.DB.First(&dist, distributionID)

		var unverifiedCount int64
		database.DB.Model(&models.Distribution{}).
			Where("report_id = ? AND status NOT IN (?)", dist.ReportID, []string{models.DistStatusVerified}).
			Count(&unverifiedCount)

		if unverifiedCount == 0 {
			UpdateReportStatus(dist.ReportID, models.StatusCompleted)
		}
	}

	return nil
}

// GetActiveDeliveries returns all distributions that are currently being delivered
func GetActiveDeliveries() ([]models.Distribution, error) {
	var dists []models.Distribution
	err := database.DB.Preload("Report").Preload("Report.User").Preload("Logistic").Preload("Volunteer").
		Where("status = ?", models.DistStatusDelivering).
		Order("timestamp desc").Find(&dists).Error
	return dists, err
}

// GetPendingVerifications returns all distributions awaiting admin verification
func GetPendingVerifications() ([]models.Distribution, error) {
	var dists []models.Distribution
	err := database.DB.Preload("Report").Preload("Report.User").Preload("Logistic").Preload("Volunteer").
		Where("status = ?", models.DistStatusDelivered).
		Order("timestamp desc").Find(&dists).Error
	return dists, err
}

// GetVolunteerActiveDeliveries returns active deliveries for a specific volunteer
func GetVolunteerActiveDeliveries(volunteerID uint) ([]models.Distribution, error) {
	var dists []models.Distribution
	err := database.DB.Preload("Report").Preload("Report.User").Preload("Logistic").Preload("Volunteer").
		Where("volunteer_id = ? AND status = ?", volunteerID, models.DistStatusDelivering).
		Order("timestamp desc").Find(&dists).Error
	return dists, err
}

// GetDistributionsByReportAndVolunteer returns active distributions for a specific report assigned to a specific volunteer
func GetDistributionsByReportAndVolunteer(reportID, volunteerID uint) ([]models.Distribution, error) {
	var dists []models.Distribution
	err := database.DB.Preload("Report").Preload("Report.User").Preload("Logistic").Preload("Volunteer").
		Where("report_id = ? AND volunteer_id = ? AND status = ?", reportID, volunteerID, models.DistStatusDelivering).
		Order("timestamp desc").Find(&dists).Error
	return dists, err
}

// MarkDeliveriesDeliveredByReport marks ALL active distributions for a report+volunteer as delivered with a single proof photo
func MarkDeliveriesDeliveredByReport(reportID, volunteerID uint, photoPath string) error {
	return database.DB.Model(&models.Distribution{}).
		Where("report_id = ? AND volunteer_id = ? AND status = ?", reportID, volunteerID, models.DistStatusDelivering).
		Updates(map[string]interface{}{
			"status":           models.DistStatusDelivered,
			"proof_photo_path": photoPath,
		}).Error
}
