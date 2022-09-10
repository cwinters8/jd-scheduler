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
func NewAuthHandler(store *session.Store, client *stytch.Client, redirectOnError bool) fiber.Handler {
	errorHandler := func(ctx *fiber.Ctx, sess *session.Session, err error, statusCode int) error {
		if redirectOnError {
			sess.Set("auth_error", err)
			sess.Save() // not handling the save error here because it is not crucial that auth_error be available
			return ctx.Redirect(fmt.Sprintf("/login?redirect=%s", ctx.Path()))
		}
		return utils.RenderError(ctx, statusCode, err)
	}
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return errorHandler(c, sess, err, http.StatusInternalServerError)
		}
		sessToken, ok := sess.Get("session_token").(string)
		if !ok || len(sessToken) < 1 {
			return errorHandler(c, sess, fmt.Errorf("unable to authenticate: session token not found"), http.StatusUnauthorized)
		}
		// validate session token
		user, err := client.AuthenticateSession(sessToken)
		if err != nil {
			return errorHandler(c, sess, err, http.StatusUnauthorized)
		}
		c.Locals("user", *user)
		fmt.Printf("successfully authenticated session for user %s\n", user.UserID)
		return c.Next()
	}
}
