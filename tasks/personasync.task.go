package tasks

import (
	"fmt"
	"metachan/entities"
	"metachan/repositories"
	"metachan/utils/api/jikan"
	"metachan/utils/logger"
	"time"
)

func ResumePersonEnrichment() {
	go PersonSync()
}

func PersonSync() error {
	stubs, err := repositories.GetAllPersonStubs()
	if err != nil {
		return fmt.Errorf("failed to load person stubs: %w", err)
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

		resp, err := jikan.GetPersonByMALID(s.MALID)
		if err != nil {
			logger.Warnf("PersonSync", "Failed to fetch person %d: %v", s.MALID, err)
			continue
		}

		d := resp.Data

		var birthday *time.Time
		if d.Birthday != nil && *d.Birthday != "" {
			layouts := []string{
				time.RFC3339,
				"2006-01-02T15:04:05-07:00",
				"2006-01-02",
			}
			for _, layout := range layouts {
				if t, err := time.Parse(layout, *d.Birthday); err == nil {
					birthday = &t
					break
				}
			}
		}

		var websiteURL string
		if d.WebsiteURL != nil {
			websiteURL = *d.WebsiteURL
		}

		var voiceRoles []entities.PersonVoiceRole
		for _, v := range d.Voices {
			voiceRoles = append(voiceRoles, entities.PersonVoiceRole{
				Role:              v.Role,
				AnimeMALID:        v.Anime.MALID,
				AnimeTitle:        v.Anime.Title,
				AnimeURL:          v.Anime.URL,
				AnimeImageURL:     v.Anime.Images.JPG.ImageURL,
				CharacterMALID:    v.Character.MALID,
				CharacterName:     v.Character.Name,
				CharacterURL:      v.Character.URL,
				CharacterImageURL: v.Character.Images.JPG.ImageURL,
			})
		}

		var animeCredits []entities.PersonAnimeCredit
		for _, a := range d.Anime {
			animeCredits = append(animeCredits, entities.PersonAnimeCredit{
				Position:      a.Position,
				AnimeMALID:    a.Anime.MALID,
				AnimeTitle:    a.Anime.Title,
				AnimeURL:      a.Anime.URL,
				AnimeImageURL: a.Anime.Images.JPG.ImageURL,
			})
		}

		var mangaCredits []entities.PersonMangaCredit
		for _, m := range d.Manga {
			mangaCredits = append(mangaCredits, entities.PersonMangaCredit{
				Position:      m.Position,
				MangaMALID:    m.Manga.MALID,
				MangaTitle:    m.Manga.Title,
				MangaURL:      m.Manga.URL,
				MangaImageURL: m.Manga.Images.JPG.ImageURL,
			})
		}

		if err := repositories.UpdatePersonDetails(
			d.MALID,
			d.URL, websiteURL, d.Images.JPG.ImageURL,
			d.Name, d.GivenName, d.FamilyName,
			d.AlternateNames, birthday,
			d.Favorites, d.About,
			voiceRoles, animeCredits, mangaCredits,
		); err != nil {
			logger.Warnf("PersonSync", "Failed to update person %d: %v", s.MALID, err)
			continue
		}

		if err := repositories.SetPersonEnriched(d.MALID); err != nil {
			logger.Warnf("PersonSync", "Failed to stamp enriched_at for person %d: %v", d.MALID, err)
		}

		enriched++
		if (i+1)%50 == 0 || (i+1) == total {
			progress, eta := calculateProgress(i+1, total, startTime)
			logger.Infof("PersonSync", "Enriching: %d/%d (%.1f%%) | ETA: %v", i+1, total, progress, eta)
		}
	}

	logger.Successf("PersonSync", "Background enrichment complete. Enriched %d people", enriched)
	return nil
}
