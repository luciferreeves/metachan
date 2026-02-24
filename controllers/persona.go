package controllers

import (
	"errors"
	"metachan/enums"
	"metachan/repositories"
	"metachan/utils/meta"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetAnimeCharacters(c *fiber.Ctx) error {
	id := meta.Request(c).MustHave().Param("id")
	provider := meta.Request(c).Default("mal").Query("provider")

	switch provider {
	case "mal", "anilist":
	default:
		return BadRequest(c, errors.New("invalid provider"))
	}

	characters, err := repositories.GetAnimeCharacters(enums.MappingType(provider), id)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(characters)
}

func GetAnimeCharacter(c *fiber.Ctx) error {
	id := meta.Request(c).MustHave().Param("id")
	characterID, ok := meta.Request(c).Param("characterId")
	if !ok {
		return BadRequest(c, errors.New("characterId is required"))
	}
	provider := meta.Request(c).Default("mal").Query("provider")

	switch provider {
	case "mal", "anilist":
	default:
		return BadRequest(c, errors.New("invalid provider"))
	}

	malID, err := strconv.Atoi(characterID)
	if err != nil {
		return BadRequest(c, errors.New("characterId must be a numeric MAL ID"))
	}

	character, err := repositories.GetAnimeCharacter(enums.MappingType(provider), id, malID)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(character)
}
