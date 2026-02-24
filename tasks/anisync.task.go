package tasks

import (
	"metachan/entities"
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

// ResumeAnimeSync is called on startup to resume any interrupted sync and refresh stale entries.
func ResumeAnimeSync() {
	go func() {
		mappings, err := repositories.GetAllMappings()
		if err != nil {
			logger.Errorf("AniSync", "Resume: failed to fetch mappings: %v", err)
			return
		}

		stubs, err := repositories.GetAllAnimeStubs()
		if err != nil {
			logger.Errorf("AniSync", "Resume: failed to fetch anime stubs: %v", err)
			return
		}

		sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
		updatedAt := make(map[int]time.Time, len(stubs))
		enrichedAt := make(map[int]*time.Time, len(stubs))
		for _, s := range stubs {
			updatedAt[s.MALID] = s.UpdatedAt
			enrichedAt[s.MALID] = s.EnrichedAt
		}

		var toProcess []entities.Mapping
		for _, m := range mappings {
			if m.MAL == 0 {
				continue
			}
			t, exists := updatedAt[m.MAL]
			if !exists || t.Before(sevenDaysAgo) || enrichedAt[m.MAL] == nil {
				toProcess = append(toProcess, m)
			}
		}

		if len(toProcess) == 0 {
			return
		}

		logger.Infof("AniSync", "Resume: %d anime to sync (missing or stale)", len(toProcess))
		startTime := time.Now()

		for i, m := range toProcess {
			progress, eta := calculateProgress(i+1, len(toProcess), startTime)
			logger.Infof("AniSync", "[Resume %d/%d] MAL ID %d - %.1f%% | ETA: %v", i+1, len(toProcess), m.MAL, progress, eta)

			if _, err := services.GetAnime(&m); err != nil {
				logger.Warnf("AniSync", "Resume: failed to sync MAL ID %d: %v", m.MAL, err)
			}
		}

		logger.Successf("AniSync", "Resume complete: synced %d anime", len(toProcess))
	}()
}
