package users

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type User struct {
	ID       int
	Name     string
	Email    string
	StytchID string
	Status   Status
	Type     Type
}

// get users by type
func (t Type) GetUsers(ctx context.Context, pool *pgxpool.Pool) ([]*User, error) {
	var users []*User
	if err := pgxscan.Select(ctx, pool, &users, "select * from users where type=$1", t); err != nil {
		return nil, fmt.Errorf("failed to get users from db: %w", err)
	}
	return users, nil
}

func (u *User) IsValid() error {
	var err error
	if u.Status >= endStatus || u.Status < UndefinedStatus {
		err = fmt.Errorf("invalid status %d provided", u.Status)
	}
	if u.Type >= endType || u.Type < UndefinedType {
		msg := fmt.Sprintf("invalid type %d provided", u.Type)
		if err != nil {
			err = fmt.Errorf("%w; %s", err, msg)
		} else {
			err = fmt.Errorf(msg)
		}
	}
	return err
}

// get all users
func GetUsers(ctx context.Context, pool *pgxpool.Pool) ([]*User, error) {
	var users []*User
	if err := pgxscan.Select(ctx, pool, &users, "select * from users"); err != nil {
		return nil, fmt.Errorf("failed to get users from db: %w", err)
	}
	return users, nil
}

func GetUserByEmail(ctx context.Context, email string, pool *pgxpool.Pool) (*User, error) {
	var user User
	if err := pgxscan.Get(ctx, pool, &user, "select * from users where email=$1", email); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func GetUserByStytchID(ctx context.Context, stytchID string, pool *pgxpool.Pool) (*User, error) {
	var user User
	if err := pgxscan.Get(ctx, pool, &user, "select * from users where stytch_id=$1", stytchID); err != nil {
		if err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result") {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// updates the user if the user could be found in the list by stytch ID when possible, falling back to email as needed.
// if the user cannot be found, the user is added instead
func (u *User) Update(ctx context.Context, pool *pgxpool.Pool) error {
	// if provided user is invalid, return error
	if err := u.IsValid(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	if u.ID < 1 {
		// try to select existing user
		var (
			user *User
			err  error
		)
		if len(u.StytchID) > 0 {
			user, err = GetUserByStytchID(ctx, u.StytchID, pool)
		} else {
			user, err = GetUserByEmail(ctx, u.Email, pool)
		}
		if err != nil {
			return fmt.Errorf("failed to select existing user: %w", err)
		}

		// if existing user is not found, insert
		if user == nil {
			var id int
			if err := pgxscan.Get(
				ctx,
				pool,
				&id,
				"insert into users(name, email, stytch_id, status, type) values ($1, $2, $3, $4, $5) returning id",
				u.Name,
				u.Email,
				u.StytchID,
				u.Status,
				u.Type,
			); err != nil {
				return fmt.Errorf("failed to insert user: %w", err)
			}
			u.ID = id
			return nil
		}
		u.ID = user.ID
	}

	// stytch ID should never need to be updated, so that field is omitted here
	if _, err := pool.Exec(
		ctx,
		"update users set name = $1, email = $2, status = $3, type = $4 where id = $5",
		u.Name,
		u.Email,
		u.Status,
		u.Type,
		u.ID,
	); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

type Type int

const (
	UndefinedType Type = iota
	RecruitType
	VolunteerType
	AdminType
	// new types should go here so we don't change the int values associated with each type
	endType
)

func (t Type) String() string {
	switch t {
	case RecruitType:
		return "recruit"
	case VolunteerType:
		return "volunteer"
	case AdminType:
		return "admin"
	default:
		return ""
	}
}

func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
