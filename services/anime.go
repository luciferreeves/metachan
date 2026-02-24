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
	"strings"
	"time"
)

func GetAnime(mapping *entities.Mapping) (*entities.Anime, error) {
	if mapping == nil {
		logger.Errorf("AnimeService", "Mapping is nil")
		return nil, fmt.Errorf("mapping is nil")
	}

	malID := mapping.MAL
	logger.Infof("AnimeService", "Fetching anime data for MAL ID: %d", malID)

	existingAnime, err := repositories.GetAnime(enums.MAL, malID)
	if err == nil {
		if time.Since(existingAnime.UpdatedAt) < 7*24*time.Hour {
			logger.Infof("AnimeService", "Returning cached anime (MAL ID: %d, age: %v)", malID, time.Since(existingAnime.UpdatedAt).Round(time.Second))
			return &existingAnime, nil
		}
		logger.Infof("AnimeService", "Cached anime is stale, refreshing (MAL ID: %d)", malID)
		return fetchAnime(mapping, &existingAnime)
	}

	logger.Infof("AnimeService", "Anime not found in database, creating new")
	return fetchAnime(mapping, nil)
}

func ForceRefreshAnime(mapping *entities.Mapping) (*entities.Anime, error) {
	if mapping == nil {
		logger.Errorf("AnimeService", "Mapping is nil")
		return nil, fmt.Errorf("mapping is nil")
	}

	logger.Infof("AnimeService", "Force refreshing anime data for MAL ID: %d", mapping.MAL)

	existingAnime, err := repositories.GetAnime(enums.MAL, mapping.MAL)
	if err == nil {
		return fetchAnime(mapping, &existingAnime)
	}
	return fetchAnime(mapping, nil)
}

