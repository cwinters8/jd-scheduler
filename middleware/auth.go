package middleware

import (
	"fmt"
	"net/http"

	"scheduler/stytch"
	"scheduler/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// if redirectOnError is true, when an error occurs the handler will:
// - set an auth_error session value, which can optionally be provided to the user
// - redirect to /login instead of rendering the error page
func NewAuthHandler(cfg *AppConfig, redirectOnError bool) fiber.Handler {
	errorHandler := func(ctx *fiber.Ctx, sess *session.Session, err error, statusCode int) error {
		if redirectOnError {
			sess.Set("auth_error", err)
			sess.Save() // not handling the save error here because it is not crucial that auth_error be available
			return ctx.Redirect("/login")
		}
		return utils.RenderError(ctx, statusCode, err)
	}
	return func(c *fiber.Ctx) error {
		sess, err := cfg.SessionStore.Get(c)
		if err != nil {
			return errorHandler(c, sess, err, http.StatusInternalServerError)
		}
		sessToken, ok := sess.Get("session_token").(string)
		if !ok || len(sessToken) < 1 {
			return errorHandler(c, sess, fmt.Errorf("unable to authenticate: session token not found"), http.StatusUnauthorized)
		}
		// validate session token
		user, err := cfg.AuthClient.AuthenticateSession(c.Context(), sessToken, cfg.Storage)
		if err != nil {
			return errorHandler(c, sess, err, http.StatusUnauthorized)
		}
		c.Locals("user", *user)
		fmt.Printf("successfully authenticated session for user %s\n", user.UserID)
		return c.Next()
	}
}

// check if user has correct roles to access the next route
func NewRoleValidator(expectedRole stytch.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(stytch.User)
		if !ok {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("unable to retrieve user local value"))
		}
		for _, r := range user.Roles {
			if r == expectedRole {
				return c.Next()
			}
		}
		return utils.RenderError(
			c,
			http.StatusForbidden,
			fmt.Errorf("user with ID %q is not allowed to access %s resources", user.UserID, expectedRole.String()),
		)
	}
}
