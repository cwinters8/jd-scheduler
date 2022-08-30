package google

import (
	"fmt"
	"net/http"

	"scheduler/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/oauth2"
)

type AuthConfig struct {
	Filter   fiber.Handler
	Store    *session.Store
	OauthCfg *oauth2.Config
}

func (cfg AuthConfig) NewAuthHandler() (fiber.Handler, error) {
	cfg.Store.RegisterType(oauth2.Token{})

	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, "failed to get session from store: "+err.Error())
		}
		possibleToken := sess.Get("token")
		if token, ok := possibleToken.(oauth2.Token); ok && token.Valid() {
			// c.Locals("client", cfg.OauthCfg.Client(c.Context(), &token))
			return c.Next()
		}
		// set origin path value in session
		sess.Set("origin_path", c.Path())
		sessID := sess.ID()
		if sessID == "" {
			return utils.RenderError(c, http.StatusInternalServerError, "session ID is empty before save")
		}
		if err := sess.Save(); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, "failed to save session: "+err.Error())
		}

		if sessID == "" {
			return utils.RenderError(c, http.StatusInternalServerError, "session ID is empty after save")
		}
		// using session ID for state token
		authURL := cfg.OauthCfg.AuthCodeURL(sessID, oauth2.AccessTypeOffline)
		return c.Redirect(authURL)
	}, nil
}

func (cfg AuthConfig) CallbackHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, "failed to get session from store: "+err.Error())
		}
		// check state value
		sessID := sess.ID()
		if state := c.FormValue("state"); state != sessID {
			fmt.Printf("invalid oauth state. expected '%s', got '%s'\n", sessID, state)
			return utils.RenderError(c, http.StatusUnauthorized, "invalid oauth state")
		}

		// get code and exchange it for token
		token, err := cfg.OauthCfg.Exchange(c.Context(), c.FormValue("code"))
		if err != nil {
			return utils.RenderError(c, http.StatusUnauthorized, "oauth exchange failed: "+err.Error())
		}

		// store token in session
		sess.Set("token", token)

		// grab the origin path before saving session
		possibleOrigin := sess.Get("origin_path")

		if err := sess.Save(); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, "failed to save session: "+err.Error())
		}

		fmt.Println("possible origin:", possibleOrigin)
		if origin, ok := possibleOrigin.(string); ok {
			return c.Redirect(origin)
		}
		return c.Redirect("/")
	}
}
