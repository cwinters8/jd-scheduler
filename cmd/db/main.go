package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

// TODO: could use cli args or a (cue?) config file to pass values instead of os.Getenv

func main() {
	init := flag.Bool("init", false, "initializes the required tables")
	drop := flag.Bool("drop", false, "used in combination with the init flag. will cause existing tables to be dropped prior to initialization")
	flag.Parse()

	if err := godotenv.Load(); err != nil && !strings.Contains(err.Error(), "no such file") {
		log.Fatalf("failed to load .env: %s", err.Error())
	}

	if *init {
		fmt.Println("attempting to initialize database...")
		ctx := context.Background()
		conn, err := pgx.Connect(ctx, os.Getenv("DSN"))
		if err != nil {
			log.Fatalf("failed to connect to db: %s", err.Error())
		}
		// initialize db
		isProd, err := strconv.ParseBool(os.Getenv("PROD"))
		if err != nil {
			fmt.Printf("failed to parse PROD env variable as bool: %s\ndefaulting to false\n", err.Error())
			isProd = false
		}
		adminName := os.Getenv("ADMIN_NAME")
		adminEmail := os.Getenv("ADMIN_EMAIL")
		e, err := mail.ParseAddress(fmt.Sprintf("%s <%s>", adminName, adminEmail))
		if err != nil {
			log.Fatalf("failed to parse admin email address: %s", err.Error())
		}
		if err := Init(ctx, *drop, isProd, e.Name, e.Address, conn); err != nil {
			log.Fatalf("failed to initialize db: %s", err.Error())
		}
	} else {
		fmt.Println("no flags passed. nothing to do 🤷‍♂️")
	}
	fmt.Println("done")
}
