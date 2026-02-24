package controllers

import (
	"errors"
	"metachan/enums"
	"metachan/repositories"
	"metachan/utils/meta"

	"github.com/gofiber/fiber/v2"
)

func GetAnime(c *fiber.Ctx) error {
	id := meta.Request(c).MustHave().Param("id")
	provider := meta.Request(c).Default("mal").Query("provider")

	switch provider {
	case "mal", "anilist":
	default:
		return BadRequest(c, errors.New("invalid provider"))
	}

	anime, err := repositories.GetAnime(enums.MappingType(provider), id)
	if err != nil {
		return NotFound(c, err)
	}

	return c.JSON(anime)
}
