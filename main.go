package main

import (

	// "database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"scheduler/google"
	"scheduler/utils"

	// _ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
	"github.com/gofiber/template/html"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func setup() error {
	if err := godotenv.Load(); err != nil {
		return errors.New("failed to load .env: " + err.Error())
	}
	// db, err := getDatabase()
	// if err != nil {
	// 	return errors.New("failed to connect to database: " + err.Error())
	// }

	engine := html.New("templates", ".html")
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})
	storage := redis.New()
	store := session.New(session.Config{
		Expiration: 24 * time.Hour,
		CookiePath: "/",
		// CookieSecure:   true, // TODO: enable for deployment
		CookieHTTPOnly: true,
		Storage:        storage,
	})
	app.Use(logger.New())
	app.Get("/", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return renderGetSessionError(c, err)
		}
		return c.Render("index", fiber.Map{
			"LoggedIn": isLoggedIn(sess),
		})
	})

	app.Get("/logout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return renderGetSessionError(c, err)
		}
		if err := sess.Destroy(); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, "failed to destroy session: "+err.Error())
		}
		return c.Redirect("/")
	})

	// TODO NEXT: test the auth mechanism end to end

	oauthCfg, err := google.GetAuthConfig(os.Getenv("CREDS_PATH"))
	if err != nil {
		return errors.New("failed to create oauth config: " + err.Error())
	}
	authCfg := google.AuthConfig{
		Store:    store,
		OauthCfg: oauthCfg,
	}
	authHandler, err := authCfg.NewAuthHandler()
	if err != nil {
		return errors.New("failed to configure google authentication: " + err.Error())
	}
	app.Get("/auth/google/callback", authCfg.CallbackHandler())

	cal := app.Group("/calendar")
	cal.Use(authHandler)
	cal.Get("/", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return renderGetSessionError(c, err)
		}
		return c.Render("calendar", fiber.Map{
			"LoggedIn": isLoggedIn(sess),
		})
	})
	cal.Post("/", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return renderGetSessionError(c, err)
		}
		cal, err := google.NewCalendar(c.Context(), authCfg.OauthCfg.Client(
			c.Context(),
			getToken(sess),
		), c.FormValue("new calendar title", "Scheduler"))
		if err != nil {
			return utils.RenderError(
				c,
				http.StatusInternalServerError,
				"failed to create new calendar: "+err.Error(),
			)
		}
		return c.Render("success", fiber.Map{
			"Message": fmt.Sprintf(
				"Successfully created a new calendar with ID %s and title %s 🎉 🗓",
				cal.Id,
				cal.Summary,
			),
		})
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

// func getDatabase() (*sql.DB, error) {
// 	return sql.Open("mysql", os.Getenv("DSN"))
// }
