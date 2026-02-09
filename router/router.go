package router

import (
	"metachan/controllers"

	"github.com/gofiber/fiber/v2"
)

func Initialize(router *fiber.App) {
	// Health
	router.Get("/health", controllers.HealthStatus)

	// Anime routes
	animeRouter := router.Group("/anime")
	animeRouter.Get("/:id", controllers.GetAnime)

	// Anime routes
	// animeRouter := router.Group("/a")
	// animeRouter.Get("/genres", controllers.GetGenres)
	// animeRouter.Get("/genres/:id", controllers.GetAnimeByGenre)
	// animeRouter.Get("/:id", controllers.GetAnime)
	// animeRouter.Get("/:id/episodes", controllers.GetAnimeEpisodes)
	// animeRouter.Get("/:id/episodes/:episodeId", controllers.GetAnimeEpisode)
	// animeRouter.Get("/:id/characters", controllers.GetAnimeCharacters)

	// // Producer routes
	// producerRouter := router.Group("/producers")
	// producerRouter.Get("/", controllers.GetProducers)
	// producerRouter.Get("/:id", controllers.GetProducer)
	// producerRouter.Get("/:id/anime", controllers.GetAnimeByProducer)

	// // 404 Default
	// router.Use(func(c *fiber.Ctx) error {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Not Found",
	// 	})
	// })
}

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	switch code {
	case fiber.StatusBadRequest:
		return controllers.BadRequest(ctx, err)
	case fiber.StatusUnauthorized:
		return controllers.Unauthorized(ctx, err)
	case fiber.StatusForbidden:
		return controllers.Forbidden(ctx, err)
	case fiber.StatusNotFound:
		return controllers.NotFound(ctx, err)
	case fiber.StatusInternalServerError:
		return controllers.InternalServerError(ctx, err)
	default:
		return controllers.DefaultError(ctx, err)
	}
}
