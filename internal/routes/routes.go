package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/handlers"
	"github.com/sigap2/sigap2/internal/middleware"
	"github.com/sigap2/sigap2/internal/models"
)

func Setup(app *fiber.App) {
	// Root redirects based on auth
	app.Get("/", func(c *fiber.Ctx) error {
		if c.Cookies("jwt") != "" {
			return c.Redirect("/dashboard")
		}
		return c.Redirect("/login")
	})

	// Public routes
	app.Get("/login", handlers.ShowLogin)
	app.Post("/login", handlers.ProcessLogin)
	app.Get("/register", handlers.ShowRegister)
	app.Post("/register", handlers.ProcessRegister)
	app.Post("/logout", handlers.Logout)
	app.Get("/logout", handlers.Logout) // Convenience for links

	// Public SOS routes
	app.Get("/sos", handlers.ShowSOS)
	app.Post("/api/sos", handlers.CreateReport)

	// Authenticated routes
	auth := app.Group("/", middleware.RequireAuth(), middleware.InjectUser())

	// Dashboard (Redirects Korban to /reports/my inside handler)
	auth.Get("/dashboard", handlers.Dashboard)
	auth.Post("/profile/change-password", handlers.ProcessChangeOwnPassword)

	// Reports (Korban)
	auth.Get("/reports/my", handlers.MyReports)
	
	// Reports view (All roles can view some part, handled in handler/template)
	auth.Get("/reports", middleware.RequireRole(models.RoleAdmin, models.RoleRelawan), handlers.ListReports)
	auth.Get("/reports/:id", handlers.ShowReportDetail)
	auth.Post("/reports/:id/complete", handlers.CompleteReport)
	
	// API routes (mostly for AJAX / Maps)
	api := auth.Group("/api")
	api.Get("/reports/markers", handlers.APIReportMarkers)
	api.Get("/notifications", middleware.RequireRole(models.RoleAdmin, models.RoleRelawan), handlers.APIGetNotifications)

	// Tracking API endpoints (authenticated, any role can update their own location)
	api.Post("/deliveries/:id/location", handlers.APIUpdateLocation)
	api.Get("/deliveries/active", handlers.APIGetActiveDeliveries)
	api.Get("/deliveries/:id/location", handlers.APIGetDeliveryLocation)

	// Logistics (Admin/Relawan)
	logistics := auth.Group("/logistics", middleware.RequireRole(models.RoleAdmin, models.RoleRelawan))
	logistics.Get("/", handlers.ListLogistics)
	logistics.Get("/create", handlers.ShowLogisticForm)
	logistics.Post("/", handlers.CreateLogistic)
	logistics.Post("/:id/delete", middleware.RequireRole(models.RoleAdmin), handlers.DeleteLogistic)

	// Distributions (Admin/Relawan)
	dist := auth.Group("/distributions", middleware.RequireRole(models.RoleAdmin, models.RoleRelawan))
	dist.Get("/", handlers.ListDistributions)
	dist.Get("/create", handlers.ShowDistributionForm)
	dist.Post("/", handlers.CreateDistribution)

	// Live Tracking (Admin only)
	tracking := auth.Group("/tracking", middleware.RequireRole(models.RoleAdmin))
	tracking.Get("/", handlers.ShowTrackingPage)
	tracking.Get("/:id", handlers.ShowTrackingDetail)

	// Delivery Confirmation (Relawan/Admin)
	deliver := auth.Group("/deliver", middleware.RequireRole(models.RoleRelawan, models.RoleAdmin))
	deliver.Get("/", handlers.ShowMyDeliveries)
	deliver.Get("/report/:report_id", handlers.ShowDeliverForm)
	deliver.Post("/report/:report_id", handlers.ProcessDeliver)

	// Verifications (Admin only)
	verify := auth.Group("/verifications", middleware.RequireRole(models.RoleAdmin))
	verify.Get("/", handlers.ShowVerifications)
	verify.Post("/:id", handlers.ProcessVerification)

	// Fitur Admin (Laporan) — Export & Hapus Semua
	reportsAdmin := auth.Group("/reports/admin", middleware.RequireRole(models.RoleAdmin))
	reportsAdmin.Get("/export", handlers.ExportReportsCSV)
	reportsAdmin.Post("/delete-all", handlers.DeleteAllReports)

	// Direktori Relawan (Hanya Admin)
	volunteers := auth.Group("/volunteers", middleware.RequireRole(models.RoleAdmin))
	volunteers.Get("/", handlers.ListRelawan)
	volunteers.Get("/create", handlers.ShowCreateRelawan)
	volunteers.Post("/create", handlers.CreateRelawan)
	volunteers.Post("/:id/reset-password", handlers.ProcessResetPassword)
	volunteers.Get("/:id/password", handlers.GetVolunteerPassword)
	volunteers.Post("/:id/delete", handlers.DeleteVolunteer)
}
