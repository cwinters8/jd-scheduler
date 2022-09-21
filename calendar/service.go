package calendar

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Service calendar.Service

// TODO: authenticate the http client
// get it from oauth2.Config.Client + stytch access & refresh tokens
// for oauth client example, see google/auth.go:64 on the begin-stytch-config branch
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	cal, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize calendar service: %w", err)
	}
	svc := Service(*cal)
	return &svc, nil
}

type Calendar calendar.Calendar

func (c *Calendar) Create(svc *Service) error {
	// TODO: check if calendar exists before attempting to insert
	cal := calendar.Calendar(*c)
	newCal, err := svc.Calendars.Insert(&cal).Do()
	if err != nil {
		return fmt.Errorf("failed to insert calendar: %w", err)
	}
	c.Id = newCal.Id
	return nil
}
