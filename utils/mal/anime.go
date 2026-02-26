package mal

import (
	"fmt"
	"metachan/utils/logger"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	producerIDPattern           = regexp.MustCompile(`/anime/producer/(\d+)`)
	genreIDPattern              = regexp.MustCompile(`/anime/genre/(\d+)`)
	youtubeIDPattern            = regexp.MustCompile(`/embed/([a-zA-Z0-9_-]+)`)
	themeSongTitlePattern       = regexp.MustCompile(`"(.+?)"`)
	themeSongArtistPattern      = regexp.MustCompile(`by\s+(.+?)(?:\s+\(eps|\s*$)`)
	themeSongEpisodesPattern    = regexp.MustCompile(`\(eps\s+([\d,\-\s]+)\)`)
	japaneseTextInParensPattern = regexp.MustCompile(`\(([^)]*[^\x00-\x7F][^)]*)\)`)
	broadcastTimePattern        = regexp.MustCompile(`(\w+)s?\s+at\s+(\d{2}:\d{2})\s+\((\w+)\)`)
	imageResizePrefixPattern    = regexp.MustCompile(`/r/\d+x\d+`)
	leadingIndexPattern         = regexp.MustCompile(`^#?\d+:?\s*`)
	trailingEpisodeInfoPattern  = regexp.MustCompile(`\s*\(eps\s.*$`)
)

const airedDateLayout = "Jan 2, 2006"

func extractSidebarValue(document *goquery.Document, label string) string {
	var extractedValue string
	document.Find("span.dark_text").Each(func(index int, selection *goquery.Selection) {
		if strings.TrimSpace(selection.Text()) == label {
			parentClone := selection.Parent().Clone()
			parentClone.Find("span.dark_text").Remove()
			parentClone.Find("sup").Remove()
			parentClone.Find("div").Remove()
			extractedValue = strings.TrimSpace(parentClone.Text())
		}
	})
	return extractedValue
}

func extractSidebarMALIDs(document *goquery.Document, label string, idPattern *regexp.Regexp) []int {
	var malIDs []int
	document.Find("span.dark_text").Each(func(index int, selection *goquery.Selection) {
		if strings.TrimSpace(selection.Text()) != label {
			return
		}
		parentNode := selection.Parent()
		if strings.Contains(parentNode.Text(), "None found") || strings.Contains(parentNode.Text(), "No genres") {
			return
		}
		parentNode.Find("a").Each(func(linkIndex int, linkElement *goquery.Selection) {
			href, exists := linkElement.Attr("href")
			if !exists {
				return
			}
			matches := idPattern.FindStringSubmatch(href)
			if len(matches) > 1 {
				parsedID, parseErr := strconv.Atoi(matches[1])
				if parseErr == nil {
					malIDs = append(malIDs, parsedID)
				}
			}
		})
	})
	return malIDs
}

func extractSidebarMALIDsMultiLabel(document *goquery.Document, labels []string, idPattern *regexp.Regexp) []int {
	for _, label := range labels {
		malIDs := extractSidebarMALIDs(document, label, idPattern)
		if len(malIDs) > 0 {
			return malIDs
		}
	}
	return nil
}

func buildImageFromBaseURL(rawURL string) Image {
	cleanedURL := imageResizePrefixPattern.ReplaceAllString(rawURL, "")
	extensionIndex := strings.LastIndex(cleanedURL, ".")
	if extensionIndex == -1 {
		return Image{}
	}

	pathBase := cleanedURL[:extensionIndex]

	return Image{
		Small:    pathBase + "t.jpg",
		Medium:   pathBase + ".jpg",
		Large:    pathBase + "l.jpg",
		Original: pathBase + ".jpg",
	}
}

func buildYouTubeThumbnail(videoID string) Image {
	thumbnailBase := fmt.Sprintf("https://img.youtube.com/vi/%s", videoID)
	return Image{
		Small:    thumbnailBase + "/default.jpg",
		Medium:   thumbnailBase + "/mqdefault.jpg",
		Large:    thumbnailBase + "/hqdefault.jpg",
		Original: thumbnailBase + "/maxresdefault.jpg",
	}
}

