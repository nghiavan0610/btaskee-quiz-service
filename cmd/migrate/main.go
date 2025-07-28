package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/env"
	"github.com/pressly/goose/v3"
)

const (
	dialect = "postgres"
)

var (
	flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir   = flags.String("dir", "internal/database/migrations", "directory with migration files")
)

func main() {
	env.Load()

	ctx := context.Background()

	flags.Usage = usage
	_ = flags.Parse(os.Args[1:])

	args := flags.Args()

	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	command := args[0]

	// Build database connection string from individual environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSchema := os.Getenv("DB_SCHEMA")
	dbSSLMode := os.Getenv("DB_SSL_MODE")

	if dbUser == "" {
		log.Fatal("DB_USER environment variable not set")
	}
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD environment variable not set")
	}
	if dbHost == "" {
		log.Fatal("DB_HOST environment variable not set")
	}
	if dbPort == "" {
		log.Fatal("DB_PORT environment variable not set")
	}
	if dbName == "" {
		log.Fatal("DB_NAME environment variable not set")
	}
	if dbSchema == "" {
		log.Fatal("DB_SCHEMA environment variable not set")
	}
	if dbSSLMode == "" {
		log.Fatal("DB_SSL_MODE environment variable not set")
	}

	// Construct the connection string
	dbConnect := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?search_path=%s&sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSchema, dbSSLMode)

	db, err := goose.OpenDBWithDriver(dialect, dbConnect)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err.Error())
		}
	}()

	if err := goose.RunContext(ctx, command, db, *dir, args[1:]...); err != nil {
		log.Fatalf("migrate %v: %v", command, err)
	}

}

func usage() {
	fmt.Println(usagePrefix)
	flags.PrintDefaults()
	fmt.Println(usageCommands)
}

var (
	usagePrefix = `Usage: migrate COMMAND
Examples:
    migrate status
`

	usageCommands = `
Commands:
    up                   Migrate the DB to the most recent version available
    up-by-one            Migrate the DB up by 1
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations`
)
