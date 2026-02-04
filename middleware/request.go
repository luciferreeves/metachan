package middleware

import (
	"metachan/utils/meta"

	"github.com/gofiber/fiber/v2"
)

const requestKey = "__request_ctx"

func request() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := meta.BuildRequest(c)
		c.Locals(requestKey, req)
		return c.Next()
	}
}
