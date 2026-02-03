package tasks

import (
	"fmt"
	"metachan/database"
	"metachan/entities"
	"metachan/utils/api/jikan"
	"metachan/utils/logger"
)

// GenreSync synchronizes genre data from MAL via Jikan API
func GenreSync() error {
	logger.Log("Starting Genre Sync from MAL", logger.LogOptions{
		Level:  logger.Info,
		Prefix: "GenreSync",
	})

	// Create Jikan client
	client := jikan.NewJikanClient()

	// Wait for rate limit
	client.WaitForRateLimit()

	// Fetch genres from Jikan API
	genresResponse, err := client.GetAnimeGenres()
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to fetch genres from MAL: %v", err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "GenreSync",
		})
		return err
	}

	logger.Log(fmt.Sprintf("Fetched %d genres from MAL", len(genresResponse.Data)), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "GenreSync",
	})

	// Update or create genres in database
	for _, genre := range genresResponse.Data {
		// Create a genre entry with AnimeID = 0 to indicate it's a master genre
		genreEntity := entities.AnimeGenre{
			AnimeID: 0, // Master genre, not tied to specific anime
			GenreID: genre.MALID,
			Name:    genre.Name,
			URL:     genre.URL,
			Count:   genre.Count,
		}

		// Update or create
		var existing entities.AnimeGenre
		result := database.DB.Where("genre_id = ? AND anime_id = 0", genre.MALID).First(&existing)

		if result.Error == nil {
			// Update existing
			existing.Name = genre.Name
			existing.URL = genre.URL
			existing.Count = genre.Count
			if err := database.DB.Save(&existing).Error; err != nil {
				logger.Log(fmt.Sprintf("Failed to update genre %s: %v", genre.Name, err), logger.LogOptions{
					Level:  logger.Warn,
					Prefix: "GenreSync",
				})
			}
		} else {
			// Create new
			if err := database.DB.Create(&genreEntity).Error; err != nil {
				logger.Log(fmt.Sprintf("Failed to create genre %s: %v", genre.Name, err), logger.LogOptions{
					Level:  logger.Warn,
					Prefix: "GenreSync",
				})
			}
		}
	}

	logger.Log("Genre Sync completed successfully", logger.LogOptions{
		Level:  logger.Success,
		Prefix: "GenreSync",
	})

	return nil
}
