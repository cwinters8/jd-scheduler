package stytch

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/storage/redis"
	"github.com/stytchauth/stytch-go/v5/stytch"
	"github.com/stytchauth/stytch-go/v5/stytch/config"
	"github.com/stytchauth/stytch-go/v5/stytch/stytchapi"
)

type Client struct {
	api *stytchapi.API
	Env config.Env
}

func NewClient(env config.Env, projectID string, secret string) (*Client, error) {
	api, err := stytchapi.NewAPIClient(env, projectID, secret)
	if err != nil {
		return nil, errors.New("failed to create Stytch client: " + err.Error())
	}
	return &Client{
		api: api,
		Env: env,
	}, nil
}

// on success, returns a session token valid for 60 minutes
func (c *Client) AuthenticateOauth(token string, sessionToken string, validator func(stytchID string) error) (string, error) {
	if len(token) < 1 {
		return "", errors.New("empty token")
	}
	resp, err := c.api.OAuth.Authenticate(&stytch.OAuthAuthenticateParams{
		Token:                  token,
		SessionDurationMinutes: 60,
		SessionToken:           sessionToken,
	})
	if err != nil {
		return "", fmt.Errorf("unable to authenticate oauth token: %w", err)
	}

	if err := validator(resp.UserID); err != nil {
		return "", fmt.Errorf("failed to validate user: %w", err)
	}

	return resp.SessionToken, nil
}

func (c *Client) AuthenticateSession(ctx context.Context, sessionToken string, storage *redis.Storage) (string, error) {
	resp, err := c.api.Sessions.Authenticate(&stytch.SessionsAuthenticateParams{
		SessionToken:           sessionToken,
		SessionDurationMinutes: 60,
	})
	if err != nil {
		return "", fmt.Errorf("unable to authenticate session token: %w", err)
	}
	return resp.User.UserID, nil
}

func (c *Client) RevokeSession(sessionToken string) error {
	if _, err := c.api.Sessions.Revoke(&stytch.SessionsRevokeParams{
		SessionToken: sessionToken,
	}); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}
