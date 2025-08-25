package tasks

import (
	"fmt"
	"metachan/config"
	"metachan/database"
	"metachan/entities"
	"metachan/services/anime"
	"metachan/types"
	"metachan/utils/logger"
	"sync"
	"time"
)

// Constants for anime update task
const (
	// UpdaterSource identifies the source of the update request
	UpdaterSource = "updater"

	// MaxConcurrentUpdates limits the number of concurrent anime updates
	MaxConcurrentUpdates = 5

	// MaxConcurrentSQLiteUpdates limits concurrent updates for SQLite to prevent locks
	MaxConcurrentSQLiteUpdates = 1

	// UpdateInterval defines how often an anime should be updated even without specific triggers
	UpdateInterval = 6 * time.Hour
)

// animeUpdateJob represents a single anime update job
type animeUpdateJob struct {
	series entities.Anime
	reason string
}

// AnimeUpdate checks for airing anime that need to be updated
func AnimeUpdate() error {
	logger.Log("Starting Anime Update Task", logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AnimeUpdate",
	})

	// Find all currently airing anime
	var airingSeries []entities.Anime
	result := database.DB.
		Where("airing = ?", true).
		Preload("NextAiringEpisode").
		Preload("AiringSchedule").
		Find(&airingSeries)

	if result.Error != nil {
		logger.Log(fmt.Sprintf("Failed to fetch airing anime: %v", result.Error), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AnimeUpdate",
		})
		return result.Error
	}

	logger.Log(fmt.Sprintf("Found %d airing anime series", len(airingSeries)), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AnimeUpdate",
	})

	// Get current timestamp
	currentTime := time.Now().Unix()

	// Log the current time for debugging
	logger.Log(fmt.Sprintf("Current timestamp: %d (%s)",
		currentTime, time.Unix(currentTime, 0).Format(time.RFC3339)), logger.LogOptions{
		Level:  logger.Debug,
		Prefix: "AnimeUpdate",
	})

	// Create a channel for jobs
	jobs := make(chan animeUpdateJob, len(airingSeries))

	// Determine max concurrency based on database type
	maxWorkers := MaxConcurrentUpdates
	if config.Config.DatabaseDriver == types.SQLite {
		maxWorkers = MaxConcurrentSQLiteUpdates
		logger.Log(fmt.Sprintf("Using reduced concurrency (%d workers) for SQLite database",
			maxWorkers), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeUpdate",
		})
	}

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Create workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			animeService := anime.NewService()

			logger.Log(fmt.Sprintf("Started worker #%d", workerID), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AnimeUpdate",
			})

			// Process jobs from the channel
			for job := range jobs {
				updateAnime(animeService, job.series, job.reason)
			}
		}(i)
	}

	// Queue updates for anime that need it
	jobsQueued := 0
	for _, series := range airingSeries {
		// Check if we need to update this anime
		needsUpdate := false
		reason := ""

		// Log details about this particular anime for debugging
		logger.Log(fmt.Sprintf("Checking anime: %s (ID: %d)",
			series.TitleRomaji, series.MALID), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeUpdate",
		})

		// If there's no next airing episode data, we should update
		if series.NextAiringEpisode == nil || series.NextAiringEpisode.AiringAt == 0 {
			needsUpdate = true
			reason = "missing next episode data"
		} else if int64(series.NextAiringEpisode.AiringAt) <= currentTime {
			// If the next episode should have aired already, update to get fresh data
			needsUpdate = true
			reason = "next episode already aired"
		}

		// Check if the anime was last updated more than the update interval ago
		if !needsUpdate && !series.LastUpdated.IsZero() && time.Since(series.LastUpdated) > UpdateInterval {
			needsUpdate = true
			reason = fmt.Sprintf("regular update (last updated %s ago)",
				time.Since(series.LastUpdated).Round(time.Second))
		}

		// Log update decision
		if !needsUpdate {
			logger.Log(fmt.Sprintf("Skipping update for %s (ID: %d) - no update needed. Next airing at: %d",
				series.TitleRomaji, series.MALID, series.NextAiringEpisode.AiringAt), logger.LogOptions{
				Level:  logger.Debug,
				Prefix: "AnimeUpdate",
			})
			continue
		}

		// Add the job to the queue
		logger.Log(fmt.Sprintf("Queueing update for %s (ID: %d) - Reason: %s",
			series.TitleRomaji, series.MALID, reason), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeUpdate",
		})

		jobs <- animeUpdateJob{
			series: series,
			reason: reason,
		}
		jobsQueued++
	}

	// Close the job channel to signal workers that no more jobs are coming
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()

	logger.Log(fmt.Sprintf("Anime Update Task Completed - Processed %d anime", jobsQueued), logger.LogOptions{
		Level:  logger.Success,
		Prefix: "AnimeUpdate",
	})

	return nil
}

