package main

import (
	"flag"
	"log"

	"sca/sca/internal/server"
	"sca/sca/internal/storage"
)

// @title Spy Cat Agency APIgo mod vendor
// @version 1.0
// @description CRUD API for Spy Cats, Missions and Targets.
// @BasePath /api/v1
// @schemes http

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "run migrations and exit")
	flag.Parse()

	db := storage.MustInitDBFromEnv()
	if *migrateOnly {
		storage.MustRunMigrations(db)
		return
	}

	r := server.Router(db)
	log.Println("listening on :8080")
	r.Run(":8080")
}
