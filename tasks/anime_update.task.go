package tasks

import (
	"fmt"
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
)

// animeUpdateJob represents a single anime update job
type animeUpdateJob struct {
	series entities.CachedAnime
	reason string
}

// AnimeUpdate checks for airing anime that need to be updated
func AnimeUpdate() error {
	logger.Log("Starting Anime Update Task", logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AnimeUpdate",
	})

	// Find all currently airing anime
	var airingSeries []entities.CachedAnime
	result := database.DB.
		Where("airing = ?", true).
		Preload("NextAiringEpisode").
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

	// Create a channel for jobs
	jobs := make(chan animeUpdateJob, len(airingSeries))

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Create workers
	for i := 0; i < MaxConcurrentUpdates; i++ {
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

		// If there's no next airing episode data, we should update
		if series.NextAiringEpisode == nil {
			needsUpdate = true
			reason = "missing next episode data"
		} else {
			// Check if the next episode has already aired
			if int64(series.NextAiringEpisode.AiringAt) < currentTime {
				needsUpdate = true
				reason = fmt.Sprintf("episode %d aired at %d (current: %d)",
					series.NextAiringEpisode.Episode,
					series.NextAiringEpisode.AiringAt,
					currentTime)
			}
		}

		// Skip if no update is needed
		if !needsUpdate {
			continue
		}

		// Add the job to the queue
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

// updateAnime handles updating a single anime
func updateAnime(animeService *anime.Service, series entities.CachedAnime, reason string) {
	// Get the mapping for this anime
	mapping, err := database.GetAnimeMappingViaMALID(series.MALID)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get mapping for MALID %d: %v", series.MALID, err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AnimeUpdate",
		})
		return
	}

	logger.Log(fmt.Sprintf("Updating anime %s (MALID: %d) - Reason: %s",
		series.TitleRomaji, series.MALID, reason), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AnimeUpdate",
	})

	// Get fresh anime data
	updatedAnime, err := animeService.GetAnimeDetailsWithSource(mapping, UpdaterSource)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to get fresh data for MALID %d: %v", series.MALID, err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AnimeUpdate",
		})
		return
	}

	// Check if significant changes occurred to save the update
	saved := false
	oldCachedAnime, err := database.GetCachedAnimeByMALID(series.MALID)

	if err != nil || shouldSaveUpdate(oldCachedAnime, updatedAnime) {
		// Save the updated data to cache
		if err := database.SaveAnimeToCache(updatedAnime); err != nil {
			logger.Log(fmt.Sprintf("Failed to save updated data for MALID %d: %v", series.MALID, err), logger.LogOptions{
				Level:  logger.Error,
				Prefix: "AnimeUpdate",
			})
			return
		}
		saved = true
	}

	status := "skipped (no significant changes)"
	if saved {
		status = "saved to database"
	}

	logger.Log(fmt.Sprintf("Update for %s (MALID: %d) complete - %s",
		series.TitleRomaji, series.MALID, status), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AnimeUpdate",
	})
}

// shouldSaveUpdate determines if the updated anime data has significant changes
// that warrant saving it to the database
func shouldSaveUpdate(oldAnime *entities.CachedAnime, newAnime *types.Anime) bool {
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

	// No significant changes detected
	return false
}
