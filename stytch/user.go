package stytch

import (
	"encoding/json"
	"fmt"

	"github.com/stytchauth/stytch-go/v5/stytch"
)

// on success, returns a string representing the stytch user ID
func (c *Client) CreateUser(email string) (string, error) {
	// check if user already exists before creating
	search, err := c.api.Users.Search(&stytch.UsersSearchParams{
		Query: &stytch.UsersSearchQuery{
			Operator: stytch.UserSearchOperatorAND,
			Operands: []json.Marshaler{
				stytch.UsersSearchQueryEmailAddressFilter{
					EmailAddresses: []string{email},
				},
			},
		},
	})
	if err == nil && len(search.Results) > 0 {
		return search.Results[0].UserID, nil
	}

	resp, err := c.api.Users.Create(&stytch.UsersCreateParams{
		Email:               email,
		CreateUserAsPending: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create user in stytch: %w", err)
	}
	return resp.UserID, nil
}