func fetchAnime(mapping *entities.Mapping, existing *entities.Anime) (*entities.Anime, error) {
	malID := mapping.MAL

	var anime *entities.Anime
	if existing != nil {
		anime = existing
	} else {
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

	anime.Episodes = nil
	anime.Characters = nil
	anime.Genres = nil
	anime.Producers = nil
	anime.Studios = nil
	anime.Licensors = nil
	anime.Schedule = nil

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

	if mapping.TVDB > 0 || mapping.TMDB > 0 {
		logger.Infof("AnimeService", "Fetching related anime seasons")
		applySeasonData(anime, mapping)
	}

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
		producer := entities.Producer{
			MALID: jp.MALID,
			URL:   jp.URL,
		}
		if jp.Name != "" {
			producer.Titles = []entities.SimpleTitle{{Title: jp.Name, Type: "Default"}}
		}
		anime.Producers = append(anime.Producers, producer)
	}

	for _, js := range jikanAnime.Data.Studios {
		studio := entities.Producer{
			MALID: js.MALID,
			URL:   js.URL,
		}
		if js.Name != "" {
			studio.Titles = []entities.SimpleTitle{{Title: js.Name, Type: "Default"}}
		}
		anime.Studios = append(anime.Studios, studio)
	}

	for _, jl := range jikanAnime.Data.Licensors {
		licensor := entities.Producer{
			MALID: jl.MALID,
			URL:   jl.URL,
		}
		if jl.Name != "" {
			licensor.Titles = []entities.SimpleTitle{{Title: jl.Name, Type: "Default"}}
		}
		anime.Licensors = append(anime.Licensors, licensor)
	}

	anime.TotalEpisodes = jikanAnime.Data.Episodes
	anime.AiredEpisodes = len(jikanEpisodes.Data)

	for _, je := range jikanEpisodes.Data {
		episode := entities.Episode{
			EpisodeNumber: je.MALID,
			URL:           je.URL,
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
			char := entities.Character{
				MALID:    jc.Character.MALID,
				Name:     jc.Character.Name,
				URL:      jc.Character.URL,
				ImageURL: jc.Character.Images.JPG.ImageURL,
				Role:     jc.Role,
			}

			for _, va := range jc.VoiceActors {
				char.VoiceActors = append(char.VoiceActors, entities.CharacterVoiceActor{
					Language: va.Language,
					VoiceActor: &entities.VoiceActor{
						MALID: va.Person.MALID,
						Name:  va.Person.Name,
						URL:   va.Person.URL,
						Image: va.Person.Images.JPG.ImageURL,
					},
				})
			}

			anime.Characters = append(anime.Characters, char)
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

	logos := extractLogosFromMALSync(malSyncData)
	if logos != nil {
		anime.Logos = logos
	}
}

func extractLogosFromMALSync(malSyncData *types.MalsyncAnimeResponse) *entities.Logos {
	if malSyncData == nil {
		return nil
	}

	crunchyrollSites, exists := malSyncData.Sites["Crunchyroll"]
	if !exists || len(crunchyrollSites) == 0 {
		logger.Debugf("AnimeService", "No Crunchyroll data found in MALSync response")
		return nil
	}

	crURL := ""
	for _, site := range crunchyrollSites {
		crURL = site.URL
		break
	}

	if crURL == "" {
		logger.Debugf("AnimeService", "No valid Crunchyroll URL found")
		return nil
	}

	seriesID := extractCrunchyrollSeriesID(crURL)
	if seriesID == "" {
		return nil
	}

	logoSizes := map[string]int{
		"Small":    320,
		"Medium":   480,
		"Large":    600,
		"XLarge":   800,
		"Original": 1200,
	}

	logos := &entities.Logos{
		Small:    fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Small"], seriesID),
		Medium:   fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Medium"], seriesID),
		Large:    fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Large"], seriesID),
		XLarge:   fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["XLarge"], seriesID),
		Original: fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Original"], seriesID),
	}

	return logos
}

func extractCrunchyrollSeriesID(crURL string) string {
	if crURL == "" {
		return ""
	}

	parts := strings.Split(crURL, "/")
	for i, part := range parts {
		if part == "series" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	logger.Debugf("AnimeService", "Could not extract series ID from URL: %s", crURL)
	return ""
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

func applySeasonData(anime *entities.Anime, mapping *entities.Mapping) {
	var relatedMappings []entities.Mapping
	malIDSet := make(map[int]bool)

	if mapping.TVDB > 0 {
		tvdbMappings, err := repositories.GetRelatedAnimeByTVDB(mapping.TVDB, mapping.MAL)
		if err == nil && len(tvdbMappings) > 0 {
			logger.Infof("AnimeService", "Found %d related anime via TVDB", len(tvdbMappings))
			for _, m := range tvdbMappings {
				if !malIDSet[m.MAL] {
					malIDSet[m.MAL] = true
					relatedMappings = append(relatedMappings, m)
				}
			}
		}
	}

	if mapping.TMDB > 0 {
		tmdbMappings, err := repositories.GetRelatedAnimeByTMDB(mapping.TMDB, mapping.MAL)
		if err == nil && len(tmdbMappings) > 0 {
			logger.Infof("AnimeService", "Found %d related anime via TMDB", len(tmdbMappings))
			for _, m := range tmdbMappings {
				if !malIDSet[m.MAL] {
					malIDSet[m.MAL] = true
					relatedMappings = append(relatedMappings, m)
				}
			}
		}
	}

	if len(relatedMappings) == 0 {
		logger.Debugf("AnimeService", "No related anime seasons found")
		anime.SeasonNumber = 1
		return
	}

	logger.Infof("AnimeService", "Fetching details for %d season(s)", len(relatedMappings))

	allSeasons := []seasonInfo{{
		malID:       anime.MALID,
		year:        anime.Year,
		seasonOrder: getSeasonOrder(anime.Season),
		isCurrent:   true,
	}}

	for _, relatedMapping := range relatedMappings {
		seasonAnime, err := jikan.GetAnimeByMALID(relatedMapping.MAL)
		if err != nil {
			logger.Warnf("AnimeService", "Failed to fetch season data for MAL ID %d: %v", relatedMapping.MAL, err)
			continue
		}

		allSeasons = append(allSeasons, seasonInfo{
			malID:       relatedMapping.MAL,
			year:        seasonAnime.Data.Year,
			seasonOrder: getSeasonOrder(seasonAnime.Data.Season),
			isCurrent:   false,
		})

		season := entities.Season{
			MALID:    relatedMapping.MAL,
			Synopsis: seasonAnime.Data.Synopsis,
			Type:     seasonAnime.Data.Type,
			Source:   seasonAnime.Data.Source,
			Airing:   seasonAnime.Data.Airing,
			Status:   seasonAnime.Data.Status,
			Duration: seasonAnime.Data.Duration,
			Season:   seasonAnime.Data.Season,
			Year:     seasonAnime.Data.Year,
		}

		if seasonAnime.Data.Title != "" || seasonAnime.Data.TitleEnglish != "" || seasonAnime.Data.TitleJapanese != "" {
			season.Title = &entities.Title{
				Romaji:   seasonAnime.Data.Title,
				English:  seasonAnime.Data.TitleEnglish,
				Japanese: seasonAnime.Data.TitleJapanese,
				Synonyms: seasonAnime.Data.TitleSynonyms,
			}
		}

		if seasonAnime.Data.Images.JPG.ImageURL != "" {
			season.Images = &entities.Images{
				Small:    seasonAnime.Data.Images.JPG.SmallImageURL,
				Large:    seasonAnime.Data.Images.JPG.LargeImageURL,
				Original: seasonAnime.Data.Images.JPG.ImageURL,
			}
		}

		season.Scores = &entities.Scores{
			Score:      seasonAnime.Data.Score,
			ScoredBy:   seasonAnime.Data.ScoredBy,
			Rank:       seasonAnime.Data.Rank,
			Popularity: seasonAnime.Data.Popularity,
			Members:    seasonAnime.Data.Members,
			Favorites:  seasonAnime.Data.Favorites,
		}

		if seasonAnime.Data.Aired.From != "" || seasonAnime.Data.Aired.To != "" {
			season.AiringStatus = &entities.AiringStatus{
				String: seasonAnime.Data.Aired.String,
			}
			if seasonAnime.Data.Aired.Prop.From.Year > 0 {
				season.AiringStatus.From = &entities.Date{
					Day:    seasonAnime.Data.Aired.Prop.From.Day,
					Month:  seasonAnime.Data.Aired.Prop.From.Month,
					Year:   seasonAnime.Data.Aired.Prop.From.Year,
					String: seasonAnime.Data.Aired.From,
				}
			}
			if seasonAnime.Data.Aired.Prop.To.Year > 0 {
				season.AiringStatus.To = &entities.Date{
					Day:    seasonAnime.Data.Aired.Prop.To.Day,
					Month:  seasonAnime.Data.Aired.Prop.To.Month,
					Year:   seasonAnime.Data.Aired.Prop.To.Year,
					String: seasonAnime.Data.Aired.To,
				}
			}
		}

		anime.Seasons = append(anime.Seasons, season)
	}

	sortSeasonsByChronology(allSeasons)

	seasonNumberMap := make(map[int]int)
	for i, s := range allSeasons {
		seasonNumberMap[s.malID] = i + 1
		if s.isCurrent {
			anime.SeasonNumber = i + 1
		}
	}

	for i := range anime.Seasons {
		anime.Seasons[i].SeasonNumber = seasonNumberMap[anime.Seasons[i].MALID]
	}

	logger.Successf("AnimeService", "Successfully fetched %d season(s), current anime is season %d", len(anime.Seasons), anime.SeasonNumber)
}

func getSeasonOrder(season string) int {
	switch strings.ToLower(season) {
	case "winter":
		return 1
	case "spring":
		return 2
	case "summer":
		return 3
	case "fall", "autumn":
		return 4
	default:
		return 0
	}
}

func sortSeasonsByChronology(seasons []seasonInfo) {
	for i := 0; i < len(seasons)-1; i++ {
		for j := 0; j < len(seasons)-i-1; j++ {
			s1, s2 := seasons[j], seasons[j+1]

			if s1.year > s2.year {
				seasons[j], seasons[j+1] = seasons[j+1], seasons[j]
			} else if s1.year == s2.year && s1.seasonOrder > s2.seasonOrder {
				seasons[j], seasons[j+1] = seasons[j+1], seasons[j]
			}
		}
	}
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

	if len(anime.Episodes) > 0 {
		if err := repositories.SaveAnimeEpisodes(anime.ID, anime.Episodes); err != nil {
			logger.Warnf("AnimeService", "Failed to save episodes: %v", err)
		}
	}

	if len(anime.Characters) > 0 {
		if err := repositories.SaveAnimeCharacters(anime.ID, anime.Characters); err != nil {
			logger.Warnf("AnimeService", "Failed to save characters: %v", err)
		}
	}

	for episodeID, skipTimes := range skipTimeMap {
		if err := repositories.SaveEpisodeSkipTimes(episodeID, skipTimes); err != nil {
			logger.Warnf("AnimeService", "Failed to save skip times for episode %s: %v", episodeID, err)
		}
	}

	logger.Successf("AnimeService", "Saved anime with %d episodes, %d characters, %d skip time entries", len(anime.Episodes), len(anime.Characters), len(skipTimeMap))
	return nil
}
