package anime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"metachan/types"
	"net/http"
	"strings"
)

func getAnimeViaAnilist(anilistID int) (*types.AnilistAnimeResponse, error) {
	graphQLQuery := fmt.Sprintf(`query {
									Media(id: %d) {
										id
										idMal
										title {
										romaji
										english
										native
										userPreferred
										}
										type
										format
										status
										description
										startDate {
										year
										month
										day
										}
										endDate {
										year
										month
										day
										}
										season
										seasonYear
										episodes
										duration
										chapters
										volumes
										countryOfOrigin
										isLicensed
										source
										hashtag
										trailer {
										id
										site
										thumbnail
										}
										updatedAt
										coverImage {
										extraLarge
										large
										medium
										color
										}
										bannerImage
										genres
										synonyms
										averageScore
										meanScore
										popularity
										isLocked
										trending
										favourites
										tags {
										id
										name
										description
										category
										rank
										isGeneralSpoiler
										isMediaSpoiler
										isAdult
										}
										relations {
										edges {
											id
											relationType
											node {
											id
											title {
												romaji
												english
												native
												userPreferred
											}
											format
											type
											status
											coverImage {
												extraLarge
												large
												medium
												color
											}
											bannerImage
											}
										}
										}
										characters {
										edges {
											role
											node {
											id
											name {
												first
												last
												middle
												full
												native
												userPreferred
											}
											image {
												large
												medium
											}
											description
											age
											}
										}
										}
										staff {
										edges {
											role
											node {
											id
											name {
												first
												middle
												last
												full
												native
												userPreferred
											}
											image {
												large
												medium
											}
											description
											primaryOccupations
											gender
											age
											languageV2
											}
										}
										}
										studios {
										edges {
											isMain
											node {
											id
											name
											}
										}
										}
										isAdult
										nextAiringEpisode {
										id
										airingAt
										timeUntilAiring
										episode
										}
										airingSchedule {
										nodes {
											id
											episode
											airingAt
											timeUntilAiring
										}
										}
										trends {
										nodes {
											date
											trending
											popularity
											inProgress
										}
										}
										externalLinks {
										id
										url
										site
										}
										streamingEpisodes {
										title
										thumbnail
										url
										site
										}
										rankings {
										id
										rank
										type
										format
										year
										season
										allTime
										context
										}
										stats {
										scoreDistribution {
											score
											amount
										}
										statusDistribution {
											status
											amount
										}
										}
										siteUrl
									}
								}`, anilistID)

	// Remove debug print that can cause issues with large queries
	// fmt.Printf("GraphQL Query: %s\n", graphQLQuery)

	apiURL := "https://graphql.anilist.co"

	// Escape quotes in the query to make valid JSON
	escapedQuery := strings.Replace(graphQLQuery, `"`, `\"`, -1)
	escapedQuery = strings.Replace(escapedQuery, "\n", " ", -1)
	escapedQuery = strings.Replace(escapedQuery, "\t", "", -1)

	// Create the JSON payload
	jsonData := []byte(fmt.Sprintf(`{"query": "%s"}`, escapedQuery))

	// Create a request with the proper body
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for better debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get anime data: %s - %s", resp.Status, string(bodyBytes))
	}

	var anilistResponse types.AnilistAnimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&anilistResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if anilistResponse.Data.Media.ID == 0 {
		return nil, fmt.Errorf("no data found for Anilist ID %d", anilistID)
	}
	return &anilistResponse, nil
}
