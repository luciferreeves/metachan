package controllers

import (
	"errors"
	"metachan/enums"
	"metachan/repositories"
	"metachan/utils/meta"

	"github.com/gofiber/fiber/v2"
)

func GetAnime(c *fiber.Ctx) error {
	id := meta.Request(c).MustHave().Param("id")
	provider := meta.Request(c).Default("mal").Query("provider")

	switch provider {
	case "mal", "anilist":
	default:
		return BadRequest(c, errors.New("invalid provider"))
	}

	anime, err := repositories.GetAnime(enums.MappingType(provider), id)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(anime)
}

// -- Old Code Below --

// import (
// 	"fmt"
// 	"metachan/database"
// 	"metachan/entities"
// 	animeService "metachan/services/anime"
// 	"metachan/utils/logger"
// 	"metachan/utils/mappers"

// 	"github.com/gofiber/fiber/v2"
// )

// // animeServiceInstance is a singleton instance of the anime service
// var animeServiceInstance *animeService.Service

// // getAnimeService returns the anime service instance, creating it if necessary
// func getAnimeService() *animeService.Service {
// 	if animeServiceInstance == nil {
// 		animeServiceInstance = animeService.NewService()
// 	}
// 	return animeServiceInstance
// }

// // GetAnimeByMALID fetches anime details by MAL ID
// func GetAnime(c *fiber.Ctx) error {
// 	mapping, err := getAnimeMapping(c)
// 	if err != nil {
// 		return err
// 	}

// 	service := getAnimeService()
// 	anime, err := service.GetAnimeDetails(mapping)
// 	if err != nil {
// 		logger.Log("Failed to fetch anime details: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "AnimeAPI",
// 		})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch anime details",
// 		})
// 	}

// 	return c.JSON(anime)
// }

// // GetAnimeEpisodesByMALID fetches anime episodes by MAL ID
// func GetAnimeEpisodes(c *fiber.Ctx) error {
// 	mapping, err := getAnimeMapping(c)
// 	if err != nil {
// 		return err
// 	}

// 	service := getAnimeService()
// 	anime, err := service.GetAnimeDetails(mapping)
// 	if err != nil {
// 		logger.Log("Failed to fetch anime episodes: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "AnimeAPI",
// 		})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch anime episodes",
// 		})
// 	}

// 	// Return only the episodes data
// 	return c.JSON(anime.Episodes)
// }

// // GetAnimeEpisode fetches a single episode by anime ID and episode ID
// func GetAnimeEpisode(c *fiber.Ctx) error {
// 	mapping, err := getAnimeMapping(c)
// 	if err != nil {
// 		return err
// 	}

// 	episodeID := c.Params("episodeId")
// 	if episodeID == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Episode ID is required",
// 		})
// 	}

// 	service := getAnimeService()
// 	anime, err := service.GetAnimeDetails(mapping)
// 	if err != nil {
// 		logger.Log("Failed to fetch anime details: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "AnimeAPI",
// 		})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch anime details",
// 		})
// 	}

// 	// Find the episode with matching ID
// 	for i, episode := range anime.Episodes.Episodes {
// 		if episode.ID == episodeID {
// 			// Fetch streaming sources for this specific episode
// 			episodeNumber := i + 1
// 			streaming, err := service.GetEpisodeStreaming(anime.Titles.Romaji, episodeNumber, episode.ID, uint(anime.MALID))
// 			if err != nil {
// 				logger.Log("Failed to fetch streaming sources: "+err.Error(), logger.LogOptions{
// 					Level:  logger.Warn,
// 					Prefix: "AnimeAPI",
// 				})
// 				// Continue without streaming data
// 			} else {
// 				episode.Streaming = streaming
// 			}
// 			return c.JSON(episode)
// 		}
// 	}

// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 		"error": "Episode not found",
// 	})
// }

// func GetAnimeCharacters(c *fiber.Ctx) error {
// 	mapping, err := getAnimeMapping(c)
// 	if err != nil {
// 		return err
// 	}

// 	service := getAnimeService()
// 	anime, err := service.GetAnimeDetails(mapping)
// 	if err != nil {
// 		logger.Log("Failed to fetch anime characters: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "AnimeAPI",
// 		})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch anime characters",
// 		})
// 	}

// 	return c.JSON(anime.Characters)
// }

// func getAnimeMapping(c *fiber.Ctx) (*entities.AnimeMapping, error) {
// 	isAnilist := c.Query("provider") == "anilist"
// 	malID := c.Params("id")
// 	if malID == "" {
// 		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Query parameter MAL ID is required",
// 		})
// 	}