// updateAnime updates a single anime series
func updateAnime(animeService *anime.Service, series entities.Anime, reason string) {
	title := series.TitleRomaji
	if series.TitleEnglish != "" {
		title = series.TitleEnglish
	}

	logger.Log(fmt.Sprintf("Updating anime: %s (MAL ID: %d) - %s", title, series.MALID, reason), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AnimeUpdate",
	})

	// Get anime mapping for the service call
	mapping, err := database.GetAnimeMappingViaMALID(series.MALID)
	if err != nil {
		logger.Log(fmt.Sprintf("Error getting anime mapping for %s (MAL ID: %d): %v", title, series.MALID, err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AnimeUpdate",
		})
		return
	}

	// Get updated anime data from API
	updatedAnime, err := animeService.GetAnimeDetailsWithSource(mapping, "mal")
	if err != nil {
		logger.Log(fmt.Sprintf("Error getting updated anime data for %s (MAL ID: %d): %v", title, series.MALID, err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AnimeUpdate",
		})
		return
	}

	logger.Log(fmt.Sprintf("Successfully updated anime: %s (MAL ID: %d)", title, series.MALID), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AnimeUpdate",
	})

	// Check if the updated anime data has significant changes that warrant saving
	if shouldSaveUpdate(&series, updatedAnime) {
		// Check if anime is still airing
		if updatedAnime.Status != "RELEASING" && updatedAnime.Status != "AIRING" {
			// Update the anime data to reflect that it's no longer airing
			updatedAnime.Airing = false
		}

		if err := database.SaveAnimeToDatabase(updatedAnime); err != nil {
			logger.Log(fmt.Sprintf("Error saving updated anime data for %s (MAL ID: %d): %v", title, series.MALID, err), logger.LogOptions{
				Level:  logger.Error,
				Prefix: "AnimeUpdate",
			})
		} else {
			logger.Log(fmt.Sprintf("Successfully saved updated data for %s (MAL ID: %d)", title, series.MALID), logger.LogOptions{
				Level:  logger.Info,
				Prefix: "AnimeUpdate",
			})

			if !updatedAnime.Airing {
				logger.Log(fmt.Sprintf("Anime %s (MAL ID: %d) is no longer airing. Status: %s", title, series.MALID, updatedAnime.Status), logger.LogOptions{
					Level:  logger.Info,
					Prefix: "AnimeUpdate",
				})
			}
		}
	} else {
		logger.Log(fmt.Sprintf("No significant changes detected for %s (MAL ID: %d), skipping database update", title, series.MALID), logger.LogOptions{
			Level:  logger.Debug,
			Prefix: "AnimeUpdate",
		})
	}
}

// shouldSaveUpdate determines if the updated anime data has significant changes
// that warrant saving it to the database
func shouldSaveUpdate(oldAnime *entities.Anime, newAnime *types.Anime) bool {
	if oldAnime == nil {
		return true
	}

	// Convert old anime to types.Anime for easier comparison
	oldAnimeConverted := database.ConvertToTypesAnime(oldAnime)

	// Check for changes in next airing episode
	oldNextEp := oldAnimeConverted.NextAiringEpisode
	newNextEp := newAnime.NextAiringEpisode

	// If next episode timestamp or number changed
	if oldNextEp.AiringAt != newNextEp.AiringAt || oldNextEp.Episode != newNextEp.Episode {
		return true
	}

	// Check if sub/dub count changed
	if oldAnimeConverted.Episodes.Subbed != newAnime.Episodes.Subbed ||
		oldAnimeConverted.Episodes.Dubbed != newAnime.Episodes.Dubbed {
		return true
	}

	// Check if airing status changed
	if oldAnimeConverted.Airing != newAnime.Airing ||
		oldAnimeConverted.Status != newAnime.Status {
		return true
	}

	// Check if the total episode count has changed
	if oldAnimeConverted.Episodes.Total != newAnime.Episodes.Total {
		return true
	}

	// Check if number of episodes in the airing schedule changed
	if len(oldAnimeConverted.AiringSchedule) != len(newAnime.AiringSchedule) {
		return true
	}

	// No significant changes detected
	return false
}
