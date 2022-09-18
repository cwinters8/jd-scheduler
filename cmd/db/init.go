package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func Init(ctx context.Context, withDrop bool, dbConn *pgx.Conn) error {
	if withDrop {
		if _, err := dbConn.Exec(ctx, "drop table users cascade"); err != nil {
			fmt.Println(fmt.Errorf("failed to drop tables: %w", err))
			// not aborting on error dropping table
		}
	}
	if _, err := dbConn.Exec(ctx, `create table users (
		id serial primary key,
		name text null,
		email text not null,
		stytch_id text null,
		status int not null,
		type int not null
	)`); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	return nil
}
