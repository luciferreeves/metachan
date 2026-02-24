package database

import (
	"metachan/entities"
	"metachan/utils/logger"
)

func migrate() {
	err := DB.AutoMigrate(
		&entities.TaskLog{},
		&entities.TaskStatus{},
		&entities.Mapping{},
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
		&entities.Genre{},
		&entities.Producer{},
		&entities.Anime{},
		&entities.Episode{},
		&entities.EpisodeSkipTime{},
		&entities.StreamingSource{},
		&entities.EpisodeSchedule{},
		&entities.NextEpisode{},
		&entities.Season{},
		&entities.Character{},
		&entities.Person{},
		&entities.AnimeCharacter{},
		&entities.CharacterVoiceActor{},
		&entities.CharacterAnimeAppearance{},
		&entities.PersonVoiceRole{},
		&entities.PersonAnimeCredit{},
		&entities.PersonMangaCredit{},
	)
	if err != nil {
		logger.Fatalf("Database", "Error during database migration: %v", err)
	}

	logger.Successf("Database", "Database migration completed successfully")
}
