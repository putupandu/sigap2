package database

import (
	"log"

	"github.com/sigap2/sigap2/internal/models"
)

func MigrateDB() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Disaster{},
		&models.Report{},
		&models.Logistic{},
		&models.Distribution{},
	)
	
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Memaksa kolom user_id menjadi nullable (GORM AutoMigrate tidak mengubah constraint NOT NULL)
	DB.Exec("ALTER TABLE reports MODIFY user_id bigint unsigned NULL")

	log.Println("Database migrated successfully")
}
