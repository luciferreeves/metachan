package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"metachan/entities"
	"metachan/enums"
	"metachan/repositories"
	"metachan/types"
	"metachan/utils/logger"
	"metachan/utils/mappers"
	"net/http"
)

const (
	fribbURL  = "https://raw.githubusercontent.com/Fribb/anime-lists/master/anime-list-full.json"
	batchSize = 1000
)

func AniFetch() error {
	logger.Infof("AniFetch", "Starting Anime Fetch")

	response, err := http.Get(fribbURL)
	if err != nil {
		logger.Errorf("AniFetch", "Anime Fetch failed: %v", err)
		return err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Errorf("AniFetch", "Failed to read response body: %v", err)
		return err
	}

	var mappings []types.MappingResponse
	if err := json.Unmarshal(body, &mappings); err != nil {
		logger.Errorf("AniFetch", "Failed to unmarshal JSON: %v", err)
		return err
	}

	total := len(mappings)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := mappings[i:end]
		processBatch(batch)
		logger.Infof("AniFetch", "Processed %d/%d mappings", end, total)
	}

	logger.Successf("AniFetch", "Anime Fetch completed")

	return nil
}

func processBatch(mappings []types.MappingResponse) {
	for _, mapping := range mappings {
		var composite *string
		if mapping.MAL != 0 && mapping.Anilist != 0 {
			comp := fmt.Sprintf("%d-%d", mapping.MAL, mapping.Anilist)
			composite = &comp
		}

		entity := entities.Mapping{
			AniDB:               mapping.AniDB,
			Anilist:             mapping.Anilist,
			AnimeCountdown:      mapping.AnimeCountdown,
			AnimePlanet:         mappers.ForceString(mapping.AnimePlanet),
			AniSearch:           mapping.AniSearch,
			IMDB:                mapping.IMDB,
			Kitsu:               mapping.Kitsu,
			LiveChart:           mapping.LiveChart,
			MAL:                 mapping.MAL,
			NotifyMoe:           mapping.NotifyMoe,
			Simkl:               mapping.Simkl,
			TMDB:                mappers.ForceInt(mapping.TMDB),
			TVDB:                mapping.TVDB,
			Type:                enums.MappingAnimeType(mapping.Type),
			MALAnilistComposite: composite,
		}

		if err := repositories.CreateOrUpdateMapping(&entity); err != nil {
			logger.Warnf("AniFetch", "Unable to process mapping %v: %v", mapping, err)
		}
	}
}
