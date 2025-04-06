package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"metachan/database"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
	"metachan/utils/mappers"
	"net/http"

	"gorm.io/gorm"
)

const fribbURL = "https://raw.githubusercontent.com/Fribb/anime-lists/master/anime-list-full.json"

func AniSync() error {
	logger.Log("Starting Anime Sync", types.LogOptions{
		Level:  types.Info,
		Prefix: "AniSync",
	})

	response, err := http.Get(fribbURL)
	if err != nil {
		logger.Log(fmt.Sprintf("Anime Sync failed: %v", err), types.LogOptions{
			Level:  types.Error,
			Prefix: "AniSync",
		})
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to read response body: %v", err), types.LogOptions{
			Level:  types.Error,
			Prefix: "AniSync",
		})
		return err
	}

	var mappings []types.AniSyncMapping
	if err := json.Unmarshal(body, &mappings); err != nil {
		logger.Log(fmt.Sprintf("Failed to unmarshal JSON: %v", err), types.LogOptions{
			Level:  types.Error,
			Prefix: "AniSync",
		})
		return err
	}

	batchSize := 1000
	total := len(mappings)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := mappings[i:end]
		processBatch(batch)
		logger.Log(fmt.Sprintf("Processed %d/%d mappings", end, total), types.LogOptions{
			Level:  types.Info,
			Prefix: "AniSync",
		})
	}

	logger.Log("Anime Sync completed", types.LogOptions{
		Level:  types.Success,
		Prefix: "AniSync",
	})

	return nil
}

func processBatch(mappings []types.AniSyncMapping) {
	for _, mapping := range mappings {
		var composite *string
		if mapping.MAL != 0 && mapping.Anilist != 0 {
			comp := fmt.Sprintf("%d-%d", mapping.MAL, mapping.Anilist)
			composite = &comp
		}

		var entity entities.AnimeMapping
		if err := database.DB.Where("mal_anilist_composite = ?", composite).First(&entity).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				newEntity := entities.AnimeMapping{
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
					Type:                entities.MappingType(mapping.Type),
					MALAnilistComposite: composite,
				}
				if err := database.DB.Create(&newEntity).Error; err != nil {
					logger.Log(fmt.Sprintf("Unable to process mapping %v: %v", mapping, err), types.LogOptions{
						Level:  types.Warn,
						Prefix: "AniSync",
					})
				}
			} else {
				logger.Log(fmt.Sprintf("Error fetching entity: %v", err), types.LogOptions{
					Level:  types.Error,
					Prefix: "AniSync",
				})
			}
		} else {
			// Update existing entity
			entity.AniDB = mapping.AniDB
			entity.Anilist = mapping.Anilist
			entity.AnimeCountdown = mapping.AnimeCountdown
			entity.AnimePlanet = mappers.ForceString(mapping.AnimePlanet)
			entity.AniSearch = mapping.AniSearch
			entity.IMDB = mapping.IMDB
			entity.Kitsu = mapping.Kitsu
			entity.LiveChart = mapping.LiveChart
			entity.MAL = mapping.MAL
			entity.NotifyMoe = mapping.NotifyMoe
			entity.Simkl = mapping.Simkl
			entity.TMDB = mappers.ForceInt(mapping.TMDB)
			entity.TVDB = mapping.TVDB
			entity.Type = entities.MappingType(mapping.Type)
			entity.MALAnilistComposite = composite
			if err := database.DB.Save(&entity).Error; err != nil {
				logger.Log(fmt.Sprintf("Unable to update mapping %v: %v", mapping, err), types.LogOptions{
					Level:  types.Warn,
					Prefix: "AniSync",
				})
			}
		}
	}
}
