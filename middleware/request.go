package middleware

import (
	"metachan/utils/meta"

	"github.com/gofiber/fiber/v2"
)

const requestKey = "__request_ctx"

func request() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(requestKey, meta.Request(c))
		return c.Next()
	}
}
