package controllers

import (
	"metachan/database"
	"metachan/types"
	"metachan/utils/anime"
	"metachan/utils/logger"
	"metachan/utils/mappers"

	"github.com/gofiber/fiber/v2"
)

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

	anime, err := anime.GetAnimeDetails(mapping)
	if err != nil {
		logger.Log("Failed to fetch anime details: "+err.Error(), types.LogOptions{
			Level:  types.Error,
			Prefix: "AnimeAPI",
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch anime details",
		})
	}

	return c.JSON(anime)
}
