package main

import (
	"database/sql"
	"fmt"
	"go-sso/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()

	if cfg.PSQL.Migrator == nil {
		panic("migrator config is not set")
	}

	fmt.Println(cfg.PSQL.DSN())

	db, err := sql.Open("postgres", cfg.PSQL.DSN())
	if err != nil {
		panic(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: cfg.PSQL.Migrator.Table,
	})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+cfg.PSQL.Migrator.Path,
		"postgres", driver)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	fmt.Println("Migrations applied successfully.")
}
