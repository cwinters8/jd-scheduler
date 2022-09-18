package users

import (
	"bytes"
	"context"
	"fmt"

	"scheduler/mail"
	"scheduler/stytch"

	"github.com/gofiber/template/html"
	"github.com/jackc/pgx/v4/pgxpool"
)

// get volunteers
func GetAllVolunteers(ctx context.Context, pool *pgxpool.Pool) ([]*User, error) {
	return VolunteerType.GetUsers(ctx, pool)
}

// create volunteer

// creates a new instance of a volunteer struct
func NewVolunteer(name string, email string, stytchID string, status Status) (*User, error) {
	user := User{
		Name:     name,
		Email:    email,
		StytchID: stytchID,
		Status:   status,
		Type:     VolunteerType,
	}
	if err := user.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}
	return &user, nil
}

func (v *User) Invite(
	ctx context.Context,
	serverAddress string,
	mailClient *mail.Client,
	engine *html.Engine,
	pool *pgxpool.Pool,
	stytchClient *stytch.Client,
) error {
	// add to stytch
	id, err := stytchClient.CreateUser(v.Email)
	if err != nil {
		return fmt.Errorf("failed to create stytch user: %w", err)
	}
	v.StytchID = id
	// add to database
	if err := v.Update(ctx, pool); err != nil {
		return fmt.Errorf("failed to add volunteer to users list: %w", err)
	}

	url := fmt.Sprintf("%s/dash", serverAddress)
	// send invitation
	var buf bytes.Buffer
	if err := engine.Render(&buf, "email_invite", map[string]interface{}{
		"URL": url,
	}, "layouts/email"); err != nil {
		return fmt.Errorf("failed to render email: %w", err)
	}
	plaintextMsg := fmt.Sprintf("Please click the following link to accept our invitation to the Justice Democrats Scheduler tool: %s", url)
	if err := mail.NewEmail(v.Name, v.Email).Send(
		"Scheduler Invitation",
		plaintextMsg,
		buf.String(),
		mailClient,
	); err != nil {
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	// update status
	v.Status = InvitedStatus
	return v.Update(ctx, pool)
}