func parseAiredDateString(dateString string) AiredDate {
	trimmedDate := strings.TrimSpace(dateString)
	if trimmedDate == "" || trimmedDate == "?" || trimmedDate == "Not available" {
		return AiredDate{}
	}
	parsedTime, parseErr := time.Parse(airedDateLayout, trimmedDate)
	if parseErr != nil {
		return AiredDate{String: trimmedDate}
	}
	return AiredDate{
		Day:    parsedTime.Day(),
		Month:  int(parsedTime.Month()),
		Year:   parsedTime.Year(),
		String: trimmedDate,
	}
}

func parseIntFromText(text string) int {
	cleanedText := strings.ReplaceAll(strings.TrimSpace(text), ",", "")
	cleanedText = strings.TrimPrefix(cleanedText, "#")
	parsedValue, _ := strconv.Atoi(cleanedText)
	return parsedValue
}

func parseFloatFromText(text string) float64 {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "N/A" || trimmedText == "" {
		return 0
	}
	parsedValue, _ := strconv.ParseFloat(trimmedText, 64)
	return parsedValue
}

func parseAnimeTitle(document *goquery.Document) Title {
	var animeTitle Title
	romajiTitle, _ := document.Find(`meta[property="og:title"]`).Attr("content")
	animeTitle.Romaji = strings.TrimSpace(romajiTitle)

	document.Find("span.dark_text").Each(func(index int, selection *goquery.Selection) {
		label := strings.TrimSpace(selection.Text())
		parentClone := selection.Parent().Clone()
		parentClone.Find("span.dark_text").Remove()
		value := strings.TrimSpace(parentClone.Text())

		switch label {
		case "English:":
			animeTitle.English = value
		case "Japanese:":
			animeTitle.Japanese = value
		case "Synonyms:":
			if value != "" {
				animeTitle.Synonyms = strings.Split(value, ", ")
			}
		}
	})

	return animeTitle
}

func parseAnimeImage(document *goquery.Document) Image {
	imageURL, exists := document.Find(`meta[property="og:image"]`).Attr("content")
	if !exists || imageURL == "" {
		return Image{}
	}
	return buildImageFromBaseURL(imageURL)
}

func parseAnimeStatistics(document *goquery.Document) Statistics {
	return Statistics{
		Score:      parseFloatFromText(document.Find(`span[itemprop="ratingValue"]`).Text()),
		ScoredBy:   parseIntFromText(document.Find(`span[itemprop="ratingCount"]`).Text()),
		Rank:       parseIntFromText(extractSidebarValue(document, "Ranked:")),
		Popularity: parseIntFromText(extractSidebarValue(document, "Popularity:")),
		Members:    parseIntFromText(extractSidebarValue(document, "Members:")),
		Favorites:  parseIntFromText(extractSidebarValue(document, "Favorites:")),
	}
}

func parseAnimeSynopsis(document *goquery.Document) string {
	synopsisNode := document.Find(`p[itemprop="description"]`)
	if synopsisNode.Length() == 0 {
		return ""
	}
	synopsisText := strings.TrimSpace(synopsisNode.Text())
	if strings.Contains(synopsisText, "No synopsis information has been added") {
		return ""
	}
	return synopsisText
}

func parseAnimeBackground(document *goquery.Document) string {
	backgroundHeading := document.Find("h2#background")
	if backgroundHeading.Length() == 0 {
		return ""
	}

	wrapperDiv := backgroundHeading.Parent()
	container := wrapperDiv.Parent()

	var foundWrapper bool
	var reachedEnd bool
	var backgroundParts []string

	container.Contents().Each(func(index int, node *goquery.Selection) {
		if reachedEnd {
			return
		}
		if !foundWrapper {
			if node.Find("h2#background").Length() > 0 {
				foundWrapper = true
			}
			return
		}
		if node.HasClass("border_top") {
			reachedEnd = true
			return
		}
		text := strings.TrimSpace(node.Text())
		if text != "" && !strings.Contains(text, "No background information") {
			backgroundParts = append(backgroundParts, text)
		}
	})

	return strings.Join(backgroundParts, " ")
}

func parseAnimeTrailer(document *goquery.Document) Trailer {
	trailerLink := document.Find("div.video-promotion a")
	if trailerLink.Length() == 0 {
		return Trailer{}
	}
	embedURL, _ := trailerLink.Attr("href")
	youtubeMatches := youtubeIDPattern.FindStringSubmatch(embedURL)
	if len(youtubeMatches) < 2 {
		return Trailer{EmbedURL: embedURL, Preview: Preview{URL: embedURL}}
	}
	videoID := youtubeMatches[1]
	return Trailer{
		YoutubeID: videoID,
		EmbedURL:  embedURL,
		Preview: Preview{
			URL:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
			Thumbnail: buildYouTubeThumbnail(videoID),
		},
	}
}

