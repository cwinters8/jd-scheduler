package stytch

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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

type User struct {
	stytch.User
	Roles []Role
}

func NewUser(ctx context.Context, stytchUser *stytch.User, storage *redis.Storage) (*User, error) {
	user := User{*stytchUser, []Role{}}
	// get roles from db
	if err := user.GetRoles(ctx, storage); err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	return &user, nil
}

func (c *Client) AuthenticateSession(ctx context.Context, sessionToken string, storage *redis.Storage) (*User, error) {
	resp, err := c.api.Sessions.Authenticate(&stytch.SessionsAuthenticateParams{
		SessionToken:           sessionToken,
		SessionDurationMinutes: 60,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to authenticate session token: %w", err)
	}
	return NewUser(ctx, &resp.User, storage)
}

func (c *Client) RevokeSession(sessionToken string) error {
	if _, err := c.api.Sessions.Revoke(&stytch.SessionsRevokeParams{
		SessionToken: sessionToken,
	}); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

type Role int

// additional roles need to be between `undefined` and `end`
// roles should be in order of least to highest privilege
const (
	undefined Role = iota
	Recruit
	Volunteer
	Admin
	end
)

func (u *User) GetRoles(ctx context.Context, storage *redis.Storage) error {
	rdb := storage.Conn()
	rawRoles, err := rdb.SMembers(ctx, fmt.Sprintf(UserRoleKeyFormat, u.UserID)).Result()
	if err != nil {
		return fmt.Errorf("failed to get roles from redis: %w", err)
	}
	roles, err := parseRoles(rawRoles)
	if err != nil {
		return fmt.Errorf("failed to parse roles: %w", err)
	}
	u.Roles = roles
	return nil
}

func (u *User) HasRole(ctx context.Context, role Role, storage *redis.Storage) (bool, error) {
	return storage.Conn().SIsMember(ctx, fmt.Sprintf(UserRoleKeyFormat, u.UserID), role).Result()
}

const UserRoleKeyFormat = "user:%s:roles"

func (r Role) String() string {
	switch r {
	case Recruit:
		return "recruit"
	case Volunteer:
		return "volunteer"
	case Admin:
		return "admin"
	default:
		return ""
	}
}

func parseRoles(rawRoles []string) ([]Role, error) {
	var roles []Role
	for _, r := range rawRoles {
		roleNum, err := strconv.Atoi(r)
		if err != nil {
			return nil, fmt.Errorf("failed to convert raw role string to int")
		}
		role := Role(roleNum)
		if role > undefined && role < end {
			roles = append(roles, role)
		}
	}
	return roles, nil
}
