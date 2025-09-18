package storage

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MustInitDBFromEnv() *gorm.DB {
	host := envOr("POSTGRES_HOST", "localhost")
	port := envOr("POSTGRES_PORT", "5432")
	user := envOr("POSTGRES_USER", "sca")
	pass := envOr("POSTGRES_PASSWORD", "sca")
	dbname := envOr("POSTGRES_DB", "sca")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, pass, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Println("DB connected")
	return db
}

func MustRunMigrations(db *gorm.DB) {
	// run raw SQL migrations in order
    files := []string{
        "migrations/001_init.sql",
        "migrations/002_constraints.sql",
        "migrations/003_targets_unique_name.sql",
    }
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}
		if err := db.Exec(string(b)).Error; err != nil {
			panic(err)
		}
	}
	log.Println("migrations applied")
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
