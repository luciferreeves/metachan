package mal

import (
	"encoding/json"
	"fmt"
	"metachan/utils/logger"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	crunchyrollOldCDNPattern = regexp.MustCompile(`img\d*\.ak\.crunchyroll\.com/i/spire\d*-tmb/([a-f0-9]{32})`)
	episodeScorePattern      = regexp.MustCompile(`(\d+\.\d+)`)
)

func containsNonLatinCharacters(text string) bool {
	for _, runeValue := range text {
		if runeValue > 127 {
			return true
		}
	}
	return false
}

func splitAlternativeTitle(alternativeTitle string) (romajiTitle string, japaneseTitle string) {
	if alternativeTitle == "" {
		return "", ""
	}
	japaneseMatches := japaneseTextInParensPattern.FindStringSubmatch(alternativeTitle)
	if len(japaneseMatches) > 1 {
		japaneseTitle = japaneseMatches[1]
		romajiTitle = strings.TrimSpace(japaneseTextInParensPattern.ReplaceAllString(alternativeTitle, ""))
		return romajiTitle, japaneseTitle
	}
	if containsNonLatinCharacters(alternativeTitle) {
		return "", alternativeTitle
	}
	return alternativeTitle, ""
}

func parseEpisodeRow(row *goquery.Selection) Episode {
	numberText := strings.TrimSpace(row.Find("td.episode-number").Text())
	episodeNumber, _ := strconv.Atoi(numberText)

	titleCell := row.Find("td.episode-title")
	titleLink := titleCell.Find("a")
	episodeURL, _ := titleLink.Attr("href")

	englishTitle := strings.TrimSpace(titleLink.Text())
	alternativeTitle := strings.TrimSpace(titleCell.Find("span.di-ib").Text())
	romajiTitle, japaneseTitle := splitAlternativeTitle(alternativeTitle)

	airedText := strings.TrimSpace(row.Find("td.episode-aired").Text())

	pollText := strings.TrimSpace(row.Find("td.episode-poll").Text())
	var episodeScore float64
	scoreMatches := episodeScorePattern.FindStringSubmatch(pollText)
	if len(scoreMatches) > 1 {
		episodeScore, _ = strconv.ParseFloat(scoreMatches[1], 64)
	}

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
			Romaji:   romajiTitle,
		},
		Aired:    parseAiredDateString(airedText),
		Score:    episodeScore,
		ForumURL: forumURL,
		Filler:   fillerTag.Length() > 0,
		Recap:    recapTag.Length() > 0,
	}
}

type aroundVideoEntry struct {
	EpisodeNumber int    `json:"episode_number"`
	Thumbnail     string `json:"thumbnail"`
}

func extractEpisodeThumbnailsFromScript(document *goquery.Document) map[int]string {
	thumbnails := make(map[int]string)

	document.Find("script").Each(func(index int, scriptElement *goquery.Selection) {
		scriptContent := scriptElement.Text()
		if !strings.Contains(scriptContent, "aroundVideos") {
			return
		}

		videosStartIndex := strings.Index(scriptContent, `videos`)
		if videosStartIndex == -1 {
			return
		}

		bracketStartIndex := strings.Index(scriptContent[videosStartIndex:], "[")
		if bracketStartIndex == -1 {
			return
		}

		arrayStartIndex := videosStartIndex + bracketStartIndex
		bracketDepth := 0
		arrayEndIndex := -1
		for charIndex := arrayStartIndex; charIndex < len(scriptContent); charIndex++ {
			if scriptContent[charIndex] == '[' {
				bracketDepth++
			} else if scriptContent[charIndex] == ']' {
				bracketDepth--
				if bracketDepth == 0 {
					arrayEndIndex = charIndex + 1
					break
				}
			}
		}

		if arrayEndIndex == -1 {
			return
		}

		videosJSON := scriptContent[arrayStartIndex:arrayEndIndex]
		var videoEntries []aroundVideoEntry
		if unmarshalErr := json.Unmarshal([]byte(videosJSON), &videoEntries); unmarshalErr != nil {
			return
		}

		for _, entry := range videoEntries {
			if entry.Thumbnail != "" {
				unescapedURL := strings.ReplaceAll(entry.Thumbnail, `\/`, `/`)
				thumbnails[entry.EpisodeNumber] = unescapedURL
			}
		}
	})

	return thumbnails
}

