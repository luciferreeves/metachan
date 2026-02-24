package controllers

import (
	"errors"
	"metachan/enums"
	"metachan/repositories"
	"metachan/utils/meta"

	"github.com/gofiber/fiber/v2"
)

func GetAnimeEpisodes(c *fiber.Ctx) error {
	id := meta.Request(c).MustHave().Param("id")
	provider := meta.Request(c).Default("mal").Query("provider")

	switch provider {
	case "mal", "anilist":
	default:
		return BadRequest(c, errors.New("invalid provider"))
	}

	episodes, err := repositories.GetAnimeEpisodes(enums.MappingType(provider), id)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(episodes)
}

func GetAnimeEpisode(c *fiber.Ctx) error {
	id := meta.Request(c).MustHave().Param("id")
	episodeID := meta.Request(c).MustHave().Param("episodeId")
	provider := meta.Request(c).Default("mal").Query("provider")

	switch provider {
	case "mal", "anilist":
	default:
		return BadRequest(c, errors.New("invalid provider"))
	}

	episode, err := repositories.GetAnimeEpisode(enums.MappingType(provider), id, episodeID)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(episode)
}
