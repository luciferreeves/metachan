package tasks

import (
	"fmt"
	"metachan/entities"
	"metachan/repositories"
	"metachan/utils/api/jikan"
	"metachan/utils/logger"
	"time"
)

func ResumeCharacterEnrichment() {
	go CharacterSync()
}

func CharacterSync() error {
	stubs, err := repositories.GetAllCharacterStubs()
	if err != nil {
		return fmt.Errorf("failed to load character stubs: %w", err)
	}

	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)

	hasWork := false
	for _, s := range stubs {
		if s.EnrichedAt == nil || !s.EnrichedAt.After(sevenDaysAgo) {
			hasWork = true
			break
		}
	}
	if !hasWork {
		return nil
	}

	total := len(stubs)
	startTime := time.Now()
	enriched := 0

	for i, s := range stubs {
		if s.EnrichedAt != nil && s.EnrichedAt.After(sevenDaysAgo) {
			continue
		}

		resp, err := jikan.GetCharacterByMALID(s.MALID)
		if err != nil {
			logger.Warnf("CharacterSync", "Failed to fetch character %d: %v", s.MALID, err)
			continue
		}

		d := resp.Data

		var voiceActors []entities.CharacterVoiceActor
		for _, v := range d.Voices {
			voiceActors = append(voiceActors, entities.CharacterVoiceActor{
				Language: v.Language,
				Person:   &entities.Person{},
			})
		}

		var animeAppearances []entities.CharacterAnimeAppearance
		for _, a := range d.Anime {
			animeAppearances = append(animeAppearances, entities.CharacterAnimeAppearance{
				AnimeMALID: a.Anime.MALID,
				Title:      a.Anime.Title,
				URL:        a.Anime.URL,
				ImageURL:   a.Anime.Images.JPG.ImageURL,
				Role:       a.Role,
			})
		}

		if err := repositories.UpdateCharacterDetails(
			d.MALID, d.Name, d.NameKanji, d.URL, d.Images.JPG.ImageURL,
			d.About, d.Nicknames, d.Favorites, voiceActors, animeAppearances,
		); err != nil {
			logger.Warnf("CharacterSync", "Failed to update character %d: %v", s.MALID, err)
			continue
		}

		if err := repositories.SetCharacterEnriched(d.MALID); err != nil {
			logger.Warnf("CharacterSync", "Failed to stamp enriched_at for character %d: %v", d.MALID, err)
		}

		enriched++
		if (i+1)%50 == 0 || (i+1) == total {
			progress, eta := calculateProgress(i+1, total, startTime)
			logger.Infof("CharacterSync", "Enriching: %d/%d (%.1f%%) | ETA: %v", i+1, total, progress, eta)
		}
	}

	logger.Successf("CharacterSync", "Background enrichment complete. Enriched %d characters", enriched)
	return nil
}
