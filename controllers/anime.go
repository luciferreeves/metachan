package controllers

import (
	"errors"
	"metachan/utils/mal"
	"metachan/utils/meta"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetAnime(c *fiber.Ctx) error {
	idString := meta.Request(c).MustHave().Param("id")

	malID, parseErr := strconv.Atoi(idString)
	if parseErr != nil {
		return BadRequest(c, errors.New("invalid MAL ID"))
	}

	anime, fetchErr := mal.GetAnimeByMALID(malID)
	if fetchErr != nil {
		return NotFound(c, fetchErr)
	}

	return c.JSON(anime)
}