// 	var mapping *entities.AnimeMapping
// 	var err error
// 	if isAnilist {
// 		mapping, err = database.GetAnimeMappingViaAnilistID(mappers.ForceInt(malID))
// 	} else {
// 		mapping, err = database.GetAnimeMappingViaMALID(mappers.ForceInt(malID))
// 	}
// 	if err != nil || mapping.MAL == 0 {
// 		return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"error": "Anime mapping not found",
// 		})
// 	}

// 	return mapping, nil
// }

// // GetGenres retrieves all genres from the database
// func GetGenres(c *fiber.Ctx) error {
// 	genres, err := database.GetAllGenres()
// 	if err != nil {
// 		logger.Log("Failed to get genres: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "Controller",
// 		})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to retrieve genres",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"genres": genres,
// 	})
// }

// // GetAnimeByGenre retrieves paginated anime list for a specific genre
// func GetAnimeByGenre(c *fiber.Ctx) error {
// 	genreID := mappers.ForceInt(c.Params("id"))
// 	page := mappers.ForceInt(c.Query("page", "1"))
// 	limit := min(mappers.ForceInt(c.Query("limit", "12")), 100)
// 	if limit < 1 {
// 		limit = 25
// 	}
// 	if page < 1 {
// 		page = 1
// 	}

// 	logger.Log(fmt.Sprintf("Fetching anime for genre %d, page %d, limit %d", genreID, page, limit), logger.LogOptions{
// 		Level:  logger.Debug,
// 		Prefix: "GenreController",
// 	})

// 	service := getAnimeService()
// 	animeList, pagination, err := service.GetAnimeByGenre(genreID, page, limit)
// 	if err != nil {
// 		logger.Log("Failed to fetch anime by genre: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "GenreController",
// 		})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to retrieve anime for genre",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"pagination": fiber.Map{
// 			"current_page":      pagination.CurrentPage,
// 			"has_next_page":     pagination.HasNextPage,
// 			"last_visible_page": pagination.LastVisiblePage,
// 			"total":             pagination.Items.Total,
// 			"per_page":          pagination.Items.PerPage,
// 		},
// 		"data": animeList,
// 	})
// }

// // GetProducers retrieves all producers with counts
// func GetProducers(c *fiber.Ctx) error {
// 	page := mappers.ForceInt(c.Query("page", "1"))
// 	limit := min(mappers.ForceInt(c.Query("limit", "25")), 100)
// 	if limit < 1 {
// 		limit = 25
// 	}
// 	if page < 1 {
// 		page = 1
// 	}

// 	producers, pagination, err := database.GetAllProducers(page, limit)
// 	if err != nil {
// 		logger.Log("Failed to get producers: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "Controller",
// 		})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to retrieve producers",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"pagination": pagination,
// 		"producers":  producers,
// 	})
// }

// func GetProducer(c *fiber.Ctx) error {
// 	producerID := mappers.ForceInt(c.Params("id"))

// 	producer, err := database.GetProducerByID(producerID)
// 	if err != nil {
// 		logger.Log("Failed to get producer: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "Controller",
// 		})
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"error": "Producer not found",
// 		})
// 	}

// 	return c.JSON(producer)
// }

// // GetAnimeByProducer retrieves paginated anime list for a specific producer
// func GetAnimeByProducer(c *fiber.Ctx) error {
// 	producerID := mappers.ForceInt(c.Params("id"))
// 	page := mappers.ForceInt(c.Query("page", "1"))
// 	limit := min(mappers.ForceInt(c.Query("limit", "12")), 100)
// 	if limit < 1 {
// 		limit = 25
// 	}
// 	if page < 1 {
// 		page = 1
// 	}

// 	logger.Log(fmt.Sprintf("Fetching anime for producer %d, page %d, limit %d", producerID, page, limit), logger.LogOptions{
// 		Level:  logger.Debug,
// 		Prefix: "ProducerController",
// 	})

// 	service := getAnimeService()
// 	animeList, pagination, err := service.GetAnimeByProducer(producerID, page, limit)
// 	if err != nil {
// 		logger.Log("Failed to fetch anime by producer: "+err.Error(), logger.LogOptions{
// 			Level:  logger.Error,
// 			Prefix: "ProducerController",
// 		})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to retrieve anime for producer",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"pagination": fiber.Map{
// 			"current_page":      pagination.CurrentPage,
// 			"has_next_page":     pagination.HasNextPage,
// 			"last_visible_page": pagination.LastVisiblePage,
// 			"total":             pagination.Items.Total,
// 			"per_page":          pagination.Items.PerPage,
// 		},
// 		"data": animeList,
// 	})
// }
