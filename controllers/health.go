package controllers

import (
	"metachan/database"
	"metachan/tasks"
	"metachan/types"
	"metachan/utils/stats"

	"github.com/gofiber/fiber/v2"
)

func HealthStatus(c *fiber.Ctx) error {
	// Check if the database is connected
	databaseStatus := database.GetConnectionStatus()

	// Get the memory stats
	memoryStats := stats.GetMemoryStats()

	// Get the task statuses
	taskStatuses := tasks.GlobalTaskManager.GetAllTaskStatuses()

	statusString := map[bool]string{
		true:  "healthy",
		false: "unhealthy",
	}[databaseStatus]
	// Create the health status response
	healthStatus := types.HealthStatus{
		Status:    statusString,
		Timestamp: stats.GetCurrentTimestamp(),
		Uptime:    stats.GetUptime(),
		Memory:    memoryStats,
		Database:  types.DatabaseStatus{Connected: databaseStatus, LastChecked: stats.GetCurrentTimestamp()},
		Tasks:     taskStatuses,
	}
	return c.JSON(healthStatus)
}
