package tasks

import (
	"metachan/enums"
	"metachan/repositories"
	"metachan/services"
	"metachan/utils/logger"
	"time"
)

func AniSync() error {
	logger.Infof("AniSync", "Starting Anime Sync - Fetching full anime details")

	mappings, err := repositories.GetAllMappings()
	if err != nil {
		logger.Errorf("AniSync", "Failed to fetch anime mappings: %v", err)
		return err
	}

	total := len(mappings)
	logger.Infof("AniSync", "Found %d anime mappings", total)

	itemsToSync := 0
	for _, mapping := range mappings {
		if mapping.MAL == 0 {
			continue
		}
		_, err := repositories.GetAnime(enums.MAL, mapping.MAL)
		if err == nil {
			continue
		}
		itemsToSync++
	}

	logger.Infof("AniSync", "Found %d anime to sync (%d already synced)", itemsToSync, total-itemsToSync)

	synced := 0
	skipped := 0
	startTime := time.Now()
	processed := 0

	for _, mapping := range mappings {
		if mapping.MAL == 0 {
			skipped++
			continue
		}

		_, err := repositories.GetAnime(enums.MAL, mapping.MAL)
		if err == nil {
			skipped++
			continue
		}

		progress, eta := calculateProgress(processed+1, itemsToSync, startTime)

		logger.Infof("AniSync", "[%d/%d] Synchronising MAL ID %d - %.1f%% | ETA: %v", processed+1, itemsToSync, mapping.MAL, progress, eta)

		_, err = services.GetAnime(&mapping)
		if err != nil {
			logger.Warnf("AniSync", "Failed to sync anime MAL ID %d: %v", mapping.MAL, err)
			continue
		}

		synced++
		processed++
	}

	return nil
}
