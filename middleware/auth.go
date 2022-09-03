package middleware

import (
	"fmt"
	"net/http"

	"scheduler/stytch"
	"scheduler/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func NewAuthHandler(store *session.Store, client *stytch.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return RenderGetSessionError(c, err)
		}
		sessToken, ok := sess.Get("session_token").(string)
		if !ok || len(sessToken) < 1 {
			return utils.RenderError(c, http.StatusUnauthorized, fmt.Errorf("unable to authenticate: session token not found"))
		}
		// validate session token
		user, err := client.AuthenticateSession(sessToken)
		if err != nil {
			return utils.RenderError(c, http.StatusUnauthorized, err)
		}
		c.Locals("user", *user)
		fmt.Printf("successfully authenticated session for user %s\n", user.UserID)
		return c.Next()
	}
}

func RenderGetSessionError(ctx *fiber.Ctx, err error) error {
	return utils.RenderError(ctx, http.StatusInternalServerError, fmt.Errorf("failed to get session store: %w", err))
}
