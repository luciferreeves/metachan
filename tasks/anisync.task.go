package tasks

import (
	"fmt"
	"metachan/database"
	"metachan/entities"
	"metachan/services/anime"
	"metachan/utils/logger"
	"time"
)

// AniSync fetches full anime details for all anime in the database
func AniSync() error {
	logger.Log("Starting Anime Sync - Fetching full anime details", logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AniSync",
	})

	// Get all anime mappings
	var mappings []entities.AnimeMapping
	if err := database.DB.Find(&mappings).Error; err != nil {
		logger.Log(fmt.Sprintf("Failed to fetch anime mappings: %v", err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "AniSync",
		})
		return err
	}

	total := len(mappings)
	logger.Log(fmt.Sprintf("Found %d anime mappings", total), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AniSync",
	})

	// Pre-count items needing sync for accurate ETA
	itemsToSync := 0
	for _, mapping := range mappings {
		if mapping.MAL == 0 {
			continue
		}
		var existingAnime entities.Anime
		err := database.DB.Where("mal_id = ?", mapping.MAL).First(&existingAnime).Error
		if err == nil {
			var episodeCount int64
			database.DB.Model(&entities.AnimeSingleEpisode{}).Where("anime_id = ?", existingAnime.ID).Count(&episodeCount)
			if episodeCount > 0 {
				continue
			}
		}
		itemsToSync++
	}

	logger.Log(fmt.Sprintf("Found %d anime to sync (%d already synced)", itemsToSync, total-itemsToSync), logger.LogOptions{
		Level:  logger.Info,
		Prefix: "AniSync",
	})

	animeService := anime.NewService()
	synced := 0
	skipped := 0
	startTime := time.Now()
	processed := 0

	for _, mapping := range mappings {
		// Skip if MAL ID is 0 (invalid)
		if mapping.MAL == 0 {
			skipped++
			continue
		}

		// Check if anime already exists in DB
		var existingAnime entities.Anime
		err := database.DB.Where("mal_id = ?", mapping.MAL).First(&existingAnime).Error

		if err == nil {
			// Check if anime has full details (has episodes)
			var episodeCount int64
			database.DB.Model(&entities.AnimeSingleEpisode{}).Where("anime_id = ?", existingAnime.ID).Count(&episodeCount)

			if episodeCount > 0 {
				skipped++
				continue
			}
		}

		// Calculate progress and ETA
		progress := float64(processed+1) / float64(itemsToSync) * 100
		eta := ""
		if processed >= 10 {
			elapsed := time.Since(startTime)
			avgTimePerItem := elapsed / time.Duration(processed)
			remainingItems := itemsToSync - processed
			remaining := time.Duration(remainingItems) * avgTimePerItem
			eta = formatDuration(remaining)
		} else {
			eta = "calculating..."
		}

		// Fetch full anime details
		logger.Log(fmt.Sprintf("[%d/%d] Synchronising MAL ID %d - %.1f%% | ETA: %s", processed+1, itemsToSync, mapping.MAL, progress, eta), logger.LogOptions{
			Level:  logger.Info,
			Prefix: "AniSync",
		})

		_, err = animeService.GetAnimeDetailsWithSource(&mapping, "anisync")
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to sync anime MAL ID %d: %v", mapping.MAL, err), logger.LogOptions{
				Level:  logger.Warn,
				Prefix: "AniSync",
			})
			continue
		}

		synced++
		processed++

		// Sleep to respect rate limits (1 second between requests)
		time.Sleep(1 * time.Second)
	}

	logger.Log(fmt.Sprintf("Anime Sync completed: %d synced, %d skipped", synced, skipped), logger.LogOptions{
		Level:  logger.Success,
		Prefix: "AniSync",
	})

	return nil
}

// formatDuration converts duration to human-readable format
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
