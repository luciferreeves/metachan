package main

import (
	"fmt"
	"metachan/config"
	"metachan/middleware"
	"metachan/router"
	"metachan/tasks"
	"metachan/utils/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func main() {
	tasks.GlobalTaskManager.StartAllTasks()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "GET, HEAD, PUT, PATCH, POST, DELETE, OPTIONS",
		AllowHeaders:  "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-API-Key, X-CSRF-Token",
		ExposeHeaders: "Content-Length, Content-Type, Content-Disposition, X-Pagination, X-Total-Count",
		MaxAge:        86400,
	}))
	app.Use(helmet.New())

	router.Initialize(app)
	middleware.Initialize(app)

	// Start the server
	if err := app.Listen(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)); err != nil {
		logger.Fatalf("Main", "Failed to the start the server on %s:%d: %v", config.Server.Host, config.Server.Port, err)
	}

	logger.Successf("Main", "Server started on %s:%d", config.Server.Host, config.Server.Port)
}
