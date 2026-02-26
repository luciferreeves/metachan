package main

import (
	"fmt"
	"metachan/config"
	"metachan/database"
	"metachan/middleware"
	"metachan/router"
	"metachan/tasks"
	"metachan/utils/api/aniskip"
	"metachan/utils/mal"
	"metachan/utils/logger"
	"os"
	"os/signal"
	"syscall"

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

	middleware.Initialize(app)
	router.Initialize(app)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)); err != nil {
			logger.Fatalf("Main", "Failed to start the server on %s:%d: %v", config.Server.Host, config.Server.Port, err)
		}
	}()

	logger.Successf("Main", "Server started on %s:%d", config.Server.Host, config.Server.Port)

	<-quit
	logger.Infof("Main", "Shutting down gracefully...")

	if err := app.Shutdown(); err != nil {
		logger.Errorf("Main", "Error during server shutdown: %v", err)
	}

	tasks.GlobalTaskManager.StopAllTasks()
	mal.StopRateLimiters()
	aniskip.StopRateLimiters()

	if sqlDB, err := database.DB.DB(); err == nil {
		sqlDB.Close()
	}

	logger.Successf("Main", "Shutdown complete")
}
