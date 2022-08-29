package main

import (

	// "database/sql"
	"errors"
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
		Views: engine,
	})
	storage := redis.New()
	store := session.New(session.Config{
		Expiration: 24 * time.Hour,
		CookiePath: "/",
		// CookieSecure:   true,
		CookieHTTPOnly: true,
		// CookieSameSite: "Strict",
		Storage: storage,
	})
	app.Use(logger.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{}, "layouts/main")
	})
	app.Get("/session/set", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, "failed to get session store: "+err.Error())
		}
		sess.Set("value", "a random value to test session store")
		if err := sess.Save(); err != nil {
			return utils.RenderError(c, http.StatusInternalServerError, "failed to save session: "+err.Error())
		}
		return c.Render("test", fiber.Map{
			"Text": "should be storing a session value",
		}, "layouts/main")
	})

	// TODO NEXT: test the auth mechanism

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
		return c.Render("calendar", fiber.Map{}, "layouts/main")
	})

	return app.Listen(":3000")
}

func main() {
	if err := setup(); err != nil {
		log.Fatal("failed to setup app: " + err.Error())
	}
}

// func getDatabase() (*sql.DB, error) {
// 	return sql.Open("mysql", os.Getenv("DSN"))
// }
