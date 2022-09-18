package main

import (
	"context"
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
	"scheduler/users"
	"scheduler/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
	"github.com/gofiber/template/html"
	"github.com/jackc/pgx/v4/pgxpool"
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

	// db config
	pool, err := pgxpool.Connect(context.Background(), os.Getenv("DSN"))
	if err != nil {
		return fmt.Errorf("failed to establish pgx pool: %w", err)
	}
	defer pool.Close()

	cfg := middleware.NewAppConfig(store, stytchClient, mailClient, storage, pool)

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
			fmt.Println(fmt.Errorf("failed to revoke stytch session: %w", err))
		}
		// destroy store session
		if err := sess.Destroy(); err != nil {
			fmt.Println(fmt.Errorf("failed to destroy session: %w", err))
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
		sessToken, err := stytchClient.AuthenticateOauth(c.Query("token"), currentSessToken, func(stytchID string) error {
			ctx := c.Context()
			user, err := users.GetUserByStytchID(ctx, stytchID, pool)
			if err != nil {
				return fmt.Errorf("failed to retrieve user: %w", err)
			}
			// could probably check for nil user pointer in GetUserByStytchID instead of in every place its called...
			if user == nil {
				return fmt.Errorf("user not found")
			}
			switch user.Status {
			case users.DeletedStatus, users.UndefinedStatus:
				return fmt.Errorf("invalid user status")
			case users.PendingStatus, users.InvitedStatus, users.InactiveStatus:
				user.Status = users.ActiveStatus
				if err := user.Update(ctx, pool); err != nil {
					err = fmt.Errorf("failed to update status for user with stytch ID %q: %w", user.StytchID, err)
					fmt.Println(err)
					return err
				}
			}
			return nil
		})
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
	authedHandler := func(tmpl string, getArgs func(ctx *fiber.Ctx) (fiber.Map, error)) fiber.Handler {
		return func(c *fiber.Ctx) error {
			args, err := getArgs(c)
			if err != nil {
				return utils.RenderError(c, http.StatusInternalServerError, fmt.Errorf("error in getArgs func: %w", err))
			}
			args["LoggedIn"] = true
			return c.Render(tmpl, args)
		}
	}

	// TODO NEXT: restrict allowed users to those invited/present in the database with a valid status

	// authenticated routes ⬇️
	app.Get("/dash", func(c *fiber.Ctx) error {
		return authedHandler("dash", func(ctx *fiber.Ctx) (fiber.Map, error) {
			return fiber.Map{
				"Message": "You made it! 🎉",
			}, nil
		})(c)
	})

	// admin portal
	admin := app.Group("/admin", middleware.NewRoleValidator(stytch.Admin))
	admin.Get("/", func(c *fiber.Ctx) error {
		return authedHandler("admin", func(ctx *fiber.Ctx) (fiber.Map, error) {
			return fiber.Map{
				"Message": "Well done, you're an admin! 👨‍💼",
			}, nil
		})(c)
	})

	admin.Get("/volunteers", func(c *fiber.Ctx) error {
		return authedHandler("volunteers", func(ctx *fiber.Ctx) (fiber.Map, error) {
			vols, err := users.GetAllVolunteers(ctx.Context(), pool)
			if err != nil {
				return fiber.Map{}, fmt.Errorf("failed to get volunteers: %w", err)
			}
			return fiber.Map{
				"Volunteers": vols,
			}, nil
		})(c)
	})
	admin.Post("/volunteer", func(c *fiber.Ctx) error {
		// create volunteer & invite
		// don't have stytch ID at this point, so passing an empty string
		volunteer, err := users.NewVolunteer(c.FormValue("name"), c.FormValue("email"), "", users.PendingStatus)
		if err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, err)
		}

		// TODO: someday this should be handled async as it causes a fairly long delay before the browser gets a response
		if err := volunteer.Invite(c.Context(), os.Getenv("SERVER_ADDRESS"), mailClient, engine, pool, stytchClient); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, err)
		}

		return c.Redirect("/admin/volunteers")
	})

	return app.Listen(":3000")
}

func main() {
	if err := setup(); err != nil {
		log.Fatal("failed to setup app: " + err.Error())
	}
}