func parseAnimePremiered(document *goquery.Document) Premiered {
	text := extractSidebarValue(document, "Premiered:")
	if text == "" || text == "?" {
		return Premiered{}
	}
	parts := strings.SplitN(text, " ", 2)
	if len(parts) != 2 {
		return Premiered{}
	}
	year, _ := strconv.Atoi(parts[1])
	return Premiered{Season: Season(parts[0]), Year: year}
}

func parseAnimeAired(document *goquery.Document) Aired {
	text := extractSidebarValue(document, "Aired:")
	if text == "" || text == "Not available" {
		return Aired{}
	}
	parts := strings.SplitN(text, " to ", 2)
	aired := Aired{String: text}
	if len(parts) >= 1 {
		aired.From = parseAiredDateString(parts[0])
	}
	if len(parts) >= 2 {
		aired.To = parseAiredDateString(parts[1])
	}
	return aired
}

func parseAnimeBroadcast(document *goquery.Document) Broadcast {
	text := extractSidebarValue(document, "Broadcast:")
	if text == "" {
		return Broadcast{}
	}
	matches := broadcastTimePattern.FindStringSubmatch(text)
	if len(matches) == 4 {
		return Broadcast{Day: matches[1], Time: matches[2], Timezone: matches[3], String: text}
	}
	return Broadcast{String: text}
}

func parseAnimeThemeSongs(document *goquery.Document, containerClass string) []ThemeSong {
	var themeSongs []ThemeSong
	document.Find(fmt.Sprintf("div.%s > table tr", containerClass)).Each(func(index int, row *goquery.Selection) {
		songCell := row.Find("td").Eq(1)
		if songCell.Length() == 0 {
			songCell = row.Find("td").First()
		}

		if songCell.Find("span.theme-song-index").Length() == 0 {
			cellText := strings.TrimSpace(songCell.Text())
			if cellText == "" || !strings.Contains(cellText, `"`) {
				return
			}
		}

		cellText := strings.TrimSpace(songCell.Text())
		if cellText == "" || strings.Contains(cellText, "No opening themes") || strings.Contains(cellText, "No ending themes") {
			return
		}

		themeSong := parseThemeSongText(cellText)

		songCell.Find("input[type=hidden]").Each(func(inputIndex int, inputElement *goquery.Selection) {
			inputID, _ := inputElement.Attr("id")
			inputValue, _ := inputElement.Attr("value")
			if inputValue == "" {
				return
			}
			switch {
			case strings.HasPrefix(inputID, "spotify_url"):
				themeSong.Links = append(themeSong.Links, ExternalLink{Name: "Spotify", URL: inputValue})
			case strings.HasPrefix(inputID, "apple_url"):
				themeSong.Links = append(themeSong.Links, ExternalLink{Name: "Apple Music", URL: inputValue})
			case strings.HasPrefix(inputID, "amazon_url"):
				themeSong.Links = append(themeSong.Links, ExternalLink{Name: "Amazon Music", URL: inputValue})
			case strings.HasPrefix(inputID, "youtube_url"):
				themeSong.Links = append(themeSong.Links, ExternalLink{Name: "YouTube", URL: inputValue})
			}
		})

		themeSongs = append(themeSongs, themeSong)
	})
	return themeSongs
}

func parseThemeSongText(rawText string) ThemeSong {
	text := leadingIndexPattern.ReplaceAllString(strings.TrimSpace(rawText), "")
	var themeSong ThemeSong

	episodeMatches := themeSongEpisodesPattern.FindStringSubmatch(text)
	if len(episodeMatches) > 1 {
		allNumbers := regexp.MustCompile(`\d+`).FindAllString(episodeMatches[1], -1)
		if len(allNumbers) > 0 {
			themeSong.Episodes.Start, _ = strconv.Atoi(allNumbers[0])
			themeSong.Episodes.End, _ = strconv.Atoi(allNumbers[len(allNumbers)-1])
		}
	}

	titleMatches := themeSongTitlePattern.FindStringSubmatch(text)
	if len(titleMatches) > 1 {
		fullTitle := titleMatches[1]
		japaneseMatches := japaneseTextInParensPattern.FindStringSubmatch(fullTitle)
		if len(japaneseMatches) > 1 {
			themeSong.Title.Japanese = japaneseMatches[1]
			themeSong.Title.Romaji = strings.TrimSpace(japaneseTextInParensPattern.ReplaceAllString(fullTitle, ""))
		} else {
			themeSong.Title.Romaji = fullTitle
		}
	}

	artistMatches := themeSongArtistPattern.FindStringSubmatch(text)
	if len(artistMatches) > 1 {
		themeSong.Artist = strings.TrimSpace(trailingEpisodeInfoPattern.ReplaceAllString(artistMatches[1], ""))
	}

	return themeSong
}

