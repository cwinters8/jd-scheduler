package middleware

import (
	"scheduler/mail"
	"scheduler/stytch"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
	"github.com/jackc/pgx/v4/pgxpool"
)

// TODO: move config stuff into its own module

type AppConfig struct {
	SessionStore *session.Store
	AuthClient   *stytch.Client
	MailClient   *mail.Client
	Storage      *redis.Storage
	PGXPool      *pgxpool.Pool
}

func NewAppConfig(
	store *session.Store,
	authClient *stytch.Client,
	mailClient *mail.Client,
	storage *redis.Storage,
	pgxPool *pgxpool.Pool,
) *AppConfig {
	return &AppConfig{
		SessionStore: store,
		AuthClient:   authClient,
		MailClient:   mailClient,
		Storage:      storage,
		PGXPool:      pgxPool,
	}
}
