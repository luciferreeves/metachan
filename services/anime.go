package services

import (
	"fmt"
	"metachan/entities"
	"metachan/enums"
	"metachan/repositories"
	"metachan/types"
	"metachan/utils/api/anilist"
	"metachan/utils/api/aniskip"
	"metachan/utils/api/jikan"
	"metachan/utils/api/malsync"
	"metachan/utils/api/streaming"
	"metachan/utils/api/tmdb"
	"metachan/utils/api/tvdb"
	"metachan/utils/logger"
)

func GetAnime(mapping *entities.Mapping) (*entities.Anime, error) {
	if mapping == nil {
		logger.Errorf("AnimeService", "Mapping is nil")
		return nil, fmt.Errorf("mapping is nil")
	}

	malID := mapping.MAL
	logger.Infof("AnimeService", "Fetching anime data for MAL ID: %d", malID)

	var anime *entities.Anime
	existingAnime, err := repositories.GetAnime(enums.MAL, malID)
	if err == nil {
		logger.Infof("AnimeService", "Found existing anime in database, will update with fresh data")
		anime = &existingAnime
	} else {
		logger.Infof("AnimeService", "Anime not found in database, creating new")
		anime = &entities.Anime{
			MALID:   malID,
			Mapping: mapping,
		}
	}

	jikanAnime, err := jikan.GetAnimeByMALID(malID)
	if err != nil {
		logger.Errorf("AnimeService", "Failed to fetch anime from Jikan: %v", err)
		return nil, fmt.Errorf("failed to fetch anime from Jikan: %w", err)
	}

	jikanEpisodes, err := jikan.GetAnimeEpisodesByMALID(malID)
	if err != nil {
		logger.Errorf("AnimeService", "Failed to fetch episodes from Jikan: %v", err)
		return nil, fmt.Errorf("failed to fetch episodes from Jikan: %w", err)
	}

	jikanCharacters, err := jikan.GetAnimeCharactersByMALID(malID)
	if err != nil {
		logger.Warnf("AnimeService", "Failed to fetch characters from Jikan: %v", err)
	}

	applyJikanData(anime, jikanAnime, jikanEpisodes, jikanCharacters)

	if mapping.Anilist > 0 {
		anilistData, err := anilist.GetAnimeByAnilistID(mapping.Anilist)
		if err != nil {
			logger.Warnf("AnimeService", "Failed to fetch Anilist data: %v", err)
		} else {
			applyAnilistData(anime, anilistData)
		}
	}

	malSyncData, err := malsync.GetAnimeByMALID(malID)
	if err != nil {
		logger.Warnf("AnimeService", "Failed to fetch MALsync data: %v", err)
	} else {
		applyMALsyncData(anime, malSyncData)
	}

	animeType := string(mapping.Type)
	if (animeType == "MOVIE" || animeType == "Movie") && mapping.TMDB > 0 {
		logger.Infof("AnimeService", "Enriching movie episode from TMDB")
		if err := tmdb.EnrichEpisodeFromMovie(anime); err != nil {
			logger.Warnf("AnimeService", "Failed to enrich movie from TMDB: %v", err)
		}
	} else {
		if mapping.TVDB > 0 {
			logger.Infof("AnimeService", "Enriching episodes from TVDB")
			tvdbEpisodes, err := tvdb.GetSeriesEpisodes(mapping.TVDB)
			if err == nil && len(tvdbEpisodes) > 0 {
				tvdb.EnrichEpisodesFromTVDB(anime, tvdbEpisodes)
				logger.Successf("AnimeService", "Successfully enriched %d episodes from TVDB", len(tvdbEpisodes))
			} else {
				logger.Warnf("AnimeService", "Failed to fetch TVDB episodes: %v, falling back to TMDB", err)
				applyTMDBData(anime)
			}
		} else {
			applyTMDBData(anime)
		}
	}

	epSkipMap := make(map[string][]entities.EpisodeSkipTime)
	if mapping.Anilist > 0 {
		logger.Infof("AnimeService", "Enriching episodes with Aniskip data")
		for i := range anime.Episodes {
			episode := &anime.Episodes[i]
			skipData, err := aniskip.GetSkipTimesForEpisode(malID, episode.EpisodeNumber)
			if err != nil {
				continue
			}
			skipTimes := applyAniskipData(episode, skipData)
			if len(skipTimes) > 0 {
				epSkipMap[episode.EpisodeID] = skipTimes
			}
		}
	}

	applyStreamingData(anime)

	if err := saveAnime(anime, epSkipMap); err != nil {
		logger.Errorf("AnimeService", "Failed to save anime to database: %v", err)
		return nil, fmt.Errorf("failed to save anime to database: %w", err)
	}

	logger.Successf("AnimeService", "Successfully fetched and saved anime (MAL ID: %d)", malID)
	return anime, nil
}

