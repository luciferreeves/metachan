package meta

import (
	"metachan/types"

	"github.com/gofiber/fiber/v2"
)

func buildQueryParams(c *fiber.Ctx) []types.HTTPParam {
	params := make([]types.HTTPParam, 0)
	c.Request().URI().QueryArgs().VisitAll(func(k, v []byte) {
		params = append(params, types.HTTPParam{
			Key:   string(k),
			Value: string(v),
		})
	})
	return params
}

func buildRouteParams(c *fiber.Ctx) []types.HTTPParam {
	params := make([]types.HTTPParam, 0)
	for k, v := range c.AllParams() {
		params = append(params, types.HTTPParam{
			Key:   k,
			Value: v,
		})
	}
	return params
}

func buildHeaders(c *fiber.Ctx) []types.HTTPParam {
	params := make([]types.HTTPParam, 0)
	c.Request().Header.VisitAll(func(k, v []byte) {
		params = append(params, types.HTTPParam{
			Key:   string(k),
			Value: string(v),
		})
	})
	return params
}
