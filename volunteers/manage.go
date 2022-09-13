package volunteers

import (
	"context"
	"encoding/json"
	"fmt"

	"scheduler/mail"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/template/html"
)

const VolunteerKey = "volunteers"

type Volunteer struct {
	Name   string
	Email  string
	Status Status
}

// get volunteers
func GetAllVolunteers(ctx context.Context, rdb *redis.Client) ([]Volunteer, error) {
	var volunteers []Volunteer
	if err := rdb.LRange(ctx, VolunteerKey, 0, -1).ScanSlice(volunteers); err != nil {
		return volunteers, fmt.Errorf("failed to retrieve volunteers: %w", err)
	}
	return volunteers, nil
}

// create volunteer

// creates a new instance of a volunteer struct
func NewVolunteer(ctx context.Context, name string, email string, status Status) Volunteer {
	if status == undefined {
		status = Pending
	}
	return Volunteer{
		Name:   name,
		Email:  email,
		Status: status,
	}
}

func (v Volunteer) Invite(ctx context.Context, mailClient *mail.Client, engine *html.Engine, rdb *redis.Client) error {
	// add to redis
	if err := rdb.RPush(ctx, VolunteerKey, v).Err(); err != nil {
		return fmt.Errorf("failed to add volunteer to list: %w", err)
	}

	// send invitation
	if err := mail.NewEmail(v.Name, v.Email).Send(
		"Justice Democrats Scheduler Invitation",
		"TODO - decide where to specify this",
		mailClient,
		engine,
	); err != nil {
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	// update status
	v.Status = Invited
	return v.Update(ctx, rdb)
}

// use this after updating one or more fields in v to persist the changes
func (v *Volunteer) Update(ctx context.Context, rdb *redis.Client) error {
	idx, err := v.GetIndexByEmail(ctx, rdb)
	if err != nil {
		return fmt.Errorf("failed to get index for volunteer: %w", err)
	}
	return rdb.LSet(ctx, VolunteerKey, idx, v).Err()
}

func (v *Volunteer) GetIndexByEmail(ctx context.Context, rdb *redis.Client) (int64, error) {
	volunteers, err := GetAllVolunteers(ctx, rdb)
	if err != nil {
		return 0, fmt.Errorf("failed to get volunteers: %w", err)
	}
	for idx, val := range volunteers {
		if val.Email == v.Email {
			return int64(idx), nil
		}
	}
	return 0, fmt.Errorf("failed to find index for email %q", v.Email)
}

func (v *Volunteer) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}

func (v *Volunteer) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, v)
}

type Status int

const (
	undefined Status = iota
	Pending
	Invited
	Active
	Inactive
	Deleted
	end
)

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Invited:
		return "invited"
	case Active:
		return "active"
	case Inactive:
		return "inactive"
	case Deleted:
		return "deleted"
	default:
		return ""
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
