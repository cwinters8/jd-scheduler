package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"scheduler/stytch"
	"scheduler/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html"
	"github.com/joho/godotenv"
	"github.com/stytchauth/stytch-go/v5/stytch/config"
	"golang.org/x/oauth2"
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

	// stytch config
	stytchClient, err := stytch.NewClient(
		config.EnvTest,
		os.Getenv("STYTCH_CLIENT_ID"),
		os.Getenv("STYTCH_SECRET"),
	)
	if err != nil {
		return err
	}

	app.Use(favicon.New(favicon.Config{
		File: "./assets/favicon.ico",
	}))
	app.Use(logger.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"LoggedIn": false,
		})
	})

	app.Get("/oauth", func(c *fiber.Ctx) error {
		_, err := stytchClient.AuthenticateOauth(c.Params("token"))
		if err != nil {
			return utils.RenderError(c, http.StatusUnauthorized, "failed to authenticate oauth token: "+err.Error())
		}
		// TODO: store session token for later use
		return nil
	})

	return app.Listen(":3000")
}

func main() {
	if err := setup(); err != nil {
		log.Fatal("failed to setup app: " + err.Error())
	}
}

func renderGetSessionError(ctx *fiber.Ctx, err error) error {
	return utils.RenderError(ctx, http.StatusInternalServerError, "failed to get session store: "+err.Error())
}

func getToken(sess *session.Session) *oauth2.Token {
	possibleToken := sess.Get("token")
	token, ok := possibleToken.(oauth2.Token)
	if !ok || !token.Valid() {
		return nil
	}
	return &token
}

func isLoggedIn(sess *session.Session) bool {
	possibleToken := sess.Get("token")
	token, ok := possibleToken.(oauth2.Token)
	if !ok || !token.Valid() {
		return false
	}
	return true
}
