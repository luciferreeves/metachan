package controllers

import (
	"metachan/utils/shortcuts"

	"github.com/gofiber/fiber/v2"
)

func BadRequest(c *fiber.Ctx, err error) error {
	return shortcuts.Response(c, fiber.Map{
		"error": err.Error(),
	}).As(fiber.StatusBadRequest)
}

func Unauthorized(c *fiber.Ctx, err error) error {
	return shortcuts.Response(c, fiber.Map{
		"error": err.Error(),
	}).As(fiber.StatusUnauthorized)
}

func Forbidden(c *fiber.Ctx, err error) error {
	return shortcuts.Response(c, fiber.Map{
		"error": err.Error(),
	}).As(fiber.StatusForbidden)
}

func NotFound(c *fiber.Ctx, err error) error {
	return shortcuts.Response(c, fiber.Map{
		"error": err.Error(),
	}).As(fiber.StatusNotFound)
}

func InternalServerError(c *fiber.Ctx, err error) error {
	return shortcuts.Response(c, fiber.Map{
		"error": "Internal Server Error",
	}).As(fiber.StatusInternalServerError)
}

func DefaultError(c *fiber.Ctx, err error) error {
	return shortcuts.Response(c, fiber.Map{
		"error": err.Error(),
	}).As(fiber.StatusInternalServerError)
}
