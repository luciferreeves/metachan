package anime

import (
	"fmt"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
)

func GetAnimeDetails(animeMapping *entities.AnimeMapping) (*types.Anime, error) {
	malID := animeMapping.MAL

	anime, err := getFullAnimeViaJikan(malID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime details: %w", err)
	}
	var anilistAnime *types.AnilistAnimeResponse
	if animeMapping.Anilist != 0 {
		anilistAnime, err = getAnimeViaAnilist(animeMapping.Anilist)
		if err != nil {
			return nil, fmt.Errorf("failed to get anime details from Anilist: %w", err)
		}
	}

	episodes, err := getAnimeEpisodesViaJikan(malID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime episodes: %w", err)
	}

	episodeData, _ := generateEpisodeDataWithDescriptions(
		episodes.Data,
		anime.Data.Title,
		anime.Data.TitleEnglish,
		animeMapping.TMDB,
	)

	var logos types.AnimeLogos
	malSyncData, err := getAnimeViaMalSync(malID)
	if err == nil {
		logos = extractLogosFromMALSync(malSyncData)
	} else {
		logger.Log(fmt.Sprintf("Failed to get MALSync data for logos: %v", err), types.LogOptions{
			Level:  types.Debug,
			Prefix: "AnimeAPI",
		})
		logos = types.AnimeLogos{}
	}

	// Get seasons information if TVDB ID is available
	var seasons []types.AnimeSeason
	if animeMapping.TVDB != 0 {
		// Get all mappings for this TVDB ID (representing different seasons)
		seasonMappings, err := FindSeasonMappings(animeMapping.TVDB)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to find season mappings: %v", err), types.LogOptions{
				Level:  types.Warn,
				Prefix: "AnimeAPI",
			})
		} else if len(seasonMappings) > 0 {
			// Process the season mappings to get season details
			seasons, err = GetAnimeSeason(&seasonMappings, malID)
			if err != nil {
				logger.Log(fmt.Sprintf("Failed to get anime seasons: %v", err), types.LogOptions{
					Level:  types.Warn,
					Prefix: "AnimeAPI",
				})
			}
		}
	}

	characterResponse, err := getAnimeCharactersViaJikan(malID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime characters: %w", err)
	}

	characters := getAnimeCharacters(characterResponse)

	animeDetails := &types.Anime{
		MALID: malID,
		Titles: types.AnimeTitles{
			Romaji:   anime.Data.Title,
			English:  anime.Data.TitleEnglish,
			Japanese: anime.Data.TitleJapanese,
			Synonyms: anime.Data.TitleSynonyms,
		},
		Synopsis: anime.Data.Synopsis,
		Type:     types.AniSyncType(animeMapping.Type),
		Source:   anime.Data.Source,
		Airing:   anime.Data.Airing,
		Status:   anime.Data.Status,
		AiringStatus: types.AiringStatus{
			From: types.AiringStatusDates{
				Day:    anime.Data.Aired.Prop.From.Day,
				Month:  anime.Data.Aired.Prop.From.Month,
				Year:   anime.Data.Aired.Prop.From.Year,
				String: anime.Data.Aired.From,
			},
			To: types.AiringStatusDates{
				Day:    anime.Data.Aired.Prop.To.Day,
				Month:  anime.Data.Aired.Prop.To.Month,
				Year:   anime.Data.Aired.Prop.To.Year,
				String: anime.Data.Aired.To,
			},
			String: anime.Data.Aired.String,
		},
		Duration: anime.Data.Duration,
		Images: types.AnimeImages{
			Small:    anime.Data.Images.JPG.SmallImageURL,
			Large:    anime.Data.Images.JPG.LargeImageURL,
			Original: anime.Data.Images.JPG.ImageURL,
		},
		Logos: logos,
		Covers: types.AnimeImages{
			Small:    anilistAnime.Data.Media.CoverImage.Medium,
			Large:    anilistAnime.Data.Media.CoverImage.Large,
			Original: anilistAnime.Data.Media.CoverImage.ExtraLarge,
		},
		Color:  anilistAnime.Data.Media.CoverImage.Color,
		Genres: generateGenres(anime.Data.Genres, anime.Data.ExplicitGenres),
		Scores: types.AnimeScores{
			Score:      anime.Data.Score,
			ScoredBy:   anime.Data.ScoredBy,
			Rank:       anime.Data.Rank,
			Popularity: anime.Data.Popularity,
			Members:    anime.Data.Members,
			Favorites:  anime.Data.Favorites,
		},
		Season: anime.Data.Season,
		Year:   anime.Data.Year,
		Broadcast: types.AnimeBroadcast{
			Day:      anime.Data.Broadcast.Day,
			Time:     anime.Data.Broadcast.Time,
			Timezone: anime.Data.Broadcast.Timezone,
			String:   anime.Data.Broadcast.String,
		},
		Producers: generateProducers(anime.Data.Producers),
		Studios:   generateStudios(anime.Data.Studios),
		Licensors: generateLicensors(anime.Data.Licensors),
		Seasons:   seasons, // Add seasons information
		Episodes: types.AnimeEpisodes{
			Total:    getEpisodeCount(anime, anilistAnime),
			Aired:    len(episodes.Data),
			Episodes: episodeData,
		},
		Characters: characters,
		Mappings: types.AnimeMappings{
			AniDB:          animeMapping.AniDB,
			Anilist:        animeMapping.Anilist,
			AnimeCountdown: animeMapping.AnimeCountdown,
			AnimePlanet:    animeMapping.AnimePlanet,
			AniSearch:      animeMapping.AniSearch,
			IMDB:           animeMapping.IMDB,
			Kitsu:          animeMapping.Kitsu,
			LiveChart:      animeMapping.LiveChart,
			NotifyMoe:      animeMapping.NotifyMoe,
			Simkl:          animeMapping.Simkl,
			TMDB:           animeMapping.TMDB,
			TVDB:           animeMapping.TVDB,
		},
	}
	return animeDetails, nil
}

