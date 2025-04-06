package main

import (
	"fmt"
	"metachan/config"
	"metachan/database"
	"metachan/middleware"
	"metachan/router"
	"metachan/tasks"
	"metachan/types"
	"metachan/utils/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func main() {
	database.AutoMigrate()

	tasks.GlobalTaskManager.StartAllTasks()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "GET, HEAD, PUT, PATCH, POST, DELETE, OPTIONS",
		AllowHeaders:  "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-API-Key, X-CSRF-Token",
		ExposeHeaders: "Content-Length, Content-Type, Content-Disposition, X-Pagination, X-Total-Count",
		MaxAge:        86400,
	}))
	app.Use(helmet.New())
	app.Use(middleware.HTTPLogger())

	// Initialize the router
	router.Initialize(app)

	// Start the server
	if err := app.Listen(fmt.Sprintf(":%d", config.Config.Port)); err != nil {
		logger.Log(fmt.Sprintf("Failed to the start the server on port %d: %v", config.Config.Port, err), types.LogOptions{
			Prefix: "Main",
			Level:  types.Error,
			Fatal:  true,
		})
	}

	logger.Log(fmt.Sprintf("Server started on port %d", config.Config.Port), types.LogOptions{
		Prefix: "Main",
		Level:  types.Success,
	})
}
