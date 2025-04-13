package migrator

import (
    "errors"
    "flag"
    "fmt"
    "github.com/golang-migrate/migrate/v4"
)

func main() {
    var storagePath, migrationPath, migrationsTable string

    flag.StringVar(&storagePath, "storage", "", "Path to the storage file")
    flag.StringVar(&migrationPath, "migrations", "", "Path to the migrations directory")
    flag.StringVar(&migrationsTable, "table", "migrations", "Name of the migrations table")
    flag.Parse()

    if storagePath == "" {
        panic("Storage path is required")
    }
    if migrationPath == "" {
        panic("Migrations path is required")
    }
    if migrationsTable == "" {
        panic("Migrations table name is required")
    }

    m, err := migrate.New(
        "file://"+migrationPath,
        fmt.Sprintf("sqlite://%s?x-migrations-table=%s", storagePath, migrationsTable),
    )
    if err != nil {
        panic(err)
    }

    if err := m.Up(); err != nil {
        if errors.Is(err, migrate.ErrNoChange) {
            fmt.Println("no migrations to apply")

            return
        }

        panic(err)
    }
}
