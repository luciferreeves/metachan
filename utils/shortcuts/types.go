package shortcuts

import "github.com/gofiber/fiber/v2"

type response struct {
	ctx    *fiber.Ctx
	data   any
	status int
}