func applyTMDBData(anime *entities.Anime) {
	if anime.Mapping != nil && anime.Mapping.TMDB > 0 {
		logger.Infof("AnimeService", "Enriching episodes from TMDB")
		if err := tmdb.AttachEpisodeDescriptions(anime); err != nil {
			logger.Warnf("AnimeService", "Failed to enrich episodes from TMDB: %v", err)
		} else {
			logger.Successf("AnimeService", "Successfully enriched episodes from TMDB")
		}
	}
}

func applyJikanData(anime *entities.Anime, jikanAnime *types.JikanAnimeResponse, jikanEpisodes *types.JikanAnimeEpisodeResponse, jikanCharacters *types.JikanAnimeCharacterResponse) {
	anime.Synopsis = jikanAnime.Data.Synopsis
	anime.Type = jikanAnime.Data.Type
	anime.Source = jikanAnime.Data.Source
	anime.Airing = jikanAnime.Data.Airing
	anime.Status = jikanAnime.Data.Status
	anime.Duration = jikanAnime.Data.Duration
	anime.Season = jikanAnime.Data.Season
	anime.Year = jikanAnime.Data.Year

	if jikanAnime.Data.Title != "" || jikanAnime.Data.TitleEnglish != "" || jikanAnime.Data.TitleJapanese != "" {
		anime.Title = &entities.Title{
			Romaji:   jikanAnime.Data.Title,
			English:  jikanAnime.Data.TitleEnglish,
			Japanese: jikanAnime.Data.TitleJapanese,
			Synonyms: jikanAnime.Data.TitleSynonyms,
		}
	}

	anime.Scores = &entities.Scores{
		Score:      jikanAnime.Data.Score,
		ScoredBy:   jikanAnime.Data.ScoredBy,
		Rank:       jikanAnime.Data.Rank,
		Popularity: jikanAnime.Data.Popularity,
		Members:    jikanAnime.Data.Members,
		Favorites:  jikanAnime.Data.Favorites,
	}

	if jikanAnime.Data.Images.JPG.ImageURL != "" {
		anime.Images = &entities.Images{
			Small:    jikanAnime.Data.Images.JPG.SmallImageURL,
			Large:    jikanAnime.Data.Images.JPG.LargeImageURL,
			Original: jikanAnime.Data.Images.JPG.ImageURL,
		}
	}

	if jikanAnime.Data.Aired.From != "" || jikanAnime.Data.Aired.To != "" {
		anime.AiringStatus = &entities.AiringStatus{
			String: jikanAnime.Data.Aired.String,
		}
		if jikanAnime.Data.Aired.Prop.From.Year > 0 {
			anime.AiringStatus.From = &entities.Date{
				Day:    jikanAnime.Data.Aired.Prop.From.Day,
				Month:  jikanAnime.Data.Aired.Prop.From.Month,
				Year:   jikanAnime.Data.Aired.Prop.From.Year,
				String: jikanAnime.Data.Aired.From,
			}
		}
		if jikanAnime.Data.Aired.Prop.To.Year > 0 {
			anime.AiringStatus.To = &entities.Date{
				Day:    jikanAnime.Data.Aired.Prop.To.Day,
				Month:  jikanAnime.Data.Aired.Prop.To.Month,
				Year:   jikanAnime.Data.Aired.Prop.To.Year,
				String: jikanAnime.Data.Aired.To,
			}
		}
	}

	if jikanAnime.Data.Broadcast.Day != "" {
		anime.Broadcast = &entities.Broadcast{
			Day:      jikanAnime.Data.Broadcast.Day,
			Time:     jikanAnime.Data.Broadcast.Time,
			Timezone: jikanAnime.Data.Broadcast.Timezone,
			String:   jikanAnime.Data.Broadcast.String,
		}
	}

	for _, jg := range jikanAnime.Data.Genres {
		anime.Genres = append(anime.Genres, entities.Genre{
			GenreID: jg.MALID,
			Name:    jg.Name,
			URL:     jg.URL,
		})
	}

	for _, jg := range jikanAnime.Data.ExplicitGenres {
		anime.Genres = append(anime.Genres, entities.Genre{
			GenreID: jg.MALID,
			Name:    jg.Name,
			URL:     jg.URL,
		})
	}

	for _, jp := range jikanAnime.Data.Producers {
		anime.Producers = append(anime.Producers, entities.Producer{
			MALID: jp.MALID,
			URL:   jp.URL,
		})
	}

	for _, js := range jikanAnime.Data.Studios {
		anime.Studios = append(anime.Studios, entities.Producer{
			MALID: js.MALID,
			URL:   js.URL,
		})
	}

	for _, jl := range jikanAnime.Data.Licensors {
		anime.Licensors = append(anime.Licensors, entities.Producer{
			MALID: jl.MALID,
			URL:   jl.URL,
		})
	}

	anime.TotalEpisodes = jikanAnime.Data.Episodes
	anime.AiredEpisodes = len(jikanEpisodes.Data)

	for _, je := range jikanEpisodes.Data {
		episode := entities.Episode{
			EpisodeNumber: je.MALID,
			Aired:         je.Aired,
			Score:         je.Score,
			Filler:        je.Filler,
			Recap:         je.Recap,
			ForumURL:      je.ForumURL,
		}

		if je.Title != "" || je.TitleJapanese != "" || je.TitleRomaji != "" {
			episode.Title = &entities.Title{
				English:  je.Title,
				Japanese: je.TitleJapanese,
				Romaji:   je.TitleRomaji,
			}
		}

		anime.Episodes = append(anime.Episodes, episode)
	}

	if jikanCharacters != nil {
		for _, jc := range jikanCharacters.Data {
			character := entities.Character{
				MALID:    jc.MALID,
				Name:     jc.Name,
				Role:     jc.Role,
				URL:      jc.URL,
				ImageURL: jc.Images.JPG.ImageURL,
			}

			if len(jc.VoiceActors) > 0 {
				for _, va := range jc.VoiceActors {
					if va.Language == "Japanese" {
						character.VoiceActors = append(character.VoiceActors, entities.VoiceActor{
							MALID:    va.MALID,
							Name:     va.Name,
							Language: va.Language,
							URL:      va.URL,
							Image:    va.Images.JPG.ImageURL,
						})
					}
				}
			}

			anime.Characters = append(anime.Characters, character)
		}
	}
}

