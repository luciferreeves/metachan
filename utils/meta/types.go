package meta

import (
	"metachan/types"

	"github.com/gofiber/fiber/v2"
)

type facade struct {
	req types.Request
	ctx *fiber.Ctx
}

type required struct {
	req types.Request
	ctx *fiber.Ctx
}

type withDefault struct {
	req types.Request
	ctx *fiber.Ctx
	def string
}
