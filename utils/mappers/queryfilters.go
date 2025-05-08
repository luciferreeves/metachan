package mappers

import "strings"

// SpecialAnimeIDMappings maps specific anime titles to their correct IDs
var SpecialAnimeIDMappings = map[string]string{
	"one piece": "ReooPAxPMsHM4KPMY",
}

// GetSpecialAnimeID checks if an anime title has a special ID mapping
// Returns the special ID and true if found, empty string and false otherwise
func GetSpecialAnimeID(title string) (string, bool) {
	uppercaseTitle := strings.ToLower(strings.TrimSpace(title))
	id, exists := SpecialAnimeIDMappings[uppercaseTitle]
	return id, exists
}