func parseAnimeExternalLinks(document *goquery.Document) []ExternalLink {
	var externalLinks []ExternalLink
	document.Find("div.external_links a").Each(func(index int, linkElement *goquery.Selection) {
		href, exists := linkElement.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
			return
		}
		linkName := strings.TrimSpace(linkElement.Find("div.caption").Text())
		if linkName == "" {
			linkName = strings.TrimSpace(linkElement.Text())
		}
		if linkName != "" {
			externalLinks = append(externalLinks, ExternalLink{Name: linkName, URL: href})
		}
	})
	return externalLinks
}

func parseAnimeStreamingLinks(document *goquery.Document) []ExternalLink {
	var streamingLinks []ExternalLink
	document.Find("div.broadcasts a.broadcast-item").Each(func(index int, linkElement *goquery.Selection) {
		href, exists := linkElement.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "javascript:") {
			return
		}
		linkName, _ := linkElement.Attr("title")
		if linkName == "" {
			linkName = strings.TrimSpace(linkElement.Find("div.caption").Text())
		}
		if linkName != "" {
			streamingLinks = append(streamingLinks, ExternalLink{Name: linkName, URL: href})
		}
	})
	return streamingLinks
}

func parsePromotionalVideos(document *goquery.Document) []PromotionalVideo {
	var videos []PromotionalVideo
	document.Find("div.promotional-video section > div").Each(func(index int, videoElement *goquery.Selection) {
		linkElement := videoElement.Find("a")
		if linkElement.Length() == 0 {
			return
		}
		embedURL, _ := linkElement.Attr("href")
		titleText := strings.TrimSpace(linkElement.Find("span").First().Text())

		youtubeMatches := youtubeIDPattern.FindStringSubmatch(embedURL)
		if len(youtubeMatches) < 2 {
			return
		}
		videoID := youtubeMatches[1]

		videos = append(videos, PromotionalVideo{
			Title: Title{Romaji: titleText},
			Preview: Preview{
				URL:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
				Thumbnail: buildYouTubeThumbnail(videoID),
			},
		})
	})
	return videos
}

func parseMusicVideos(document *goquery.Document) []MusicVideo {
	var videos []MusicVideo
	document.Find("div.music-video section > div").Each(func(index int, videoElement *goquery.Selection) {
		linkElement := videoElement.Find("a")
		if linkElement.Length() == 0 {
			return
		}
		embedURL, _ := linkElement.Attr("href")
		titleText := strings.TrimSpace(linkElement.Find("span").First().Text())

		youtubeMatches := youtubeIDPattern.FindStringSubmatch(embedURL)
		if len(youtubeMatches) < 2 {
			return
		}
		videoID := youtubeMatches[1]

		musicVideo := MusicVideo{
			Title: Title{Romaji: titleText},
			Preview: Preview{
				URL:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
				Thumbnail: buildYouTubeThumbnail(videoID),
			},
		}

		metadataText := strings.TrimSpace(videoElement.Find("div div").Last().Text())
		if separatorIndex := strings.Index(metadataText, " - "); separatorIndex != -1 {
			musicVideo.Artist = strings.TrimSpace(metadataText[:separatorIndex])
		}

		videos = append(videos, musicVideo)
	})
	return videos
}

