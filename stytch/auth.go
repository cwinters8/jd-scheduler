package stytch

import (
	"errors"

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
func (c *Client) AuthenticateOauth(token string) (string, error) {
	if len(token) < 1 {
		return "", errors.New("empty token")
	}
	resp, err := c.api.OAuth.Authenticate(&stytch.OAuthAuthenticateParams{
		Token:                  token,
		SessionDurationMinutes: 60,
	})
	if err != nil {
		return "", errors.New("unable to authenticate: " + err.Error())
	}
	return resp.SessionToken, nil
}
