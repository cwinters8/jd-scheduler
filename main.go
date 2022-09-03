package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"scheduler/middleware"
	"scheduler/stytch"
	"scheduler/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
	"github.com/gofiber/template/html"
	"github.com/joho/godotenv"
	"github.com/stytchauth/stytch-go/v5/stytch/config"
)

func setup() error {
	if err := godotenv.Load(); err != nil && !strings.Contains(err.Error(), "no such file") {
		return errors.New("failed to load .env: " + err.Error())
	}

	engine := html.New("templates", ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})
	storage := redis.New()
	store := session.New(session.Config{
		Expiration:     24 * time.Hour,
		CookiePath:     "/",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		Storage:        storage,
	})

	// stytch config
	stytchClient, err := stytch.NewClient(
		config.EnvTest,
		os.Getenv("STYTCH_CLIENT_ID"),
		os.Getenv("STYTCH_SECRET"),
	)
	if err != nil {
		return err
	}
	googleLoginURL := fmt.Sprintf(
		"%s?public_token=%s",
		os.Getenv("GOOGLE_OAUTH_START"),
		os.Getenv("STYTCH_PUBLIC_TOKEN"),
	)

	app.Use(favicon.New(favicon.Config{
		File: "./assets/favicon.ico",
	}))
	app.Use(logger.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			// how can I inexpensively find out if the user is logged in? 🤔
			"LoggedIn": false,
		})
	})
	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{
			"GoogleLoginURL": googleLoginURL,
		})
	})
	app.Get("/logout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return renderGetSessionError(c, err)
		}
		sessToken, _ := sess.Get("session_token").(string)
		// revoke stytch session
		if err := stytchClient.RevokeSession(sessToken); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("failed to revoke stytch session: %w", err))
		}
		// destroy store session
		if err := sess.Destroy(); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("failed to destroy session: %w", err))
		}
		return c.Redirect("/")
	})

	app.Get("/oauth", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return renderGetSessionError(c, err)
		}
		// try to get an existing session token from the store
		currentSessToken, _ := sess.Get("session_token").(string)
		// authenticate
		sessToken, err := stytchClient.AuthenticateOauth(c.Query("token"), currentSessToken)
		if err != nil {
			return utils.RenderError(c, http.StatusUnauthorized, fmt.Errorf("failed to authenticate oauth token: %w", err))
		}
		// store session token for later use
		sess.Set("session_token", sessToken)
		// try getting a redirect path from the store
		redirect, _ := sess.Get("auth_redirect").(string)
		// save the session
		if err := sess.Save(); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("failed to save session: %w", err))
		}
		// go to either the redirect path or authenticated dashboard
		if redirect == "" {
			redirect = "/dash"
		}
		return c.Redirect(redirect)
	})

	app.Use(middleware.NewAuthHandler(store, stytchClient))
	// authenticated routes ⬇️
	app.Get("/dash", func(c *fiber.Ctx) error {
		return c.Render("dash", fiber.Map{
			"Message":  "You made it! 🎉",
			"LoggedIn": true,
		})
	})

	return app.Listen(":3000")
}

func main() {
	if err := setup(); err != nil {
		log.Fatal("failed to setup app: " + err.Error())
	}
}

// TODO: replace this with utils.RenderGetSessionError
func renderGetSessionError(ctx *fiber.Ctx, err error) error {
	return utils.RenderError(ctx, http.StatusInternalServerError, fmt.Errorf("failed to get session store: %w", err))
}
