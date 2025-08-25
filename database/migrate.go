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
		&entities.CachedNextEpisode{},
		&entities.CachedScheduleEpisode{},
		&entities.CachedEpisodeTitles{},
		&entities.CachedAnimeSingleEpisode{},
		&entities.CachedAnimeCharacter{},
		&entities.CachedAnimeVoiceActor{},
		&entities.CachedAnimeSeason{},
	)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to migrate database: %v", err), logger.LogOptions{
			Level:  logger.Error,
			Prefix: "Database",
		})
		panic(err)
	}

	logger.Log("Database migration completed successfully", logger.LogOptions{
		Level:  logger.Info,
		Prefix: "Database",
	})
}

// Migrate creates and migrations all tables
func Migrate() {
	// Use AutoMigrate to ensure consistent behavior
	AutoMigrate()
}
