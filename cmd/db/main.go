package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

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
		if err := Init(ctx, *drop, conn); err != nil {
			log.Fatalf("failed to initialize db: %s", err.Error())
		}
	} else {
		fmt.Println("no flags passed. nothing to do 🤷‍♂️")
	}
	fmt.Println("done")
}
