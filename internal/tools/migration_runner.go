package tools

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"log"
	"path/filepath"
	"runtime"
)

func MigrateUp(migrationsPath, migrationsTableName string, pool *pgxpool.Pool) error {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return errors.New("failed to get caller information")
	}

	mainDir := filepath.Dir(filename)
	absolutePath := filepath.Join(mainDir, migrationsPath)
	log.Println("Attempting to apply migrations from", absolutePath)

	sqlDB := stdlib.OpenDBFromPool(pool)
	//defer sqlDB.Close()
	//derived from pool, which lifetime is managed by main app code

	driver, err := pgx.WithInstance(sqlDB, &pgx.Config{
		MigrationsTable: migrationsTableName,
	})
	if err != nil {
		log.Fatalf("Failed to craete migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+absolutePath,
		"pgx",
		driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Applied migration: %d, Dirty: %t\n", version, dirty)

	return nil
}