func buildCrunchyrollThumbnail(thumbnailHash string) Image {
	cdnBase := "https://imgsrv.crunchyroll.com/cdn-cgi/image/fit=contain,format=auto,quality=70"
	imagePath := fmt.Sprintf("/catalog/crunchyroll/%s.jpg", thumbnailHash)

	return Image{
		Small:    fmt.Sprintf("%s,width=320%s", cdnBase, imagePath),
		Medium:   fmt.Sprintf("%s,width=640%s", cdnBase, imagePath),
		Large:    fmt.Sprintf("%s,width=1280%s", cdnBase, imagePath),
		Original: fmt.Sprintf("%s,width=1920%s", cdnBase, imagePath),
	}
}

func buildEpisodeThumbnail(rawURL string) Image {
	crunchyrollMatches := crunchyrollOldCDNPattern.FindStringSubmatch(rawURL)
	if len(crunchyrollMatches) > 1 {
		return buildCrunchyrollThumbnail(crunchyrollMatches[1])
	}
	return Image{
		Original: rawURL,
	}
}

func extractEpisodeSynopsis(document *goquery.Document) string {
	var synopsisText string

	document.Find("h2").Each(func(index int, headingElement *goquery.Selection) {
		if synopsisText != "" {
			return
		}
		if !strings.Contains(headingElement.Text(), "Synopsis") {
			return
		}

		nextSibling := headingElement.Next()
		for nextSibling.Length() > 0 {
			tagName := goquery.NodeName(nextSibling)
			if tagName == "h2" || tagName == "h3" || tagName == "br" {
				break
			}
			if nextSibling.HasClass("border_top") {
				break
			}
			text := strings.TrimSpace(nextSibling.Text())
			if text != "" && !strings.Contains(text, "No synopsis information") {
				synopsisText = text
				return
			}
			nextSibling = nextSibling.Next()
		}
	})

	if synopsisText == "" {
		metaDescription, exists := document.Find(`meta[property="og:description"]`).Attr("content")
		if exists {
			trimmedDescription := strings.TrimSpace(metaDescription)
			if trimmedDescription != "" && !strings.Contains(trimmedDescription, "No synopsis information") {
				synopsisText = trimmedDescription
			}
		}
	}

	return synopsisText
}

func enrichEpisodesWithDetails(episodes []Episode, malID int) {
	if len(episodes) == 0 {
		return
	}

	logger.Debugf("MALScraper", "Enriching %d episodes with details for MAL ID %d", len(episodes), malID)

	thumbnailMap := make(map[int]string)
	thumbnailsExtracted := false

	for episodeIndex := range episodes {
		if episodes[episodeIndex].URL == "" {
			continue
		}

		logger.Debugf("MALScraper", "Fetching episode %d/%d detail page for MAL ID %d",
			episodes[episodeIndex].Number, len(episodes), malID)

		episodeDocument, fetchErr := makeRequest(episodes[episodeIndex].URL)
		if fetchErr != nil {
			logger.Warnf("MALClient", "Failed to fetch episode %d detail page for MAL ID %d: %v",
				episodes[episodeIndex].Number, malID, fetchErr)
			continue
		}

		if !thumbnailsExtracted {
			thumbnailMap = extractEpisodeThumbnailsFromScript(episodeDocument)
			thumbnailsExtracted = true
			logger.Debugf("MALScraper", "Extracted %d episode thumbnails from aroundVideos script", len(thumbnailMap))
		}

		episodes[episodeIndex].Synopsis = extractEpisodeSynopsis(episodeDocument)

		if thumbnailURL, exists := thumbnailMap[episodes[episodeIndex].Number]; exists {
			episodes[episodeIndex].Preview = Preview{
				URL:       episodes[episodeIndex].URL,
				Thumbnail: buildEpisodeThumbnail(thumbnailURL),
			}
			logger.Debugf("MALScraper", "Episode %d: synopsis=%d chars, thumbnail=yes",
				episodes[episodeIndex].Number, len(episodes[episodeIndex].Synopsis))
		} else {
			logger.Debugf("MALScraper", "Episode %d: synopsis=%d chars, thumbnail=no",
				episodes[episodeIndex].Number, len(episodes[episodeIndex].Synopsis))
		}
	}
}

func GetAnimeEpisodesByMALID(malID int) ([]Episode, error) {
	var allEpisodes []Episode
	offset := 0

	for {
		pageURL := fmt.Sprintf("%s/anime/%d/_/episode?offset=%d", malBaseURL, malID, offset)
		logger.Debugf("MALScraper", "Fetching episode list page at offset %d for MAL ID %d", offset, malID)
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

	enrichEpisodesWithDetails(allEpisodes, malID)

	return allEpisodes, nil
}