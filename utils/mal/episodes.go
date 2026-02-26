package mal

import (
	"fmt"
	"metachan/utils/logger"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseEpisodeRow(row *goquery.Selection) Episode {
	numberText := strings.TrimSpace(row.Find("td.episode-number").Text())
	episodeNumber, _ := strconv.Atoi(numberText)

	titleCell := row.Find("td.episode-title")
	titleLink := titleCell.Find("a")
	episodeURL, _ := titleLink.Attr("href")

	englishTitle := strings.TrimSpace(titleLink.Text())
	japaneseTitle := strings.TrimSpace(titleCell.Find("span.di-ib").Text())

	airedText := strings.TrimSpace(row.Find("td.episode-aired").Text())

	forumLink := row.Find("td.episode-forum a")
	forumURL, _ := forumLink.Attr("href")

	fillerTag := row.Find("span.filler")
	recapTag := row.Find("span.recap")

	return Episode{
		Number: episodeNumber,
		URL:    episodeURL,
		Title: Title{
			English:  englishTitle,
			Japanese: japaneseTitle,
		},
		Aired:    parseAiredDateString(airedText),
		ForumURL: forumURL,
		Filler:   fillerTag.Length() > 0,
		Recap:    recapTag.Length() > 0,
	}
}

func GetAnimeEpisodesByMALID(malID int) ([]Episode, error) {
	var allEpisodes []Episode
	offset := 0

	for {
		pageURL := fmt.Sprintf("%s/anime/%d/_/episode?offset=%d", malBaseURL, malID, offset)
		document, fetchErr := makeRequest(pageURL)
		if fetchErr != nil {
			if len(allEpisodes) > 0 {
				logger.Warnf("MALClient", "Failed to fetch episodes page at offset %d for MAL ID %d: %v", offset, malID, fetchErr)
				break
			}
			logger.Errorf("MALClient", "Failed to fetch episodes for MAL ID %d: %v", malID, fetchErr)
			return nil, fmt.Errorf("failed to fetch episodes for MAL ID %d: %w", malID, fetchErr)
		}

		episodeRows := document.Find("table.episode_list tbody tr")
		if episodeRows.Length() == 0 {
			break
		}

		episodeRows.Each(func(index int, row *goquery.Selection) {
			episode := parseEpisodeRow(row)
			if episode.Number > 0 {
				allEpisodes = append(allEpisodes, episode)
			}
		})

		nextPageLink := document.Find("a.link-blue-box.next")
		if nextPageLink.Length() == 0 {
			break
		}

		offset += 100
	}

	return allEpisodes, nil
}