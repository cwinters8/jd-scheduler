package users

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type User struct {
	Name   string
	Email  string
	Status Status
	Type   Type
}

const redisKey = "Users"

// get users by type
func (t Type) GetUsers(ctx context.Context, rdb *redis.Client) ([]User, error) {
	users, err := GetUsers(ctx, rdb)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	var filteredUsers []User
	for _, u := range users {
		if u.Type == t {
			filteredUsers = append(filteredUsers, u)
		}
	}
	return filteredUsers, nil
}

func NewUser(name string, email string, status Status, userType Type) (*User, error) {
	var err error
	if status >= endStatus || status < UndefinedStatus {
		err = fmt.Errorf("invalid status provided")
	}
	if userType >= endType || userType < UndefinedType {
		msg := "invalid type provided"
		if err != nil {
			err = fmt.Errorf("%w; %s", err, msg)
		} else {
			err = fmt.Errorf(msg)
		}
	}
	if err != nil {
		return nil, err
	}
	if status == UndefinedStatus {
		status = PendingStatus
	}
	return &User{
		Name:   name,
		Email:  email,
		Status: status,
	}, nil
}

func GetUsers(ctx context.Context, rdb *redis.Client) ([]User, error) {
	var users []User
	if err := rdb.LRange(ctx, redisKey, 0, -1).ScanSlice(users); err != nil {
		return users, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
}

// if user was found, returns a 0 or greater int64
// if user was not found, returns a negative value (currently -1 specifically)
func (u *User) GetIndexByEmail(ctx context.Context, rdb *redis.Client) (int64, error) {
	// get users
	users, err := GetUsers(ctx, rdb)
	if err != nil {
		return 0, fmt.Errorf("failed to get users: %w", err)
	}

	// iterate through users and find the first one with the matching email
	for idx, user := range users {
		if user.Email == u.Email {
			return int64(idx), nil
		}
	}

	return -1, nil
}

// TODO: need a way to update a user's email address without adding a new user

// updates the user if the user could be found in the list by email.
// if the user cannot be found, the user is added instead
func (u *User) Update(ctx context.Context, rdb *redis.Client) error {
	idx, err := u.GetIndexByEmail(ctx, rdb)
	if err != nil {
		return fmt.Errorf("failed to get index for user %q: %w", u.Email, err)
	}
	// if the returned index is negative, add the user instead of updating
	if idx < 0 {
		return rdb.RPush(ctx, redisKey, u).Err()
	}
	return rdb.LSet(ctx, redisKey, idx, u).Err()
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

func getType(s string) Type {
	for i := RecruitType; i < endType; i++ {
		if i.String() == s {
			return i
		}
	}
	return UndefinedType
}

func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
