package controllers

import (
	"metachan/database"
	animeService "metachan/services/anime"
	"metachan/utils/logger"
	"metachan/utils/mappers"

	"github.com/gofiber/fiber/v2"
)

// animeServiceInstance is a singleton instance of the anime service
var animeServiceInstance *animeService.Service

// getAnimeService returns the anime service instance, creating it if necessary
func getAnimeService() *animeService.Service {
	if animeServiceInstance == nil {
		animeServiceInstance = animeService.NewService()
	}
	return animeServiceInstance
}

// GetAnimeByMALID fetches anime details by MAL ID
func GetAnimeByMALID(c *fiber.Ctx) error {
	malID := c.Params("mal_id")
	if malID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter MAL ID is required",
		})
	}

	mapping, err := database.GetAnimeMappingViaMALID(mappers.ForceInt(malID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Anime not found",
		})
	}

	service := getAnimeService()
	anime, err := service.GetAnimeDetails(mapping)
	if err != nil {
		logger.Log("Failed to fetch anime details: "+err.Error(), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AnimeAPI",
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch anime details",
		})
	}

	return c.JSON(anime)
}
