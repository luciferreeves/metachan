package router

import (
	"metachan/controllers"

	"github.com/gofiber/fiber/v2"
)

func Initialize(router *fiber.App) {
	// Health
	router.Get("/health", controllers.HealthStatus)

	// Anime routes
	animeRouter := router.Group("/a")
	animeRouter.Get("/:id", controllers.GetAnime)
	animeRouter.Get("/:id/episodes", controllers.GetAnimeEpisodes)
	animeRouter.Get("/:id/characters", controllers.GetAnimeCharacters)

	// 404 Default
	router.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	})
}
