package tasks

import (
	"fmt"
	"metachan/config"
	"metachan/entities"
	"metachan/enums"
	"metachan/repositories"
	"metachan/services"
	"metachan/utils/logger"
	"sync"
	"time"
)

const (
	UpdaterSource = "updater"

	MaxConcurrentUpdates = 5

	MaxConcurrentSQLiteUpdates = 1

	UpdateInterval = 6 * time.Hour
)

type animeUpdateJob struct {
	series entities.Anime
	reason string
}

func AnimeUpdate() error {
	logger.Infof("AnimeUpdate", "Starting Anime Update Task")

	airingSeries, err := repositories.GetAiringAnime()
	if err != nil {
		logger.Errorf("AnimeUpdate", "Failed to fetch airing anime: %v", err)
		return err
	}

	logger.Infof("AnimeUpdate", "Found %d airing anime series", len(airingSeries))

	currentTime := time.Now().Unix()

	logger.Debugf("AnimeUpdate", "Current timestamp: %d (%s)", currentTime, time.Unix(currentTime, 0).Format(time.RFC3339))

	jobs := make(chan animeUpdateJob, len(airingSeries))

	maxWorkers := MaxConcurrentUpdates
	if config.Database.Driver == "sqlite" {
		maxWorkers = MaxConcurrentSQLiteUpdates
		logger.Debugf("AnimeUpdate", "Using reduced concurrency (%d workers) for SQLite database", maxWorkers)
	}

	var wg sync.WaitGroup

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			logger.Debugf("AnimeUpdate", "Started worker #%d", workerID+1)

			for job := range jobs {
				updateAnime(job.series, job.reason)
			}
		}(i)
	}

	jobsQueued := 0
	for _, series := range airingSeries {
		needsUpdate := false
		reason := ""

		title := series.Title.Romaji
		if title == "" {
			title = series.Title.English
		}

		logger.Debugf("AnimeUpdate", "Checking anime: %s (ID: %d)", title, series.MALID)

		if series.NextAiringAt == 0 {
			needsUpdate = true
			reason = "missing next episode data"
		} else if int64(series.NextAiringAt) <= currentTime {
			needsUpdate = true
			reason = "next episode already aired"
		}

		if !needsUpdate && !series.LastUpdated.IsZero() && time.Since(series.LastUpdated) > UpdateInterval {
			needsUpdate = true
			reason = fmt.Sprintf("regular update (last updated %s ago)",
				time.Since(series.LastUpdated).Round(time.Second))
		}

		if !needsUpdate {
			logger.Debugf("AnimeUpdate", "Skipping update for %s (ID: %d) - no update needed. Next airing at: %d",
				title, series.MALID, series.NextAiringAt)
			continue
		}

		logger.Debugf("AnimeUpdate", "Queueing update for %s (ID: %d) - Reason: %s",
			title, series.MALID, reason)

		jobs <- animeUpdateJob{
			series: series,
			reason: reason,
		}
		jobsQueued++
	}

	close(jobs)

	wg.Wait()

	logger.Successf("AnimeUpdate", "Anime Update Task Completed - Processed %d anime", jobsQueued)

	return nil
}

func updateAnime(series entities.Anime, reason string) {
	title := series.Title.English
	if title == "" {
		title = series.Title.Romaji
	}

	logger.Infof("AnimeUpdate", "Updating anime: %s (MAL ID: %d) - %s", title, series.MALID, reason)

	mapping, err := repositories.GetAnimeMapping(enums.MAL, series.MALID)
	if err != nil {
		logger.Errorf("AnimeUpdate", "Error getting anime mapping for %s (MAL ID: %d): %v", title, series.MALID, err)
		return
	}

	updatedAnime, err := services.ForceRefreshAnime(&mapping)
	if err != nil {
		logger.Errorf("AnimeUpdate", "Error getting updated anime data for %s (MAL ID: %d): %v", title, series.MALID, err)
		return
	}

	logger.Successf("AnimeUpdate", "Successfully updated anime: %s (MAL ID: %d)", title, series.MALID)

	if shouldSaveUpdate(&series, updatedAnime) {
		if updatedAnime.Status != "RELEASING" && updatedAnime.Status != "AIRING" {
			updatedAnime.Airing = false
		}

		if err := repositories.CreateOrUpdateAnime(updatedAnime); err != nil {
			logger.Errorf("AnimeUpdate", "Error saving updated anime data for %s (MAL ID: %d): %v", title, series.MALID, err)
		} else {
			logger.Infof("AnimeUpdate", "Successfully saved updated data for %s (MAL ID: %d)", title, series.MALID)

			if !updatedAnime.Airing {
				logger.Infof("AnimeUpdate", "Anime %s (MAL ID: %d) is no longer airing. Status: %s", title, series.MALID, updatedAnime.Status)
			}
		}
	} else {
		logger.Debugf("AnimeUpdate", "No significant changes detected for %s (MAL ID: %d), skipping database update", title, series.MALID)
	}
}

func shouldSaveUpdate(oldAnime *entities.Anime, newAnime *entities.Anime) bool {
	if oldAnime == nil {
		return true
	}

	oldHasNext := oldAnime.NextAiringAt > 0
	newHasNext := newAnime.NextAiringAt > 0

	if oldHasNext != newHasNext {
		return true
	}

	if oldHasNext && newHasNext {
		if oldAnime.NextAiringAt != newAnime.NextAiringAt ||
			oldAnime.NextAiringEpisode != newAnime.NextAiringEpisode {
			return true
		}
	}

	if oldAnime.SubbedCount != newAnime.SubbedCount ||
		oldAnime.DubbedCount != newAnime.DubbedCount {
		return true
	}

	if oldAnime.Airing != newAnime.Airing ||
		oldAnime.Status != newAnime.Status {
		return true
	}

	if oldAnime.TotalEpisodes != newAnime.TotalEpisodes {
		return true
	}

	if len(oldAnime.Schedule) != len(newAnime.Schedule) {
		return true
	}

	return false
}
