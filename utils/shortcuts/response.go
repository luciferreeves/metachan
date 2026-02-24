package shortcuts

import "github.com/gofiber/fiber/v2"

func Response(ctx *fiber.Ctx, data any) *response {
	return &response{
		ctx:    ctx,
		data:   data,
		status: fiber.StatusOK,
	}
}

func (r *response) As(status int) error {
	r.status = status
	return r.ctx.Status(status).JSON(r.data)
}
