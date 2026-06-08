package database

import (
	"log"

	"github.com/sigap2/sigap2/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedData() {
	var count int64
	DB.Model(&models.User{}).Count(&count)

	if count == 0 {
		log.Println("Seeding database...")

		// Admin
		adminHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		admin := models.User{
			Name:         "Super Admin",
			Email:        "admin@sigap.local",
			PasswordHash: string(adminHash),
			Role:         models.RoleAdmin,
			Phone:        "081234567890",
		}
		DB.Create(&admin)

		// Relawan
		relawanHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		relawan := models.User{
			Name:         "Tim Relawan 1",
			Email:        "relawan@sigap.local",
			PasswordHash: string(relawanHash),
			Role:         models.RoleRelawan,
			Phone:        "081298765432",
		}
		DB.Create(&relawan)

		// Korban (User)
		korbanHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		korban := models.User{
			Name:         "Budi Santoso",
			Email:        "budi@example.com",
			PasswordHash: string(korbanHash),
			Role:         models.RoleKorban,
			Phone:        "085512341234",
		}
		DB.Create(&korban)

		// Logistics
		logistics := []models.Logistic{
			{ItemName: "Beras Premium", Quantity: 500, Unit: "kg"},
			{ItemName: "Air Mineral Botol 600ml", Quantity: 200, Unit: "dus"},
			{ItemName: "Air Mineral Galon", Quantity: 100, Unit: "galon"},
			{ItemName: "Mie Instan Goreng", Quantity: 300, Unit: "dus"},
			{ItemName: "Mie Instan Kuah", Quantity: 300, Unit: "dus"},
			{ItemName: "Sarden Kaleng", Quantity: 250, Unit: "kaleng"},
			{ItemName: "Biskuit / Roti Kering", Quantity: 150, Unit: "dus"},
			{ItemName: "Susu Bayi Formula", Quantity: 100, Unit: "kotak"},
			{ItemName: "Susu UHT", Quantity: 150, Unit: "dus"},
			{ItemName: "Selimut Tebal", Quantity: 400, Unit: "pcs"},
			{ItemName: "Pakaian Layak Pakai (Dewasa)", Quantity: 500, Unit: "set"},
			{ItemName: "Pakaian Layak Pakai (Anak)", Quantity: 300, Unit: "set"},
			{ItemName: "Popok Bayi (Pampers)", Quantity: 200, Unit: "pack"},
			{ItemName: "Pembalut Wanita", Quantity: 200, Unit: "pack"},
			{ItemName: "Tenda Darurat / Terpal", Quantity: 50, Unit: "lembar"},
			{ItemName: "Matras / Karpet", Quantity: 100, Unit: "lembar"},
			{ItemName: "Obat P3K & Vitamin", Quantity: 100, Unit: "box"},
			{ItemName: "Minyak Kayu Putih / Tolak Angin", Quantity: 200, Unit: "botol"},
			{ItemName: "Sabun & Sikat Gigi", Quantity: 300, Unit: "paket"},
			{ItemName: "Senter & Baterai", Quantity: 150, Unit: "set"},
			{ItemName: "Masker Medis", Quantity: 100, Unit: "box"},
		}
		DB.Create(&logistics)

		log.Println("Seeding completed!")
	} else {
		log.Println("Database already seeded, skipping.")
	}

	// Always ensure the special "Tim Evakuasi Darurat" item exists for the pure-rescue workflow
	var evacItemCount int64
	DB.Model(&models.Logistic{}).Where("item_name = ?", "Tim Evakuasi Darurat").Count(&evacItemCount)
	if evacItemCount == 0 {
		DB.Create(&models.Logistic{
			ItemName: "Tim Evakuasi Darurat",
			Quantity: 999999,
			Unit:     "Tim",
		})
		log.Println("Added special item 'Tim Evakuasi Darurat'")
	}
}
