package bootstrap

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"spin-transactions/migrations"
)

var DB *sql.DB

func connectDB(migrate bool) error {
	if DB != nil {
		if migrate {
			return runMigrations(DB)
		}
		return nil
	}
	dsn, ok := databaseDSN()
	if !ok {
		log.Println("database not configured; skipping DB connection")
		return nil
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("database ping: %w", err)
	}
	if migrate {
		if err := runMigrations(db); err != nil {
			_ = db.Close()
			return err
		}
	}
	DB = db
	log.Println("database ok")
	return nil
}

func runMigrations(db *sql.DB) error {
	if os.Getenv("RUN_MIGRATIONS") != "true" {
		log.Println("RUN_MIGRATIONS is not true; skipping migrations")
		return nil
	}
	for i, stmt := range splitSQLStatements(migrations.CreateTransactionsSQL) {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration statement %d: %w", i+1, err)
		}
	}
	return nil
}

func splitSQLStatements(sqlText string) []string {
	parts := strings.Split(sqlText, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func databaseDSN() (string, bool) {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn, true
	}
	host, user, password, name := os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME")
	if host == "" || user == "" || password == "" || name == "" {
		return "", false
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   name,
	}
	q := u.Query()
	q.Set("sslmode", "require")
	u.RawQuery = q.Encode()
	return u.String(), true
}
