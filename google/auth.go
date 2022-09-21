package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type AuthError struct {
	Error    error
	Redirect bool
}

func Authenticate(cfg *oauth2.Config) error {

	return nil
}

func GetAuthConfig(credsFilePath string) (*oauth2.Config, error) {
	creds, err := os.ReadFile(credsFilePath)
	if err != nil {
		return nil, errors.New("failed to read credentials file: " + err.Error())
	}
	authCfg, err := google.ConfigFromJSON(creds, calendar.CalendarScope)
	if err != nil {
		return nil, errors.New("failed to build config from credentials: " + err.Error())
	}
	return authCfg, nil
}

func GetClient(ctx context.Context, scope string, credsFilePath string) (*http.Client, *AuthError) {
	creds, err := os.ReadFile(credsFilePath)
	if err != nil {
		return nil, &AuthError{
			Error:    errors.New("failed to read credentials file: " + err.Error()),
			Redirect: false,
		}
	}
	cfg, err := google.ConfigFromJSON(creds, scope)
	if err != nil {
		return nil, &AuthError{
			Error:    errors.New("failed to build config from credentials: " + err.Error()),
			Redirect: false,
		}
	}
	// TODO: figure out how to store tokens uniquely and keep track of them (by user??)
	tokenFile := "token.json"
	// try to get the token by the contents of the file
	// if that fails, redirect the user to the auth URL
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		return nil, &AuthError{
			Error:    errors.New("failed to get valid token: " + err.Error()),
			Redirect: true,
		}
	}
	return cfg.Client(ctx, token), nil // cfg.Client glues the oauth2.Token to the http client
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", file, err)
	}
	defer f.Close()
	var token oauth2.Token
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token from file %s: %w", file, err)
	}
	if !token.Valid() {
		return nil, fmt.Errorf("token from file %s is invalid", file)
	}
	return &token, nil
}