func parseAnimeDocument(document *goquery.Document, malID int) Anime {
	pageURL, _ := document.Find(`meta[property="og:url"]`).Attr("content")
	statusText := extractSidebarValue(document, "Status:")
	ratingText := extractSidebarValue(document, "Rating:")
	if ratingText == "None" {
		ratingText = ""
	}

	return Anime{
		MALID:        malID,
		URL:          pageURL,
		Image:        parseAnimeImage(document),
		Title:        parseAnimeTitle(document),
		Type:         Type(extractSidebarValue(document, "Type:")),
		Source:       Source(extractSidebarValue(document, "Source:")),
		Status:       Status(statusText),
		Airing:       statusText == string(StatusAiring),
		Rating:       Rating(ratingText),
		Synopsis:     parseAnimeSynopsis(document),
		Background:   parseAnimeBackground(document),
		Duration:     extractSidebarValue(document, "Duration:"),
		EpisodeCount: parseIntFromText(extractSidebarValue(document, "Episodes:")),
		Premiered:    parseAnimePremiered(document),
		Aired:        parseAnimeAired(document),
		Broadcast:    parseAnimeBroadcast(document),
		Statistics:   parseAnimeStatistics(document),
		Trailer:      parseAnimeTrailer(document),

		Openings: parseAnimeThemeSongs(document, "opnening"),
		Endings:  parseAnimeThemeSongs(document, "ending"),

		Genres:         extractSidebarMALIDsMultiLabel(document, []string{"Genres:", "Genre:"}, genreIDPattern),
		ExplicitGenres: extractSidebarMALIDs(document, "Explicit Genres:", genreIDPattern),
		Themes:         extractSidebarMALIDsMultiLabel(document, []string{"Themes:", "Theme:"}, genreIDPattern),
		Demographics:   extractSidebarMALIDsMultiLabel(document, []string{"Demographics:", "Demographic:"}, genreIDPattern),
		Producers:      extractSidebarMALIDs(document, "Producers:", producerIDPattern),
		Studios:        extractSidebarMALIDs(document, "Studios:", producerIDPattern),
		Licensors:      extractSidebarMALIDs(document, "Licensors:", producerIDPattern),

		External:  parseAnimeExternalLinks(document),
		Streaming: parseAnimeStreamingLinks(document),
	}
}

func fixThemeSongEpisodeRanges(themeSongs []ThemeSong, totalEpisodes int) {
	if len(themeSongs) != 1 || totalEpisodes <= 0 {
		return
	}
	if themeSongs[0].Episodes.Start == 0 && themeSongs[0].Episodes.End == 0 {
		themeSongs[0].Episodes.Start = 1
		themeSongs[0].Episodes.End = totalEpisodes
	}
}

func GetAnimeByMALID(malID int) (*Anime, error) {
	animePageURL := fmt.Sprintf("%s/anime/%d", malBaseURL, malID)
	animeDocument, fetchErr := makeRequest(animePageURL)
	if fetchErr != nil {
		logger.Errorf("MALClient", "Failed to fetch anime page for MAL ID %d: %v", malID, fetchErr)
		return nil, fmt.Errorf("failed to fetch anime page for MAL ID %d: %w", malID, fetchErr)
	}

	logger.Debugf("MALScraper", "Parsing anime page for MAL ID %d", malID)
	anime := parseAnimeDocument(animeDocument, malID)
	logger.Debugf("MALScraper", "Parsed anime page: Title=%q, EpisodeCount=%d", anime.Title.Romaji, anime.EpisodeCount)

	fixThemeSongEpisodeRanges(anime.Openings, anime.EpisodeCount)
	fixThemeSongEpisodeRanges(anime.Endings, anime.EpisodeCount)

	logger.Debugf("MALScraper", "Fetching videos page for MAL ID %d", malID)
	videosPageURL := fmt.Sprintf("%s/anime/%d/_/video", malBaseURL, malID)
	videosDocument, videosFetchErr := makeRequest(videosPageURL)
	if videosFetchErr != nil {
		logger.Warnf("MALClient", "Failed to fetch videos page for MAL ID %d: %v", malID, videosFetchErr)
	} else {
		anime.Videos = parsePromotionalVideos(videosDocument)
		anime.MusicVideos = parseMusicVideos(videosDocument)
		logger.Debugf("MALScraper", "Parsed videos: %d promotional, %d music", len(anime.Videos), len(anime.MusicVideos))
	}

	logger.Debugf("MALScraper", "Fetching episodes for MAL ID %d", malID)
	episodes, episodesFetchErr := GetAnimeEpisodesByMALID(malID)
	if episodesFetchErr != nil {
		logger.Warnf("MALClient", "Failed to fetch episodes for MAL ID %d: %v", malID, episodesFetchErr)
	} else {
		anime.Episodes = episodes
		logger.Debugf("MALScraper", "Fetched %d episodes for MAL ID %d", len(episodes), malID)
	}

	return &anime, nil
}