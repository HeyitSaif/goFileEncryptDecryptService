package routes

import (
	Controllers "Iagon/app/controllers"

	"github.com/gofiber/fiber/v2"
)

// PublicRoutes func for describe group of public routes.
func PublicRoutes(a *fiber.App) {
	// Create routes group.
	route := a.Group("/api/v1")

	// upload files
	route.Post("/upload", Controllers.UploadFile)
	route.Get("/download", Controllers.Download)
}
