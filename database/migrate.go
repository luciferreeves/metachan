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

		// Anime entities
		&entities.Anime{},
		&entities.AnimeImages{},
		&entities.AnimeLogos{},
		&entities.AnimeCovers{},
		&entities.AnimeScores{},
		&entities.AiringStatusDates{},
		&entities.AiringStatus{},
		&entities.AnimeBroadcast{},
		&entities.AnimeGenre{},
		&entities.AnimeProducer{},
		&entities.AnimeStudio{},
		&entities.AnimeLicensor{},
		&entities.NextEpisode{},
		&entities.ScheduleEpisode{},
		&entities.EpisodeTitles{},
		&entities.AnimeSingleEpisode{},
		&entities.AnimeCharacter{},
		&entities.AnimeVoiceActor{},
		&entities.AnimeSeason{},

		// Streaming entities
		&entities.EpisodeStreaming{},
		&entities.EpisodeStreamingSource{},
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
