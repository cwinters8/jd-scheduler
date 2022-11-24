package main

import (
	"context"
	"fmt"
	"os"

	"scheduler/stytch"
	"scheduler/users"

	"github.com/jackc/pgx/v4"
	"github.com/stytchauth/stytch-go/v5/stytch/config"
)

func Init(ctx context.Context, withDrop bool, isProd bool, adminName string, adminEmail string, dbConn *pgx.Conn) error {
	if withDrop {
		// TODO: delete existing calendars?
		if _, err := dbConn.Exec(ctx, "drop table users cascade"); err != nil {
			fmt.Println(fmt.Errorf("failed to drop tables: %w", err))
			// not aborting on error dropping table - table may not exist
		}
	}
	if _, err := dbConn.Exec(ctx, `create table users (
		id serial primary key,
		name text null,
		email text not null,
		stytch_id text not null,
		status int not null,
		type int not null
	)`); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	cfg := config.EnvTest
	if isProd {
		cfg = config.EnvLive
	}
	// TODO: pass either the client or the client ID & secret to Init in order to remove the os.Getenv calls
	client, err := stytch.NewClient(
		cfg,
		os.Getenv("STYTCH_CLIENT_ID"),
		os.Getenv("STYTCH_SECRET"),
	)
	if err != nil {
		return fmt.Errorf("failed to create new stytch client: %w", err)
	}
	// add initial admin user
	stytchID, err := client.CreateUser(adminEmail)
	if err != nil {
		return fmt.Errorf("failed to create stytch user: %w", err)
	}

	if _, err := dbConn.Exec(
		ctx,
		"insert into users(name, email, stytch_id, status, type) values ($1, $2, $3, $4, $5)",
		adminName,
		adminEmail,
		stytchID,
		users.InvitedStatus,
		users.AdminType,
	); err != nil {
		return fmt.Errorf("failed to add admin user to db: %w", err)
	}

	// get a connection to redis

	// create calendar

	// allow admin user to access

	return nil
}