func getEpisodeCount(malAnime *types.JikanAnimeResponse, anilistAnime *types.AnilistAnimeResponse) int {
	streamingScheduleLength := len(anilistAnime.Data.Media.AiringSchedule.Nodes)
	episodes := max(malAnime.Data.Episodes, anilistAnime.Data.Media.Episodes)
	episodes = max(episodes, streamingScheduleLength)

	return episodes
}

func getAnimeCharacters(characterResponse *types.JikanAnimeCharacterResponse) []types.AnimeCharacter {
	characters := make([]types.AnimeCharacter, len(characterResponse.Data))
	for i, character := range characterResponse.Data {
		characters[i] = types.AnimeCharacter{
			MALID:       character.Character.MALID,
			URL:         character.Character.URL,
			ImageURL:    character.Character.Images.JPG.ImageURL,
			Name:        character.Character.Name,
			Role:        character.Role,
			VoiceActors: make([]types.AnimeVoiceActor, len(character.VoiceActors)),
		}

		for j, voiceActor := range character.VoiceActors {
			characters[i].VoiceActors[j] = types.AnimeVoiceActor{
				MALID:    voiceActor.Person.MALID,
				URL:      voiceActor.Person.URL,
				Image:    voiceActor.Person.Images.JPG.ImageURL,
				Name:     voiceActor.Person.Name,
				Language: voiceActor.Language,
			}
		}
	}
	return characters
}

func generateGenres(genres, explicitGenres []types.JikanGenericMALStructure) []types.AnimeGenres {
	animeGenres := make([]types.AnimeGenres, len(genres)+len(explicitGenres))
	counter := 0

	for _, genre := range genres {
		animeGenres[counter] = types.AnimeGenres{
			Name:    genre.Name,
			GenreID: genre.MALID,
			URL:     genre.URL,
		}
		counter++
	}

	for _, genre := range explicitGenres {
		animeGenres[counter] = types.AnimeGenres{
			Name:    genre.Name,
			GenreID: genre.MALID,
			URL:     genre.URL,
		}
		counter++
	}

	return animeGenres
}

func generateStudios(genericPLS []types.JikanGenericMALStructure) []types.AnimeStudio {
	studios := make([]types.AnimeStudio, len(genericPLS))
	for i, studio := range genericPLS {
		studios[i] = types.AnimeStudio{
			Name:     studio.Name,
			StudioID: studio.MALID,
			URL:      studio.URL,
		}
	}
	return studios
}

func generateProducers(genericPLS []types.JikanGenericMALStructure) []types.AnimeProducer {
	producers := make([]types.AnimeProducer, len(genericPLS))
	for i, producer := range genericPLS {
		producers[i] = types.AnimeProducer{
			Name:       producer.Name,
			ProducerID: producer.MALID,
			URL:        producer.URL,
		}
	}
	return producers
}

func generateLicensors(genericPLS []types.JikanGenericMALStructure) []types.AnimeLicensor {
	licensors := make([]types.AnimeLicensor, len(genericPLS))
	for i, licensor := range genericPLS {
		licensors[i] = types.AnimeLicensor{
			Name:       licensor.Name,
			ProducerID: licensor.MALID,
			URL:        licensor.URL,
		}
	}
	return licensors
}
