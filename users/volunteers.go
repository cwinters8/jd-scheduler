package users

import (
	"context"
	"fmt"

	"scheduler/mail"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/template/html"
)

// get volunteers
func GetAllVolunteers(ctx context.Context, rdb *redis.Client) ([]User, error) {
	return VolunteerType.GetUsers(ctx, rdb)
}

// create volunteer

// creates a new instance of a volunteer struct
func NewVolunteer(name string, email string, status Status) (*User, error) {
	return NewUser(name, email, status, VolunteerType)
}

func (v *User) Invite(ctx context.Context, mailClient *mail.Client, engine *html.Engine, rdb *redis.Client) error {
	// add to redis
	if err := v.Update(ctx, rdb); err != nil {
		return fmt.Errorf("failed to add volunteer to users list: %w", err)
	}

	// send invitation
	if err := mail.NewEmail(v.Name, v.Email).Send(
		"Justice Democrats Scheduler Invitation",
		"TODO - decide where to specify this", // TODO
		mailClient,
		engine,
	); err != nil {
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	// update status
	v.Status = InvitedStatus
	return v.Update(ctx, rdb)
}
