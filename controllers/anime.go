package controllers

import (
	"metachan/database"
	"metachan/entities"
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
func GetAnime(c *fiber.Ctx) error {
	mapping, err := getAnimeMapping(c)
	if err != nil {
		return err
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

// GetAnimeEpisodesByMALID fetches anime episodes by MAL ID
func GetAnimeEpisodes(c *fiber.Ctx) error {
	mapping, err := getAnimeMapping(c)
	if err != nil {
		return err
	}

	service := getAnimeService()
	anime, err := service.GetAnimeDetails(mapping)
	if err != nil {
		logger.Log("Failed to fetch anime episodes: "+err.Error(), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AnimeAPI",
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch anime episodes",
		})
	}

	// Return only the episodes data
	return c.JSON(anime.Episodes)
}

func GetAnimeCharacters(c *fiber.Ctx) error {
	mapping, err := getAnimeMapping(c)
	if err != nil {
		return err
	}

	service := getAnimeService()
	anime, err := service.GetAnimeDetails(mapping)
	if err != nil {
		logger.Log("Failed to fetch anime characters: "+err.Error(), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AnimeAPI",
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch anime characters",
		})
	}

	return c.JSON(anime.Characters)
}

func getAnimeMapping(c *fiber.Ctx) (*entities.AnimeMapping, error) {
	isAnilist := c.Query("provider") == "anilist"
	malID := c.Params("id")
	if malID == "" {
		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter MAL ID is required",
		})
	}

	var mapping *entities.AnimeMapping
	var err error
	if isAnilist {
		mapping, err = database.GetAnimeMappingViaAnilistID(mappers.ForceInt(malID))
	} else {
		mapping, err = database.GetAnimeMappingViaMALID(mappers.ForceInt(malID))
	}
	if err != nil {
		return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Anime mapping not found",
		})
	}
	return mapping, nil
}
