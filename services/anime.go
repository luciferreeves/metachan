package services

import (
	"context"
	"crypto/md5"
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

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

var flightGroup singleflight.Group

func GetAnime(mapping *entities.Mapping) (*entities.Anime, error) {
	if mapping == nil {
		logger.Errorf("AnimeService", "Mapping is nil")
		return nil, fmt.Errorf("mapping is nil")
	}

	key := fmt.Sprintf("anime:%d", mapping.MAL)
	result, err, _ := flightGroup.Do(key, func() (interface{}, error) {
		return getAnimeInternal(mapping)
	})

	if err != nil {
		return nil, err
	}
	return result.(*entities.Anime), nil
}

func getAnimeInternal(mapping *entities.Mapping) (*entities.Anime, error) {
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

	var jikanAnime *types.JikanAnimeResponse
	var jikanEpisodes *types.JikanAnimeEpisodeResponse
	var jikanCharacters *types.JikanAnimeCharacterResponse
	var anilistData *types.AnilistAnimeResponse
	var malSyncData *types.MalsyncAnimeResponse

	fetchGroup, _ := errgroup.WithContext(context.Background())

	fetchGroup.Go(func() error {
		var err error
		jikanAnime, err = jikan.GetAnimeByMALID(malID)
		if err != nil {
			return fmt.Errorf("failed to fetch anime from Jikan: %w", err)
		}
		return nil
	})

	fetchGroup.Go(func() error {
		var err error
		jikanEpisodes, err = jikan.GetAnimeEpisodesByMALID(malID)
		if err != nil {
			return fmt.Errorf("failed to fetch episodes from Jikan: %w", err)
		}
		return nil
	})

	fetchGroup.Go(func() error {
		var err error
		jikanCharacters, err = jikan.GetAnimeCharactersByMALID(malID)
		if err != nil {
			logger.Warnf("AnimeService", "Failed to fetch characters from Jikan: %v", err)
		}
		return nil
	})

	if mapping.Anilist > 0 {
		fetchGroup.Go(func() error {
			var err error
			anilistData, err = anilist.GetAnimeByAnilistID(mapping.Anilist)
			if err != nil {
				logger.Warnf("AnimeService", "Failed to fetch Anilist data: %v", err)
			}
			return nil
		})
	}

	fetchGroup.Go(func() error {
		var err error
		malSyncData, err = malsync.GetAnimeByMALID(malID)
		if err != nil {
			logger.Warnf("AnimeService", "Failed to fetch MALsync data: %v", err)
		}
		return nil
	})

	if err := fetchGroup.Wait(); err != nil {
		logger.Errorf("AnimeService", "Failed to fetch anime data: %v", err)
		return nil, err
	}

	anime.Episodes = nil
	anime.Characters = nil
	anime.Genres = nil
	anime.Themes = nil
	anime.Demographics = nil
	anime.Producers = nil
	anime.Studios = nil
	anime.Licensors = nil
	anime.Schedule = nil

	applyJikanData(anime, jikanAnime, jikanEpisodes, jikanCharacters)

	if anilistData != nil {
		applyAnilistData(anime, anilistData)
	}

	if malSyncData != nil {
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
	anime.Rating = jikanAnime.Data.Rating
	anime.Background = jikanAnime.Data.Background

	anime.Title = entities.AnimeTitle{
		Romaji:   jikanAnime.Data.Title,
		English:  jikanAnime.Data.TitleEnglish,
		Japanese: jikanAnime.Data.TitleJapanese,
		Synonyms: jikanAnime.Data.TitleSynonyms,
	}

	anime.Scores = entities.AnimeScores{
		Score:      jikanAnime.Data.Score,
		ScoredBy:   jikanAnime.Data.ScoredBy,
		Rank:       jikanAnime.Data.Rank,
		Popularity: jikanAnime.Data.Popularity,
		Members:    jikanAnime.Data.Members,
		Favorites:  jikanAnime.Data.Favorites,
	}

	anime.Images = entities.AnimeImages{
		Small:    jikanAnime.Data.Images.JPG.SmallImageURL,
		Large:    jikanAnime.Data.Images.JPG.LargeImageURL,
		Original: jikanAnime.Data.Images.JPG.ImageURL,
	}

	anime.Aired = entities.AnimeAired{
		String: jikanAnime.Data.Aired.String,
	}
	if jikanAnime.Data.Aired.From != "" {
		if parsedTime, err := time.Parse(time.RFC3339, jikanAnime.Data.Aired.From); err == nil {
			anime.Aired.From = &parsedTime
		}
	}
	if jikanAnime.Data.Aired.To != "" {
		if parsedTime, err := time.Parse(time.RFC3339, jikanAnime.Data.Aired.To); err == nil {
			anime.Aired.To = &parsedTime
		}
	}

	anime.Broadcast = entities.AnimeBroadcast{
		Day:      jikanAnime.Data.Broadcast.Day,
		Time:     jikanAnime.Data.Broadcast.Time,
		Timezone: jikanAnime.Data.Broadcast.Timezone,
		String:   jikanAnime.Data.Broadcast.String,
	}

	anime.Trailer = entities.AnimeTrailer{
		YoutubeID: jikanAnime.Data.Trailer.YoutubeID,
		URL:       jikanAnime.Data.Trailer.URL,
		EmbedURL:  jikanAnime.Data.Trailer.EmbedURL,
	}

	for _, genreEntry := range jikanAnime.Data.Genres {
		anime.Genres = append(anime.Genres, entities.Genre{
			GenreID: genreEntry.MALID,
			Name:    genreEntry.Name,
			URL:     genreEntry.URL,
		})
	}

	for _, genreEntry := range jikanAnime.Data.ExplicitGenres {
		anime.Genres = append(anime.Genres, entities.Genre{
			GenreID: genreEntry.MALID,
			Name:    genreEntry.Name,
			URL:     genreEntry.URL,
		})
	}

	for _, themeEntry := range jikanAnime.Data.Themes {
		anime.Themes = append(anime.Themes, entities.Genre{
			GenreID: themeEntry.MALID,
			Name:    themeEntry.Name,
			URL:     themeEntry.URL,
		})
	}

	for _, demographicEntry := range jikanAnime.Data.Demographics {
		anime.Demographics = append(anime.Demographics, entities.Genre{
			GenreID: demographicEntry.MALID,
			Name:    demographicEntry.Name,
			URL:     demographicEntry.URL,
		})
	}

	for _, producerEntry := range jikanAnime.Data.Producers {
		producer := entities.Producer{
			MALID: producerEntry.MALID,
			URL:   producerEntry.URL,
		}
		if producerEntry.Name != "" {
			producer.Titles = []entities.SimpleTitle{{Title: producerEntry.Name, Type: "Default"}}
		}
		anime.Producers = append(anime.Producers, producer)
	}

	for _, studioEntry := range jikanAnime.Data.Studios {
		studio := entities.Producer{
			MALID: studioEntry.MALID,
			URL:   studioEntry.URL,
		}
		if studioEntry.Name != "" {
			studio.Titles = []entities.SimpleTitle{{Title: studioEntry.Name, Type: "Default"}}
		}
		anime.Studios = append(anime.Studios, studio)
	}

	for _, licensorEntry := range jikanAnime.Data.Licensors {
		licensor := entities.Producer{
			MALID: licensorEntry.MALID,
			URL:   licensorEntry.URL,
		}
		if licensorEntry.Name != "" {
			licensor.Titles = []entities.SimpleTitle{{Title: licensorEntry.Name, Type: "Default"}}
		}
		anime.Licensors = append(anime.Licensors, licensor)
	}

	anime.TotalEpisodes = jikanAnime.Data.Episodes
	anime.AiredEpisodes = len(jikanEpisodes.Data)

	for _, jikanEpisode := range jikanEpisodes.Data {
		episode := entities.Episode{
			EpisodeNumber: jikanEpisode.MALID,
			URL:           jikanEpisode.URL,
			Aired:         jikanEpisode.Aired,
			Score:         jikanEpisode.Score,
			Filler:        jikanEpisode.Filler,
			Recap:         jikanEpisode.Recap,
			ForumURL:      jikanEpisode.ForumURL,
			Title: entities.EpisodeTitle{
				English:  jikanEpisode.Title,
				Japanese: jikanEpisode.TitleJapanese,
				Romaji:   jikanEpisode.TitleRomaji,
			},
		}

		titleForID := jikanEpisode.Title
		if titleForID == "" {
			titleForID = jikanEpisode.TitleRomaji
		}
		episode.EpisodeID = generateEpisodeID(anime.MALID, jikanEpisode.MALID, titleForID)

		anime.Episodes = append(anime.Episodes, episode)
	}

	if jikanCharacters != nil {
		for _, jikanCharacter := range jikanCharacters.Data {
			character := entities.Character{
				MALID:    jikanCharacter.Character.MALID,
				Name:     jikanCharacter.Character.Name,
				URL:      jikanCharacter.Character.URL,
				ImageURL: jikanCharacter.Character.Images.JPG.ImageURL,
				Role:     jikanCharacter.Role,
			}

			for _, voiceActor := range jikanCharacter.VoiceActors {
				character.VoiceActors = append(character.VoiceActors, entities.CharacterVoiceActor{
					Language: voiceActor.Language,
					Person: &entities.Person{
						Image: voiceActor.Person.Images.JPG.ImageURL,
					},
				})
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

	if anime.Covers.Original == "" && (media.CoverImage.Medium != "" || media.CoverImage.Large != "" || media.CoverImage.ExtraLarge != "") {
		anime.Covers = entities.AnimeImages{
			Small:    media.CoverImage.Medium,
			Large:    media.CoverImage.Large,
			Original: media.CoverImage.ExtraLarge,
		}
	}

	if media.NextAiringEpisode.AiringAt > 0 {
		anime.NextAiringAt = media.NextAiringEpisode.AiringAt
		anime.NextAiringEpisode = media.NextAiringEpisode.Episode
	}

	for _, scheduleNode := range media.AiringSchedule.Nodes {
		anime.Schedule = append(anime.Schedule, entities.EpisodeSchedule{
			Episode:  scheduleNode.Episode,
			AiringAt: scheduleNode.AiringAt,
		})
	}
}

func applyMALsyncData(anime *entities.Anime, malSyncData *types.MalsyncAnimeResponse) {
	if malSyncData == nil {
		return
	}

	logos := extractLogosFromMALSync(malSyncData)
	if logos.Small != "" {
		anime.Logos = logos
	}
}

func extractLogosFromMALSync(malSyncData *types.MalsyncAnimeResponse) entities.AnimeLogos {
	if malSyncData == nil {
		return entities.AnimeLogos{}
	}

	crunchyrollSites, exists := malSyncData.Sites["Crunchyroll"]
	if !exists || len(crunchyrollSites) == 0 {
		logger.Debugf("AnimeService", "No Crunchyroll data found in MALSync response")
		return entities.AnimeLogos{}
	}

	crURL := ""
	for _, site := range crunchyrollSites {
		crURL = site.URL
		break
	}

	if crURL == "" {
		logger.Debugf("AnimeService", "No valid Crunchyroll URL found")
		return entities.AnimeLogos{}
	}

	seriesID := extractCrunchyrollSeriesID(crURL)
	if seriesID == "" {
		return entities.AnimeLogos{}
	}

	logoSizes := map[string]int{
		"Small":    320,
		"Medium":   480,
		"Large":    600,
		"XLarge":   800,
		"Original": 1200,
	}

	return entities.AnimeLogos{
		Small:    fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Small"], seriesID),
		Medium:   fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Medium"], seriesID),
		Large:    fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Large"], seriesID),
		XLarge:   fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["XLarge"], seriesID),
		Original: fmt.Sprintf("https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=85,width=%d/keyart/%s-title_logo-en-us", logoSizes["Original"], seriesID),
	}
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
	searchTitle := anime.Title.Romaji
	if searchTitle == "" {
		searchTitle = anime.Title.English
	}
	if searchTitle == "" {
		return
	}

	subCount, dubCount, err := streaming.GetStreamingCounts(searchTitle)
	if err != nil && anime.Title.English != "" && anime.Title.English != searchTitle {
		subCount, dubCount, err = streaming.GetStreamingCounts(anime.Title.English)
		if err == nil {
			searchTitle = anime.Title.English
		}
	}

	if err != nil {
		logger.Warnf("AnimeService", "Failed to fetch streaming counts: %v", err)
		return
	}

	anime.SubbedCount = subCount
	anime.DubbedCount = dubCount

	if len(anime.Episodes) > 0 {
		episodeNumbers := make([]int, len(anime.Episodes))
		for i, episode := range anime.Episodes {
			episodeNumbers[i] = episode.EpisodeNumber
		}
		sourcesMap, err := streaming.FetchAllEpisodeSources(searchTitle, episodeNumbers)
		if err == nil {
			for i := range anime.Episodes {
				episode := &anime.Episodes[i]
				if sources, ok := sourcesMap[episode.EpisodeNumber]; ok {
					subSources := make([]entities.StreamingSource, len(sources.Sub))
					for j, source := range sources.Sub {
						subSources[j] = entities.StreamingSource{URL: source.URL, Server: source.Server, Type: source.Type}
					}
					dubSources := make([]entities.StreamingSource, len(sources.Dub))
					for j, source := range sources.Dub {
						dubSources[j] = entities.StreamingSource{URL: source.URL, Server: source.Server, Type: source.Type}
					}
					episode.StreamInfo = &entities.StreamInfo{
						SubSources: subSources,
						DubSources: dubSources,
					}
				}
			}
		}
	}
}

func generateEpisodeID(malID int, episodeNumber int, title string) string {
	unique := fmt.Sprintf("%d-%d-%s", malID, episodeNumber, title)
	hash := md5.Sum([]byte(unique))
	return fmt.Sprintf("%x", hash)
}

func applySeasonData(anime *entities.Anime, mapping *entities.Mapping) {
	var relatedMappings []entities.Mapping
	malIDSet := make(map[int]bool)

	if mapping.TVDB > 0 {
		tvdbMappings, err := repositories.GetRelatedAnimeByTVDB(mapping.TVDB, mapping.MAL)
		if err == nil && len(tvdbMappings) > 0 {
			logger.Infof("AnimeService", "Found %d related anime via TVDB", len(tvdbMappings))
			for _, relatedMapping := range tvdbMappings {
				if !malIDSet[relatedMapping.MAL] {
					malIDSet[relatedMapping.MAL] = true
					relatedMappings = append(relatedMappings, relatedMapping)
				}
			}
		}
	}

	if mapping.TMDB > 0 {
		tmdbMappings, err := repositories.GetRelatedAnimeByTMDB(mapping.TMDB, mapping.MAL)
		if err == nil && len(tmdbMappings) > 0 {
			logger.Infof("AnimeService", "Found %d related anime via TMDB", len(tmdbMappings))
			for _, relatedMapping := range tmdbMappings {
				if !malIDSet[relatedMapping.MAL] {
					malIDSet[relatedMapping.MAL] = true
					relatedMappings = append(relatedMappings, relatedMapping)
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
			MALID:         relatedMapping.MAL,
			TitleEnglish:  seasonAnime.Data.TitleEnglish,
			TitleRomaji:   seasonAnime.Data.Title,
			ImageOriginal: seasonAnime.Data.Images.JPG.ImageURL,
			Year:          seasonAnime.Data.Year,
			SeasonName:    seasonAnime.Data.Season,
			Type:          seasonAnime.Data.Type,
			Status:        seasonAnime.Data.Status,
		}

		anime.Seasons = append(anime.Seasons, season)
	}

	sortSeasonsByChronology(allSeasons)

	seasonNumberMap := make(map[int]int)
	for i, seasonEntry := range allSeasons {
		seasonNumberMap[seasonEntry.malID] = i + 1
		if seasonEntry.isCurrent {
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
			current, next := seasons[j], seasons[j+1]

			if current.year > next.year {
				seasons[j], seasons[j+1] = seasons[j+1], seasons[j]
			} else if current.year == next.year && current.seasonOrder > next.seasonOrder {
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

	for i := range anime.Themes {
		if err := repositories.CreateOrUpdateGenre(&anime.Themes[i]); err != nil {
			logger.Warnf("AnimeService", "Failed to save theme: %v", err)
		}
	}

	for i := range anime.Demographics {
		if err := repositories.CreateOrUpdateGenre(&anime.Demographics[i]); err != nil {
			logger.Warnf("AnimeService", "Failed to save demographic: %v", err)
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
		for i := range anime.Episodes {
			episode := &anime.Episodes[i]
			if episode.StreamInfo != nil && episode.EpisodeID != "" {
				if err := repositories.SaveEpisodeStreamInfo(anime.ID, episode.EpisodeID, episode.StreamInfo); err != nil {
					logger.Warnf("AnimeService", "Failed to save stream info for episode %s: %v", episode.EpisodeID, err)
				}
			}
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
	if err := repositories.SetAnimeEnriched(anime.MALID); err != nil {
		logger.Warnf("AnimeService", "Failed to stamp enriched_at for anime %d: %v", anime.MALID, err)
	}
	return nil
}
