package utils

import (
	"github.com/gofiber/fiber/v2"
)

func RenderError(ctx *fiber.Ctx, errCode int, msg string) error {
	return ctx.Status(errCode).Render("error", fiber.Map{
		"Error": msg,
	})
}