func applyAnilistData(anime *entities.Anime, anilistData *types.AnilistAnimeResponse) {
	if anilistData == nil || anilistData.Data.Media.ID == 0 {
		return
	}

	media := anilistData.Data.Media

	if anime.Color == "" && media.CoverImage.Color != "" {
		anime.Color = media.CoverImage.Color
	}

	if anime.Covers == nil && (media.CoverImage.Medium != "" || media.CoverImage.Large != "" || media.CoverImage.ExtraLarge != "") {
		anime.Covers = &entities.Images{
			Small:    media.CoverImage.Medium,
			Large:    media.CoverImage.Large,
			Original: media.CoverImage.ExtraLarge,
		}
	}

	if media.NextAiringEpisode.AiringAt > 0 {
		anime.NextAiring = &entities.NextEpisode{
			Episode:  media.NextAiringEpisode.Episode,
			AiringAt: media.NextAiringEpisode.AiringAt,
		}
	}

	for _, ep := range media.AiringSchedule.Nodes {
		anime.Schedule = append(anime.Schedule, entities.EpisodeSchedule{
			Episode:  ep.Episode,
			AiringAt: ep.AiringAt,
		})
	}
}

func applyMALsyncData(anime *entities.Anime, malSyncData *types.MalsyncAnimeResponse) {
	if malSyncData == nil {
		return
	}

	if anime.Logos == nil {
		anime.Logos = &entities.Logos{}
	}

	for _, site := range malSyncData.Sites {
		for _, entry := range site {
			if entry.Image != "" {
				anime.Logos.Original = entry.Image
				break
			}
		}
		if anime.Logos.Original != "" {
			break
		}
	}
}

