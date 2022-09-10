package main

import (
	"bytes"
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
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))

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
			"LoggedIn": false,
		})
	})
	app.Get("/login", func(c *fiber.Ctx) error {
		redirect := c.Query("redirect")
		if len(redirect) > 0 {
			if sess, err := store.Get(c); err == nil {
				sess.Set("auth_redirect", redirect)
				sess.Save()
			}
		}
		return c.Render("login", fiber.Map{
			"GoogleLoginURL": googleLoginURL,
		})
	})
	app.Get("/logout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return utils.RenderGetSessionError(c, err)
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
			return utils.RenderGetSessionError(c, err)
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
		sess.Delete("auth_redirect")
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

	app.Use(middleware.NewAuthHandler(store, stytchClient, true))
	// authenticated routes ⬇️
	app.Get("/dash", func(c *fiber.Ctx) error {
		return c.Render("dash", fiber.Map{
			"Message":  "You made it! 🎉",
			"LoggedIn": true,
		})
	})
	// TODO: PLAYGROUND START get rid of these
	app.Get("/redis/list/append", func(c *fiber.Ctx) error {
		rdb := storage.Conn()
		list := []string{"a", "b", "c"}
		count, err := rdb.RPush(c.Context(), "list_test", list).Result()
		if err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("failed to initialize list and add 'a', 'b', and 'c' to it: %w", err))
		}
		if count != 3 {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("incorrect number of elements pushed: expected 3, got %d", count))
		}
		return c.Render("data", fiber.Map{
			"Data": fmt.Sprintf("successfully initialized redis list and pushed %d elements to it", count),
		})
	})
	app.Get("/redis/list", func(c *fiber.Ctx) error {
		rdb := storage.Conn()
		values, err := rdb.LRange(c.Context(), "list_test", 0, -1).Result()
		if err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("failed to get list items: %w", err))
		}
		return c.Render("data", fiber.Map{
			"Data": fmt.Sprintf("found values: %v", values),
		})
	})
	app.Get("/mail", func(c *fiber.Ctx) error {
		return c.Render("mail", fiber.Map{
			"LoggedIn": true,
		})
	})
	app.Post("/mail", func(c *fiber.Ctx) error {
		subject := c.FormValue("subject")
		message := c.FormValue("message")

		layout := "layouts/email"
		content := fiber.Map{
			"Title": "Hello from Scheduler 🚀",
			"Body":  message,
		}
		from := mail.NewEmail("Scheduler Admin", "clark@orionsbelt.dev") // TODO: read from env variables maybe
		to := mail.NewEmail(c.FormValue("to-name"), c.FormValue("to-email"))
		var buf bytes.Buffer
		if err := engine.Render(&buf, "email", content, layout); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("failed to render email: %w", err))
		}
		email := mail.NewSingleEmail(from, subject, to, message, buf.String())

		_, err := client.Send(email)
		if err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("failed to send mail to %s: %w", to.Address, err))
		}
		content["Name"] = to.Name
		content["Email"] = to.Address
		content["LoggedIn"] = true
		return c.Render("mail_success", content)
	})
	// PLAYGROUND END get rid of these

	return app.Listen(":3000")
}

func main() {
	if err := setup(); err != nil {
		log.Fatal("failed to setup app: " + err.Error())
	}
}
