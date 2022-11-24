package calendar

import (
	"context"
	"fmt"

	"scheduler/settings"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Service calendar.Service

// `svc.Acl.Insert()` will add users to a calendar
// Example:
// rule := calendar.AclRule{
// 	Role: "writer", // `writer` allows read & write
// 	Scope: &calendar.AclRuleScope{
// 		Type: "user",
// 		Value: "user_email_address",
// 	},
// }
// svc.Acl.Insert(c.Id, &rule).Do()
// if the calendar should show up for them in their google calendar views,
// probably need to figure out how to use this as well:
// https://developers.google.com/calendar/api/v3/reference/calendarList/insert

// TODO: authenticate the http client
// get it from oauth2.Config.Client + stytch access & refresh tokens
// for oauth client example, see google/auth.go:64 on the begin-stytch-config branch
func NewService(ctx context.Context, accessToken string, refreshToken string, gcpCredsJSON []byte) (*Service, error) {
	token := oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	// check if token is valid
	if !token.Valid() {
		return nil, fmt.Errorf("invalid oauth token")
	}
	// scope user-facing description:
	// Make secondary Google calendars, and see, create, change, and delete events on them
	calendarCreatedScope := "https://www.googleapis.com/auth/calendar.app.created"
	cfg, err := google.ConfigFromJSON(gcpCredsJSON, calendarCreatedScope)
	if err != nil {
		return nil, fmt.Errorf("failed to generate oauth config: %w", err)
	}
	cal, err := calendar.NewService(ctx, option.WithHTTPClient(cfg.Client(ctx, &token)))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize calendar service: %w", err)
	}
	svc := Service(*cal)
	return &svc, nil
}

type Calendar calendar.Calendar

const calendarIDKey = "calendar_id"

// in addition to creating a new google calendar,
// calling this method will overwrite the existing calendar ID,
// which is stored in redis under the key indicated by the `calendarIDKey` constant
func (c *Calendar) Create(ctx context.Context, svc *Service, db *redis.Client) error {
	// TODO: check if calendar exists before attempting to insert
	cal := calendar.Calendar(*c)
	newCal, err := svc.Calendars.Insert(&cal).Do()
	if err != nil {
		return fmt.Errorf("failed to insert calendar: %w", err)
	}
	c.Id = newCal.Id
	// store calendar ID setting
	if err := settings.New(calendarIDKey, c.Id).Save(ctx, db); err != nil {
		return fmt.Errorf("failed to save calendar ID setting: %w", err)
	}
	return nil
}
