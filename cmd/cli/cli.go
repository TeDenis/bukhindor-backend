package main

import (
	"database/sql"
	"log"

	"github.com/TeDenis/bukhindor-backend/internal/config"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "bukhindor-cli",
		Short: "CLI инструмент для управления Bukhindor Backend",
	}

	// Добавляем команды
	rootCmd.AddCommand(migrateCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func migrateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "migrate",
		Short: "Управление миграциями базы данных",
	}

	cmd.AddCommand(migrateUpCmd())
	cmd.AddCommand(migrateDownCmd())
	cmd.AddCommand(migrateStatusCmd())

	return cmd
}

func migrateUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Применить все миграции",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.New()
			return runMigrations(cfg, "up")
		},
	}
}

func migrateDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Откатить последнюю миграцию",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.New()
			return runMigrations(cfg, "down")
		},
	}
}

func migrateStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Показать статус миграций",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.New()
			return runMigrations(cfg, "status")
		},
	}
}

func runMigrations(cfg *config.Config, command string) error {
	// Подключаемся к базе данных PostgreSQL через PgPool
	dsn := cfg.GetPgPoolDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Настраиваем goose
	goose.SetDialect("postgres")

	// Выполняем команду
	switch command {
	case "up":
		log.Println("Applying migrations...")
		if err := goose.Up(db, "deployments/postgres/migrations"); err != nil {
			return err
		}
		log.Println("Migrations applied successfully")
	case "down":
		log.Println("Rolling back last migration...")
		if err := goose.Down(db, "deployments/postgres/migrations"); err != nil {
			return err
		}
		log.Println("Migration rolled back successfully")
	case "status":
		log.Println("Migration status:")
		if err := goose.Status(db, "deployments/postgres/migrations"); err != nil {
			return err
		}
	}

	return nil
}
