package database

import (
	"fmt"
	"metachan/entities"
	"metachan/types"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	CacheExpirationTime = 24 * time.Hour
)

func GetCachedAnimeByMALID(malID int) (*entities.CachedAnime, error) {
	var anime entities.CachedAnime
	err := DB.
		Preload("Images").
		Preload("Logos").
		Preload("Covers").
		Preload("Scores").
		Preload("AiringStatus").
		Preload("AiringStatus.From").
		Preload("AiringStatus.To").
		Preload("Broadcast").
		Preload("NextAiringEpisode").
		Preload("Genres").
		Preload("Producers").
		Preload("Studios").
		Preload("Licensors").
		Preload("Episodes").
		Preload("Episodes.Titles").
		Preload("Characters").
		Preload("Characters.VoiceActors").
		Preload("AiringSchedule").
		Preload("Seasons").
		Where("mal_id = ?", malID).
		First(&anime).Error

	if err != nil {
		return nil, err
	}
	return &anime, nil
}

func IsCacheValid(anime *entities.CachedAnime) bool {
	if anime == nil {
		return false
	}
	return time.Since(anime.LastUpdated) < CacheExpirationTime
}

func SaveAnimeToCache(animeData *types.Anime) error {
	if animeData == nil {
		return fmt.Errorf("anime data is nil")
	}

	var tx *gorm.DB = DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Exec("DELETE FROM cached_animes WHERE mal_id = ?", animeData.MALID).Error; err != nil {
		tx.Rollback()
		return err
	}

	anime := &entities.CachedAnime{
		MALID:         animeData.MALID,
		TitleRomaji:   animeData.Titles.Romaji,
		TitleEnglish:  animeData.Titles.English,
		TitleJapanese: animeData.Titles.Japanese,
		TitleSynonyms: strings.Join(animeData.Titles.Synonyms, ","),
		Synopsis:      animeData.Synopsis,
		Type:          string(animeData.Type),
		Source:        animeData.Source,
		Airing:        animeData.Airing,
		Status:        animeData.Status,
		Duration:      animeData.Duration,
		Color:         animeData.Color,
		Season:        animeData.Season,
		Year:          animeData.Year,
		SubbedCount:   animeData.Episodes.Subbed,
		DubbedCount:   animeData.Episodes.Dubbed,
		TotalEpisodes: animeData.Episodes.Total,
		AiredEpisodes: animeData.Episodes.Aired,
		LastUpdated:   time.Now(),
	}

	if err := tx.Create(anime).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Now create related records with proper foreign key references
	if animeData.Images.Small != "" || animeData.Images.Large != "" || animeData.Images.Original != "" {
		images := &entities.CachedAnimeImages{
			AnimeID:  anime.ID,
			Small:    animeData.Images.Small,
			Large:    animeData.Images.Large,
			Original: animeData.Images.Original,
		}
		if err := tx.Create(images).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if animeData.Logos.Small != "" || animeData.Logos.Medium != "" || animeData.Logos.Large != "" || animeData.Logos.XLarge != "" || animeData.Logos.Original != "" {
		logos := &entities.CachedAnimeLogos{
			AnimeID:  anime.ID,
			Small:    animeData.Logos.Small,
			Medium:   animeData.Logos.Medium,
			Large:    animeData.Logos.Large,
			XLarge:   animeData.Logos.XLarge,
			Original: animeData.Logos.Original,
		}
		if err := tx.Create(logos).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if animeData.Covers.Small != "" || animeData.Covers.Large != "" || animeData.Covers.Original != "" {
		covers := &entities.CachedAnimeCovers{
			AnimeID:  anime.ID,
			Small:    animeData.Covers.Small,
			Large:    animeData.Covers.Large,
			Original: animeData.Covers.Original,
		}
		if err := tx.Create(covers).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if animeData.Scores.Score > 0 || animeData.Scores.ScoredBy > 0 {
		scores := &entities.CachedAnimeScores{
			AnimeID:    anime.ID,
			Score:      animeData.Scores.Score,
			ScoredBy:   animeData.Scores.ScoredBy,
			Rank:       animeData.Scores.Rank,
			Popularity: animeData.Scores.Popularity,
			Members:    animeData.Scores.Members,
			Favorites:  animeData.Scores.Favorites,
		}
		if err := tx.Create(scores).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if animeData.AiringStatus.From.Year > 0 || animeData.AiringStatus.From.String != "" || animeData.AiringStatus.To.Year > 0 || animeData.AiringStatus.To.String != "" || animeData.AiringStatus.String != "" {
		var fromDateID, toDateID *uint

		if animeData.AiringStatus.From.Year > 0 || animeData.AiringStatus.From.String != "" {
			fromDate := &entities.CachedAiringStatusDates{
				Day:    animeData.AiringStatus.From.Day,
				Month:  animeData.AiringStatus.From.Month,
				Year:   animeData.AiringStatus.From.Year,
				String: animeData.AiringStatus.From.String,
			}
			if err := tx.Create(fromDate).Error; err != nil {
				tx.Rollback()
				return err
			}
			fromDateID = &fromDate.ID
		}

		if animeData.AiringStatus.To.Year > 0 || animeData.AiringStatus.To.String != "" {
			toDate := &entities.CachedAiringStatusDates{
				Day:    animeData.AiringStatus.To.Day,
				Month:  animeData.AiringStatus.To.Month,
				Year:   animeData.AiringStatus.To.Year,
				String: animeData.AiringStatus.To.String,
			}
			if err := tx.Create(toDate).Error; err != nil {
				tx.Rollback()
				return err
			}
			toDateID = &toDate.ID
		}

		airingStatus := &entities.CachedAiringStatus{
			AnimeID: anime.ID,
			FromID:  fromDateID,
			ToID:    toDateID,
			String:  animeData.AiringStatus.String,
		}

		if err := tx.Create(airingStatus).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if animeData.Broadcast.Day != "" || animeData.Broadcast.Time != "" {
		broadcast := &entities.CachedAnimeBroadcast{
			AnimeID:  anime.ID,
			Day:      animeData.Broadcast.Day,
			Time:     animeData.Broadcast.Time,
			Timezone: animeData.Broadcast.Timezone,
			String:   animeData.Broadcast.String,
		}
		if err := tx.Create(broadcast).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if animeData.NextAiringEpisode.AiringAt > 0 && animeData.NextAiringEpisode.Episode > 0 {
		nextEpisode := &entities.CachedNextEpisode{
			AnimeID:  anime.ID,
			AiringAt: animeData.NextAiringEpisode.AiringAt,
			Episode:  animeData.NextAiringEpisode.Episode,
		}
		if err := tx.Create(nextEpisode).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Create array records
	for _, genre := range animeData.Genres {
		genreRecord := &entities.CachedAnimeGenre{
			AnimeID: anime.ID,
			Name:    genre.Name,
			GenreID: genre.GenreID,
			URL:     genre.URL,
		}
		if err := tx.Create(genreRecord).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, producer := range animeData.Producers {
		producerRecord := &entities.CachedAnimeProducer{
			AnimeID:    anime.ID,
			Name:       producer.Name,
			ProducerID: producer.ProducerID,
			URL:        producer.URL,
		}
		if err := tx.Create(producerRecord).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, studio := range animeData.Studios {
		studioRecord := &entities.CachedAnimeStudio{
			AnimeID:  anime.ID,
			Name:     studio.Name,
			StudioID: studio.StudioID,
			URL:      studio.URL,
		}
		if err := tx.Create(studioRecord).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, licensor := range animeData.Licensors {
		licensorRecord := &entities.CachedAnimeLicensor{
			AnimeID:    anime.ID,
			Name:       licensor.Name,
			ProducerID: licensor.ProducerID,
			URL:        licensor.URL,
		}
		if err := tx.Create(licensorRecord).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, episode := range animeData.AiringSchedule {
		scheduleRecord := &entities.CachedScheduleEpisode{
			AnimeID:  anime.ID,
			AiringAt: episode.AiringAt,
			Episode:  episode.Episode,
		}
		if err := tx.Create(scheduleRecord).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, episode := range animeData.Episodes.Episodes {
		titles := &entities.CachedEpisodeTitles{
			English:  episode.Titles.English,
			Japanese: episode.Titles.Japanese,
			Romaji:   episode.Titles.Romaji,
		}
		if err := tx.Create(titles).Error; err != nil {
			tx.Rollback()
			return err
		}

		episodeRecord := &entities.CachedAnimeSingleEpisode{
			EpisodeID:    episode.ID,
			AnimeID:      anime.ID,
			TitlesID:     titles.ID,
			Description:  episode.Description,
			Aired:        episode.Aired,
			Score:        episode.Score,
			Filler:       episode.Filler,
			Recap:        episode.Recap,
			ForumURL:     episode.ForumURL,
			URL:          episode.URL,
			ThumbnailURL: episode.ThumbnailURL,
		}
		if err := tx.Create(episodeRecord).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, character := range animeData.Characters {
		charRecord := &entities.CachedAnimeCharacter{
			AnimeID:  anime.ID,
			MALID:    character.MALID,
			URL:      character.URL,
			ImageURL: character.ImageURL,
			Name:     character.Name,
			Role:     character.Role,
		}

		if err := tx.Create(charRecord).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, va := range character.VoiceActors {
			vaRecord := &entities.CachedAnimeVoiceActor{
				CharacterID: charRecord.ID,
				MALID:       va.MALID,
				URL:         va.URL,
				Image:       va.Image,
				Name:        va.Name,
				Language:    va.Language,
			}
			if err := tx.Create(vaRecord).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	for _, season := range animeData.Seasons {
		seasonRecord := &entities.CachedAnimeSeason{
			ParentAnimeID: anime.ID,
			MALID:         season.MALID,
			TitleRomaji:   season.Titles.Romaji,
			TitleEnglish:  season.Titles.English,
			TitleJapanese: season.Titles.Japanese,
			TitleSynonyms: strings.Join(season.Titles.Synonyms, ","),
			Synopsis:      season.Synopsis,
			Type:          string(season.Type),
			Source:        season.Source,
			Airing:        season.Airing,
			Status:        season.Status,
			Duration:      season.Duration,
			Season:        season.Season,
			Year:          season.Year,
			Current:       season.Current,
		}

		if err := tx.Create(seasonRecord).Error; err != nil {
			tx.Rollback()
			return err
		}

		if season.Images.Small != "" || season.Images.Large != "" || season.Images.Original != "" {
			seasonImages := &entities.CachedAnimeImages{
				AnimeID:  seasonRecord.ID,
				Small:    season.Images.Small,
				Large:    season.Images.Large,
				Original: season.Images.Original,
			}
			if err := tx.Create(seasonImages).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		if season.Scores.Score > 0 || season.Scores.ScoredBy > 0 {
			seasonScores := &entities.CachedAnimeScores{
				AnimeID:    seasonRecord.ID,
				Score:      season.Scores.Score,
				ScoredBy:   season.Scores.ScoredBy,
				Rank:       season.Scores.Rank,
				Popularity: season.Scores.Popularity,
				Members:    season.Scores.Members,
				Favorites:  season.Scores.Favorites,
			}
			if err := tx.Create(seasonScores).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if animeData.Mappings.AniDB > 0 || animeData.Mappings.Anilist > 0 || animeData.Mappings.AnimeCountdown > 0 || animeData.Mappings.AnimePlanet != "" || animeData.Mappings.AniSearch > 0 || animeData.Mappings.IMDB != "" || animeData.Mappings.Kitsu > 0 || animeData.Mappings.LiveChart > 0 || animeData.Mappings.NotifyMoe != "" || animeData.Mappings.Simkl > 0 || animeData.Mappings.TMDB > 0 || animeData.Mappings.TVDB > 0 {
		mapping := &entities.AnimeMapping{
			AniDB:          animeData.Mappings.AniDB,
			Anilist:        animeData.Mappings.Anilist,
			AnimeCountdown: animeData.Mappings.AnimeCountdown,
			AnimePlanet:    animeData.Mappings.AnimePlanet,
			AniSearch:      animeData.Mappings.AniSearch,
			IMDB:           animeData.Mappings.IMDB,
			Kitsu:          animeData.Mappings.Kitsu,
			LiveChart:      animeData.Mappings.LiveChart,
			MAL:            animeData.MALID,
			NotifyMoe:      animeData.Mappings.NotifyMoe,
			Simkl:          animeData.Mappings.Simkl,
			TMDB:           animeData.Mappings.TMDB,
			TVDB:           animeData.Mappings.TVDB,
		}

		if err := tx.Create(mapping).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func ConvertToTypesAnime(cached *entities.CachedAnime) *types.Anime {
	if cached == nil {
		return nil
	}

	anime := &types.Anime{
		MALID: cached.MALID,
		Titles: types.AnimeTitles{
			Romaji:   cached.TitleRomaji,
			English:  cached.TitleEnglish,
			Japanese: cached.TitleJapanese,
			Synonyms: strings.Split(cached.TitleSynonyms, ","),
		},
		Synopsis: cached.Synopsis,
		Type:     types.AniSyncType(cached.Type),
		Source:   cached.Source,
		Airing:   cached.Airing,
		Status:   cached.Status,
		Duration: cached.Duration,
		Color:    cached.Color,
		Season:   cached.Season,
		Year:     cached.Year,
		Episodes: types.AnimeEpisodes{
			Total:  cached.TotalEpisodes,
			Aired:  cached.AiredEpisodes,
			Subbed: cached.SubbedCount,
			Dubbed: cached.DubbedCount,
		},
	}

	if cached.Images != nil {
		anime.Images = types.AnimeImages{
			Small:    cached.Images.Small,
			Large:    cached.Images.Large,
			Original: cached.Images.Original,
		}
	}

	if cached.Logos != nil {
		anime.Logos = types.AnimeLogos{
			Small:    cached.Logos.Small,
			Medium:   cached.Logos.Medium,
			Large:    cached.Logos.Large,
			XLarge:   cached.Logos.XLarge,
			Original: cached.Logos.Original,
		}
	}

	if cached.Covers != nil {
		anime.Covers = types.AnimeImages{
			Small:    cached.Covers.Small,
			Large:    cached.Covers.Large,
			Original: cached.Covers.Original,
		}
	}

	if cached.Scores != nil {
		anime.Scores = types.AnimeScores{
			Score:      cached.Scores.Score,
			ScoredBy:   cached.Scores.ScoredBy,
			Rank:       cached.Scores.Rank,
			Popularity: cached.Scores.Popularity,
			Members:    cached.Scores.Members,
			Favorites:  cached.Scores.Favorites,
		}
	}

	if cached.AiringStatus != nil {
		airingStatus := types.AiringStatus{
			String: cached.AiringStatus.String,
		}

		if cached.AiringStatus.From != nil {
			airingStatus.From = types.AiringStatusDates{
				Day:    cached.AiringStatus.From.Day,
				Month:  cached.AiringStatus.From.Month,
				Year:   cached.AiringStatus.From.Year,
				String: cached.AiringStatus.From.String,
			}
		}

		if cached.AiringStatus.To != nil {
			airingStatus.To = types.AiringStatusDates{
				Day:    cached.AiringStatus.To.Day,
				Month:  cached.AiringStatus.To.Month,
				Year:   cached.AiringStatus.To.Year,
				String: cached.AiringStatus.To.String,
			}
		}

		anime.AiringStatus = airingStatus
	}

	if cached.Broadcast != nil {
		anime.Broadcast = types.AnimeBroadcast{
			Day:      cached.Broadcast.Day,
			Time:     cached.Broadcast.Time,
			Timezone: cached.Broadcast.Timezone,
			String:   cached.Broadcast.String,
		}
	}

	if len(cached.Genres) > 0 {
		anime.Genres = make([]types.AnimeGenres, len(cached.Genres))
		for i, genre := range cached.Genres {
			anime.Genres[i] = types.AnimeGenres{
				Name:    genre.Name,
				GenreID: genre.GenreID,
				URL:     genre.URL,
			}
		}
	}

	if len(cached.Producers) > 0 {
		anime.Producers = make([]types.AnimeProducer, len(cached.Producers))
		for i, producer := range cached.Producers {
			anime.Producers[i] = types.AnimeProducer{
				Name:       producer.Name,
				ProducerID: producer.ProducerID,
				URL:        producer.URL,
			}
		}
	}

	if len(cached.Studios) > 0 {
		anime.Studios = make([]types.AnimeStudio, len(cached.Studios))
		for i, studio := range cached.Studios {
			anime.Studios[i] = types.AnimeStudio{
				Name:     studio.Name,
				StudioID: studio.StudioID,
				URL:      studio.URL,
			}
		}
	}

	if len(cached.Licensors) > 0 {
		anime.Licensors = make([]types.AnimeLicensor, len(cached.Licensors))
		for i, licensor := range cached.Licensors {
			anime.Licensors[i] = types.AnimeLicensor{
				Name:       licensor.Name,
				ProducerID: licensor.ProducerID,
				URL:        licensor.URL,
			}
		}
	}

	if cached.NextAiringEpisode != nil {
		anime.NextAiringEpisode = types.AnimeAiringEpisode{
			AiringAt: cached.NextAiringEpisode.AiringAt,
			Episode:  cached.NextAiringEpisode.Episode,
		}
	}

	if len(cached.AiringSchedule) > 0 {
		anime.AiringSchedule = make([]types.AnimeAiringEpisode, len(cached.AiringSchedule))
		for i, episode := range cached.AiringSchedule {
			anime.AiringSchedule[i] = types.AnimeAiringEpisode{
				AiringAt: episode.AiringAt,
				Episode:  episode.Episode,
			}
		}
	}

	if len(cached.Episodes) > 0 {
		anime.Episodes.Episodes = make([]types.AnimeSingleEpisode, len(cached.Episodes))
		for i, episode := range cached.Episodes {
			episodeData := types.AnimeSingleEpisode{
				ID:           episode.EpisodeID,
				Description:  episode.Description,
				Aired:        episode.Aired,
				Score:        episode.Score,
				Filler:       episode.Filler,
				Recap:        episode.Recap,
				ForumURL:     episode.ForumURL,
				URL:          episode.URL,
				ThumbnailURL: episode.ThumbnailURL,
			}

			if episode.Titles != nil {
				episodeData.Titles = types.EpisodeTitles{
					English:  episode.Titles.English,
					Japanese: episode.Titles.Japanese,
					Romaji:   episode.Titles.Romaji,
				}
			}

			anime.Episodes.Episodes[i] = episodeData
		}
	}

	if len(cached.Characters) > 0 {
		anime.Characters = make([]types.AnimeCharacter, len(cached.Characters))
		for i, character := range cached.Characters {
			animeCharacter := types.AnimeCharacter{
				MALID:    character.MALID,
				URL:      character.URL,
				ImageURL: character.ImageURL,
				Name:     character.Name,
				Role:     character.Role,
			}

			if len(character.VoiceActors) > 0 {
				animeCharacter.VoiceActors = make([]types.AnimeVoiceActor, len(character.VoiceActors))
				for j, va := range character.VoiceActors {
					animeCharacter.VoiceActors[j] = types.AnimeVoiceActor{
						MALID:    va.MALID,
						URL:      va.URL,
						Image:    va.Image,
						Name:     va.Name,
						Language: va.Language,
					}
				}
			}

			anime.Characters[i] = animeCharacter
		}
	}

	if len(cached.Seasons) > 0 {
		anime.Seasons = make([]types.AnimeSeason, len(cached.Seasons))
		for i, season := range cached.Seasons {
			animeSeason := types.AnimeSeason{
				MALID:    season.MALID,
				Synopsis: season.Synopsis,
				Type:     types.AniSyncType(season.Type),
				Source:   season.Source,
				Airing:   season.Airing,
				Status:   season.Status,
				Duration: season.Duration,
				Season:   season.Season,
				Year:     season.Year,
				Current:  season.Current,
				Titles: types.AnimeTitles{
					Romaji:   season.TitleRomaji,
					English:  season.TitleEnglish,
					Japanese: season.TitleJapanese,
					Synonyms: strings.Split(season.TitleSynonyms, ","),
				},
			}

			if season.Images != nil {
				animeSeason.Images = types.AnimeImages{
					Small:    season.Images.Small,
					Large:    season.Images.Large,
					Original: season.Images.Original,
				}
			}

			if season.Scores != nil {
				animeSeason.Scores = types.AnimeScores{
					Score:      season.Scores.Score,
					ScoredBy:   season.Scores.ScoredBy,
					Rank:       season.Scores.Rank,
					Popularity: season.Scores.Popularity,
					Members:    season.Scores.Members,
					Favorites:  season.Scores.Favorites,
				}
			}

			if season.AiringStatus != nil {
				airingStatus := types.AiringStatus{
					String: season.AiringStatus.String,
				}

				if season.AiringStatus.From != nil {
					airingStatus.From = types.AiringStatusDates{
						Day:    season.AiringStatus.From.Day,
						Month:  season.AiringStatus.From.Month,
						Year:   season.AiringStatus.From.Year,
						String: season.AiringStatus.From.String,
					}
				}

				if season.AiringStatus.To != nil {
					airingStatus.To = types.AiringStatusDates{
						Day:    season.AiringStatus.To.Day,
						Month:  season.AiringStatus.To.Month,
						Year:   season.AiringStatus.To.Year,
						String: season.AiringStatus.To.String,
					}
				}

				animeSeason.AiringStatus = airingStatus
			}

			anime.Seasons[i] = animeSeason
		}
	}

	var mapping entities.AnimeMapping
	if err := DB.Where("mal = ?", cached.MALID).First(&mapping).Error; err == nil {
		anime.Mappings = types.AnimeMappings{
			AniDB:          mapping.AniDB,
			Anilist:        mapping.Anilist,
			AnimeCountdown: mapping.AnimeCountdown,
			AnimePlanet:    mapping.AnimePlanet,
			AniSearch:      mapping.AniSearch,
			IMDB:           mapping.IMDB,
			Kitsu:          mapping.Kitsu,
			LiveChart:      mapping.LiveChart,
			NotifyMoe:      mapping.NotifyMoe,
			Simkl:          mapping.Simkl,
			TMDB:           mapping.TMDB,
			TVDB:           mapping.TVDB,
		}
	}

	return anime
}
