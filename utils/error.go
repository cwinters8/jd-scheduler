package utils

import (
	"github.com/gofiber/fiber/v2"
)

func RenderError(ctx *fiber.Ctx, errCode int, err error) error {
	return ctx.Status(errCode).Render("error", fiber.Map{
		"Error": err,
	})
}
