package services

import (
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
)

func GetLogistics() ([]models.Logistic, error) {
	var logistics []models.Logistic
	err := database.DB.Find(&logistics).Error
	return logistics, err
}

func CreateLogistic(item models.Logistic) error {
	return database.DB.Create(&item).Error
}

func UpdateLogisticStock(id uint, quantityChange int) error {
	var logistic models.Logistic
	err := database.DB.First(&logistic, id).Error
	if err != nil {
		return err
	}

	logistic.Quantity += quantityChange
	return database.DB.Save(&logistic).Error
}

func DeleteLogistic(id uint) error {
	return database.DB.Delete(&models.Logistic{}, id).Error
}
