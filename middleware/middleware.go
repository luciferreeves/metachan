package middleware

import "github.com/gofiber/fiber/v2"

func Initialize(app *fiber.App) {
	app.Use(httpLogger())
	app.Use(request())
}
