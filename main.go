package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"scheduler/mail"
	"scheduler/middleware"
	"scheduler/stytch"
	"scheduler/utils"
	"scheduler/volunteers"

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
	redisPort, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		return fmt.Errorf("failed to parse REDIS_PORT as int: %w", err)
	}
	cert, err := tls.LoadX509KeyPair(os.Getenv("PUBLIC_KEY_PATH"), os.Getenv("PRIVATE_KEY_PATH"))
	if err != nil {
		return fmt.Errorf("failed to load cert keypair: %w", err)
	}
	redisCfg := redis.Config{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     redisPort,
		Password: os.Getenv("REDIS_PASSWORD"),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
	storage := redis.New(redisCfg)
	store := session.New(session.Config{
		Expiration:     24 * time.Hour,
		CookiePath:     "/",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		Storage:        storage,
	})

	mailClient := mail.NewClient(
		os.Getenv("SENDGRID_API_KEY"),
		mail.NewEmail(os.Getenv("EMAIL_FROM_NAME"), os.Getenv("EMAIL_FROM_ADDRESS")),
	)

	// stytch config
	stytchClient, err := stytch.NewClient(
		config.EnvTest,
		os.Getenv("STYTCH_CLIENT_ID"),
		os.Getenv("STYTCH_SECRET"),
	)
	if err != nil {
		return fmt.Errorf("failed to create new stytch client: %w", err)
	}
	cfg := middleware.NewAppConfig(store, stytchClient, mailClient, storage)

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

	app.Use(middleware.NewAuthHandler(cfg, true))
	authedHandler := func(tmpl string, getArgs func() fiber.Map) fiber.Handler {
		return func(c *fiber.Ctx) error {
			args := getArgs()
			args["LoggedIn"] = true
			return c.Render(tmpl, args)
		}
	}
	// authenticated routes ⬇️
	app.Get("/dash", authedHandler("dash", func() fiber.Map {
		return fiber.Map{
			"Message": "You made it! 🎉",
		}
	}))

	// admin portal
	admin := app.Group("/admin", middleware.NewRoleValidator(stytch.Admin))
	admin.Get("/", authedHandler("admin", func() fiber.Map {
		return fiber.Map{
			"Message": "Well done, you're an admin! 👨‍💼",
		}
	}))
	admin.Get("/volunteers", authedHandler("volunteers", func() fiber.Map {
		// TODO NEXT: get volunteers and pass in the map here
		return fiber.Map{}
	}))
	admin.Post("/volunteer", func(c *fiber.Ctx) error {
		// create volunteer & invite
		volunteer := volunteers.NewVolunteer(c.Context(), c.FormValue("name"), c.FormValue("email"), 0)
		if err := volunteer.Invite(c.Context(), mailClient, engine, storage.Conn()); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, err)
		}

		// TODO: test the volunteer creation flow
		return c.Redirect("/volunteers", http.StatusCreated)
	})

	return app.Listen(":3000")
}

func main() {
	if err := setup(); err != nil {
		log.Fatal("failed to setup app: " + err.Error())
	}
}
