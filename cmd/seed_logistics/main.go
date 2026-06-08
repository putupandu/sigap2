package main

import (
	"log"

	"github.com/sigap2/sigap2/internal/config"
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
)

func main() {
	config.LoadConfig()
	database.ConnectDB()

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

	for _, item := range logistics {
		var count int64
		database.DB.Model(&models.Logistic{}).Where("item_name = ?", item.ItemName).Count(&count)
		if count == 0 {
			database.DB.Create(&item)
			log.Printf("Added logistic: %s", item.ItemName)
		}
	}
	log.Println("Seeding logistics done!")
}
