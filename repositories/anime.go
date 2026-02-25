package repositories

import (
	"errors"
	"fmt"
	"metachan/entities"
	"metachan/enums"
	"metachan/utils/logger"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAnime[T idType](maptype enums.MappingType, id T) (entities.Anime, error) {
	var anime entities.Anime

	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		logger.Errorf("Anime", "Failed to get anime mapping: %v", err)
		return entities.Anime{}, errors.New("anime not found")
	}

	result := DB.
		Preload("Mapping").
		Preload("Genres").
		Preload("Themes").
		Preload("Demographics").
		Preload("Producers").
		Preload("Producers.Image").
		Preload("Producers.Titles").
		Preload("Producers.ExternalURLs").
		Preload("Studios").
		Preload("Studios.Image").
		Preload("Studios.Titles").
		Preload("Studios.ExternalURLs").
		Preload("Licensors").
		Preload("Licensors.Image").
		Preload("Licensors.Titles").
		Preload("Licensors.ExternalURLs").
		Preload("Episodes").
		Preload("Episodes.SkipTimes").
		Preload("Episodes.StreamInfo").
		Preload("Episodes.StreamInfo.SubSources").
		Preload("Episodes.StreamInfo.DubSources").
		Preload("Schedule").
		Preload("Seasons").
		Where("mapping_id = ?", mapping.ID).
		First(&anime)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return entities.Anime{}, result.Error
		}
		logger.Errorf("Anime", "Failed to get anime details: %v", result.Error)
		return entities.Anime{}, errors.New("anime not found")
	}

	loadAnimeCharacters(&anime)

	return anime, nil
}

func CreateOrUpdateAnime(anime *entities.Anime) error {
	if anime == nil {
		return fmt.Errorf("anime is nil")
	}

	var existingAnime entities.Anime
	result := DB.Where("mal_id = ?", anime.MALID).First(&existingAnime)
	if result.Error == nil {
		anime.ID = existingAnime.ID
	}

	now := time.Now()
	anime.LastUpdated = now

	result = DB.Session(&gorm.Session{FullSaveAssociations: true}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Omit("Characters", "Episodes").Save(anime)

	if result.Error != nil {
		return fmt.Errorf("failed to save anime: %w", result.Error)
	}

	logger.Infof("Anime", "Saved anime (MAL ID: %d) with %d episodes, %d characters", anime.MALID, len(anime.Episodes), len(anime.Characters))
	return nil
}

func SaveAnimeEpisodes(animeID uint, episodes []entities.Episode) error {
	for i := range episodes {
		ep := &episodes[i]
		ep.AnimeID = animeID

		var existing entities.Episode
		if DB.Where("episode_id = ?", ep.EpisodeID).First(&existing).Error == nil {
			ep.ID = existing.ID
			DB.Model(ep).Omit("SkipTimes", "StreamInfo").Updates(ep)
		} else {
			DB.Session(&gorm.Session{FullSaveAssociations: true}).
				Omit("SkipTimes", "StreamInfo").
				Create(ep)
		}
	}
	return nil
}

func SaveEpisodeSkipTimes(episodeID string, skipTimes []entities.EpisodeSkipTime) error {
	if len(skipTimes) == 0 {
		return nil
	}

	DB.Where("episode_id = ?", episodeID).Delete(&entities.EpisodeSkipTime{})

	for i := range skipTimes {
		skipTimes[i].EpisodeID = episodeID
		if err := DB.Create(&skipTimes[i]).Error; err != nil {
			return fmt.Errorf("failed to save skip time: %w", err)
		}
	}

	return nil
}

func GetAnimeEpisode[T idType](maptype enums.MappingType, id T, episodeID string) (entities.Episode, error) {
	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		return entities.Episode{}, errors.New("anime not found")
	}

	var anime entities.Anime
	if err := DB.Where("mapping_id = ?", mapping.ID).Select("id").First(&anime).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.Episode{}, err
		}
		return entities.Episode{}, errors.New("anime not found")
	}

	var episode entities.Episode
	result := DB.
		Preload("SkipTimes").
		Preload("StreamInfo").
		Preload("StreamInfo.SubSources").
		Preload("StreamInfo.DubSources").
		Where("anime_id = ? AND episode_id = ?", anime.ID, episodeID).
		First(&episode)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return entities.Episode{}, result.Error
		}
		return entities.Episode{}, errors.New("failed to fetch episode")
	}

	return episode, nil
}

func GetAnimeEpisodes[T idType](maptype enums.MappingType, id T) ([]entities.Episode, error) {
	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		return nil, errors.New("anime not found")
	}

	var anime entities.Anime
	if err := DB.Where("mapping_id = ?", mapping.ID).Select("id").First(&anime).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("anime not found")
	}

	var episodes []entities.Episode
	result := DB.
		Preload("SkipTimes").
		Preload("StreamInfo").
		Preload("StreamInfo.SubSources").
		Preload("StreamInfo.DubSources").
		Where("anime_id = ?", anime.ID).
		Order("episode_number asc").
		Find(&episodes)

	if result.Error != nil {
		return nil, errors.New("failed to fetch episodes")
	}

	return episodes, nil
}

func SaveEpisodeStreamInfo(animeID uint, episodeID string, info *entities.StreamInfo) error {
	info.AnimeID = animeID
	info.EpisodeID = episodeID
	info.LastFetch = time.Now()

	var existing entities.StreamInfo
	if DB.Where("episode_id = ? AND anime_id = ?", episodeID, animeID).First(&existing).Error == nil {
		info.ID = existing.ID
		DB.Where("stream_info_id = ?", existing.ID).Delete(&entities.StreamingSource{})
	}

	return DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(info).Error
}

func GetAllAnimeStubs() ([]animeStub, error) {
	var stubs []animeStub
	if err := DB.Model(&entities.Anime{}).Select("mal_id, updated_at, enriched_at").Scan(&stubs).Error; err != nil {
		return nil, err
	}
	return stubs, nil
}

func SetAnimeEnriched(malID int) error {
	now := time.Now()
	return DB.Model(&entities.Anime{}).Where("mal_id = ?", malID).Update("enriched_at", now).Error
}

func GetAiringAnime() ([]entities.Anime, error) {
	var anime []entities.Anime

	result := DB.
		Where("airing = ?", true).
		Preload("Schedule").
		Find(&anime)

	if result.Error != nil {
		logger.Errorf("Anime", "Failed to fetch airing anime: %v", result.Error)
		return nil, errors.New("failed to fetch airing anime")
	}

	return anime, nil
}
