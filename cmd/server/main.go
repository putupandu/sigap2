package main

import (
	"log"
	"os/exec"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	
	"github.com/sigap2/sigap2/internal/config"
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/routes"
)

func main() {
	// Init Config
	config.LoadConfig()

	// Init DB
	database.ConnectDB()
	database.MigrateDB()
	database.SeedData()

	// Initialize standard Go html template engine
	engine := html.New("./web/templates", ".html")
	
	// Define custom template functions if needed
	engine.AddFunc("dict", func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, nil // error
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, nil // error
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	})

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "SIGAP2 - Disaster Alert System",
		Views:   engine,
	})

	// Static files
	app.Static("/static", "./web/static")

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Routes
	routes.Setup(app)

	// Start server
	port := config.AppConfig.AppPort
	log.Printf("Starting server on port %s...", port)
	
	// Buka browser secara otomatis khusus untuk mempermudah (ala Laragon)
	go func() {
		time.Sleep(1 * time.Second)
		exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:"+port).Start()
	}()
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