func applyAniskipData(episode *entities.Episode, skipData []types.AniskipResult) []entities.EpisodeSkipTime {
	if len(skipData) == 0 {
		return nil
	}

	skipTimes := make([]entities.EpisodeSkipTime, 0, len(skipData))
	for _, result := range skipData {
		if result.EpisodeLength > 0 && episode.EpisodeLength != result.EpisodeLength {
			episode.EpisodeLength = result.EpisodeLength
		}

		skipTimes = append(skipTimes, entities.EpisodeSkipTime{
			SkipType:  result.SkipType,
			StartTime: result.Interval.StartTime,
			EndTime:   result.Interval.EndTime,
		})
	}

	return skipTimes
}

func applyStreamingData(anime *entities.Anime) {
	if anime.Title == nil {
		return
	}

	searchTitle := anime.Title.Romaji
	if searchTitle == "" {
		searchTitle = anime.Title.English
	}
	if searchTitle == "" {
		return
	}

	logger.Infof("AnimeService", "Fetching streaming counts for: %s", searchTitle)
	subCount, dubCount, err := streaming.GetStreamingCounts(searchTitle)
	if err != nil {
		if anime.Title.English != "" && anime.Title.English != searchTitle {
			subCount, dubCount, err = streaming.GetStreamingCounts(anime.Title.English)
		}
	}

	if err != nil {
		logger.Warnf("AnimeService", "Failed to fetch streaming counts: %v", err)
		return
	}

	anime.SubbedCount = subCount
	anime.DubbedCount = dubCount
	logger.Infof("AnimeService", "Streaming counts - Subbed: %d, Dubbed: %d", subCount, dubCount)
}

func saveAnime(anime *entities.Anime, skipTimeMap map[string][]entities.EpisodeSkipTime) error {
	if anime.Mapping != nil {
		if err := repositories.CreateOrUpdateMapping(anime.Mapping); err != nil {
			return fmt.Errorf("failed to save mapping: %w", err)
		}
		anime.MappingID = anime.Mapping.ID
	}

	for i := range anime.Genres {
		if err := repositories.CreateOrUpdateGenre(&anime.Genres[i]); err != nil {
			logger.Warnf("AnimeService", "Failed to save genre: %v", err)
		}
	}

	for i := range anime.Producers {
		if err := repositories.CreateOrUpdateProducer(&anime.Producers[i]); err != nil {
			logger.Warnf("AnimeService", "Failed to save producer: %v", err)
		}
	}

	for i := range anime.Studios {
		if err := repositories.CreateOrUpdateProducer(&anime.Studios[i]); err != nil {
			logger.Warnf("AnimeService", "Failed to save studio: %v", err)
		}
	}

	for i := range anime.Licensors {
		if err := repositories.CreateOrUpdateProducer(&anime.Licensors[i]); err != nil {
			logger.Warnf("AnimeService", "Failed to save licensor: %v", err)
		}
	}

	if err := repositories.CreateOrUpdateAnime(anime); err != nil {
		return fmt.Errorf("failed to save anime: %w", err)
	}

	for episodeID, skipTimes := range skipTimeMap {
		if err := repositories.SaveEpisodeSkipTimes(episodeID, skipTimes); err != nil {
			logger.Warnf("AnimeService", "Failed to save skip times for episode %s: %v", episodeID, err)
		}
	}

	logger.Successf("AnimeService", "Saved anime with %d episodes, %d characters, %d skip time entries", len(anime.Episodes), len(anime.Characters), len(skipTimeMap))
	return nil
}
