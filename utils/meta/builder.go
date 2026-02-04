package meta

import (
	"metachan/types"

	"github.com/gofiber/fiber/v2"
)

func BuildRequest(c *fiber.Ctx) types.Request {
	return types.Request{
		Path:        c.Path(),
		Method:      c.Method(),
		Query:       buildQueryParams(c),
		Params:      buildRouteParams(c),
		Headers:     buildHeaders(c),
		QueryString: string(c.Request().URI().QueryString()),
		IP:          c.IP(),
		URL:         c.OriginalURL(),
	}
}
