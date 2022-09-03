package utils

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func RenderError(ctx *fiber.Ctx, statusCode int, err error) error {
	return ctx.Status(statusCode).Render("error", fiber.Map{
		"Error": err,
	})
}

func RenderGetSessionError(ctx *fiber.Ctx, err error) error {
	return RenderError(ctx, http.StatusInternalServerError, fmt.Errorf("failed to get session store: %w", err))
}
