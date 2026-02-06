package tasks

import (
	"metachan/entities"
	"metachan/repositories"
	"metachan/utils/api/jikan"
	"metachan/utils/logger"
)

func GenreSync() error {
	logger.Infof("GenreSync", "Starting Genre Sync from MAL")

	genresResponse, err := jikan.GetAnimeGenres()
	if err != nil {
		logger.Errorf("GenreSync", "Failed to fetch genres from MAL: %v", err)
		return err
	}

	logger.Infof("GenreSync", "Fetched %d genres from MAL", len(genresResponse.Data))

	for _, genre := range genresResponse.Data {
		genreEntity := entities.Genre{
			GenreID: genre.MALID,
			Name:    genre.Name,
			URL:     genre.URL,
			Count:   genre.Count,
		}

		if err := repositories.CreateOrUpdateGenre(&genreEntity); err != nil {
			logger.Warnf("GenreSync", "Failed to sync genre %s: %v", genre.Name, err)
		}
	}

	logger.Successf("GenreSync", "Genre Sync completed successfully. Synced %d genres", len(genresResponse.Data))
	return nil
}
