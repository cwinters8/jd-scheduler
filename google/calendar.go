package google

import (
	"context"
	"errors"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func NewCalendar(ctx context.Context, title string) (*calendar.Calendar, error) {
	// TODO: figure out if I can reuse a service
	svc, err := calendar.NewService(ctx, option.WithHTTPClient(nil))
	if err != nil {
		return nil, errors.New("failed to create calendar service: " + err.Error())
	}
	cal := calendar.Calendar{
		Summary: title,
	}
	return svc.Calendars.Insert(&cal).Do()
}

type CalendarRequest struct {
	Title string `form:"title"`
}
