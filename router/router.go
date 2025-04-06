package router

import "github.com/gofiber/fiber/v2"

func Initialize(router *fiber.App) {

	// 404 Default
	router.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	})
}
