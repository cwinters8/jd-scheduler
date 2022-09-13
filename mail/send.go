package mail

import (
	"bytes"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Email mail.Email

type Client struct {
	SendGrid *sendgrid.Client
	From     *Email
}

func NewEmail(name string, address string) Email {
	return Email{
		Name:    name,
		Address: address,
	}
}

func NewClient(apiKey string, from Email) *Client {
	client := sendgrid.NewSendClient(apiKey)
	return &Client{
		SendGrid: client,
		From:     &from,
	}
}

func (e Email) Send(subject string, msg string, client *Client, engine *html.Engine) error {
	from := mail.NewEmail(client.From.Name, client.From.Address)
	var buf bytes.Buffer
	if err := engine.Render(&buf, "email", fiber.Map{
		"Body": msg,
	}, "layouts/email"); err != nil {
		return fmt.Errorf("failed to render email: %w", err)
	}
	email := mail.NewSingleEmail(from, subject, mail.NewEmail(e.Name, e.Address), msg, buf.String())
	if _, err := client.SendGrid.Send(email); err != nil {
		return fmt.Errorf("failed to send mail to %s: %w", e.Address, err)
	}
	return nil
}
