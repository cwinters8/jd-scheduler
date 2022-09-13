package middleware

import (
	"scheduler/mail"
	"scheduler/stytch"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
)

type AppConfig struct {
	SessionStore *session.Store
	AuthClient   *stytch.Client
	MailClient   *mail.Client
	Storage      *redis.Storage
}

func NewAppConfig(store *session.Store, authClient *stytch.Client, mailClient *mail.Client, storage *redis.Storage) *AppConfig {
	return &AppConfig{
		SessionStore: store,
		AuthClient:   authClient,
		MailClient:   mailClient,
		Storage:      storage,
	}
}
