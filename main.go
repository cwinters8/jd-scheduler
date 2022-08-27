package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
)

func setup() error {
	engine := html.New("templates", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Use(logger.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{}, "layouts/main")
	})
	return app.Listen(":3000")
}

func main() {
	err := setup()
	if err != nil {
		log.Fatal("failed to setup app: " + err.Error())
	}
}
