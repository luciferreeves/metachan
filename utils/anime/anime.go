package anime

import (
	"fmt"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
)

func GetAnimeDetails(animeMapping *entities.AnimeMapping) (*types.Anime, error) {
	malID := animeMapping.MAL

	anime, err := getAnimeViaJikan(malID)
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

	episodeData, err := generateEpisodeDataWithDescriptions(
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
		Status:   anime.Data.Status,
		Duration: anime.Data.Duration,
		Logos:    logos,
		Episodes: types.AnimeEpisodes{
			Total:    getEpisodeCount(anime, anilistAnime),
			Aired:    len(episodes.Data),
			Episodes: episodeData,
		},
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
