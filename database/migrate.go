package database

import (
	"metachan/entities"
	"metachan/utils/logger"
)

func migrate() {
	err := DB.AutoMigrate(
		// Task entities
		&entities.TaskLog{},
		&entities.TaskStatus{},

		// Mapping entity
		&entities.Mapping{},

		// Meta entities (shared/reusable)
		&entities.Title{},
		&entities.Scores{},
		&entities.Date{},
		&entities.AiringStatus{},
		&entities.Broadcast{},
		&entities.Images{},
		&entities.Logos{},
		&entities.ExternalURL{},
		&entities.SimpleTitle{},
		&entities.SimpleImage{},

		// Genre entity
		&entities.Genre{},

		// Producer entity
		&entities.Producer{},

		// Anime entity
		&entities.Anime{},

		// Episode entities
		&entities.Episode{},
		&entities.EpisodeSkipTime{},
		&entities.StreamingSource{},
		&entities.EpisodeSchedule{},
		&entities.NextEpisode{},

		// Season entity
		&entities.Season{},

		// Character/Persona entities
		&entities.Character{},
		&entities.VoiceActor{},
		&entities.AnimeCharacter{},
		&entities.CharacterVoiceActor{},
	)
	if err != nil {
		logger.Fatalf("Database", "Error during database migration: %v", err)
	}

	logger.Successf("Database", "Database migration completed successfully")
}
