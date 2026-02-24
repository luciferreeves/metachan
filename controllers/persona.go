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
	characterID, ok := meta.Request(c).Param("characterId")
	if !ok {
		return BadRequest(c, errors.New("characterId is required"))
	}

	malID, err := strconv.Atoi(characterID)
	if err != nil {
		return BadRequest(c, errors.New("characterId must be a numeric MAL ID"))
	}

	character, err := repositories.GetCharacterByMALID(malID)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(character)
}

func GetAnimePeople(c *fiber.Ctx) error {
	id := meta.Request(c).MustHave().Param("id")
	provider := meta.Request(c).Default("mal").Query("provider")

	switch provider {
	case "mal", "anilist":
	default:
		return BadRequest(c, errors.New("invalid provider"))
	}

	people, err := repositories.GetAnimePeople(enums.MappingType(provider), id)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(people)
}

func GetPerson(c *fiber.Ctx) error {
	personID, ok := meta.Request(c).Param("personId")
	if !ok {
		return BadRequest(c, errors.New("personId is required"))
	}

	malID, err := strconv.Atoi(personID)
	if err != nil {
		return BadRequest(c, errors.New("personId must be a numeric MAL ID"))
	}

	person, err := repositories.GetPerson(malID)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(person)
}
