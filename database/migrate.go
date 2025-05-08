package database

import (
	"fmt"
	"metachan/entities"
	"metachan/utils/logger"
)

func AutoMigrate() {
	err := DB.AutoMigrate(
		// Base entities
		&entities.TaskLog{},
		&entities.AnimeMapping{},

		// Cache entities
		&entities.CachedAnime{},
		&entities.CachedAnimeImages{},
		&entities.CachedAnimeLogos{},
		&entities.CachedAnimeCovers{},
		&entities.CachedAnimeScores{},
		&entities.CachedAiringStatusDates{},
		&entities.CachedAiringStatus{},
		&entities.CachedAnimeBroadcast{},
		&entities.CachedAnimeGenre{},
		&entities.CachedAnimeProducer{},
		&entities.CachedAnimeStudio{},
		&entities.CachedAnimeLicensor{},
		&entities.CachedAiringEpisode{},
		&entities.CachedEpisodeTitles{},
		&entities.CachedAnimeSingleEpisode{},
		&entities.CachedAnimeCharacter{},
		&entities.CachedAnimeVoiceActor{},
		&entities.CachedAnimeSeason{},
	)
	if err != nil {
		logger.Log(fmt.Sprintf("Error during auto migration: %v", err), logger.LogOptions{
			Prefix: "Database",
			Level:  logger.Error,
			Fatal:  true,
		})
	} else {
		logger.Log("Auto migration completed successfully", logger.LogOptions{
			Prefix: "Database",
			Level:  logger.Success,
		})
	}
}

// Migrate creates and migrations all tables
func Migrate() {
	// Use AutoMigrate to ensure consistent behavior
	AutoMigrate()
}
