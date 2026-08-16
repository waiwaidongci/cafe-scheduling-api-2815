package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cafe-scheduling-api/internal/config"

	"github.com/jackc/pgx/v5"
)

func main() {
	migrationsDir := flag.String("migrations", "migrations", "directory containing .sql migration files")
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		log.Fatalf("create schema_migrations: %v", err)
	}

	files, err := migrationFiles(*migrationsDir)
	if err != nil {
		log.Fatalf("read migrations: %v", err)
	}

	for _, file := range files {
		version := filepath.Base(file)
		var exists bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists); err != nil {
			log.Fatalf("check migration %s: %v", version, err)
		}
		if exists {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("read migration %s: %v", file, err)
		}
		if err := applyMigration(ctx, conn, version, string(content)); err != nil {
			log.Fatalf("apply migration %s: %v", version, err)
		}
		log.Printf("applied %s", version)
	}
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(files)
	return files, nil
}

func applyMigration(ctx context.Context, conn *pgx.Conn, version, sql string) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	for _, statement := range splitSQL(sql) {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := tx.Exec(ctx, statement); err != nil {
			return fmt.Errorf("execute statement: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func splitSQL(sql string) []string {
	var statements []string
	var current strings.Builder
	singleQuoted := false
	dollarTag := ""

	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		if dollarTag != "" {
			if strings.HasPrefix(sql[i:], dollarTag) {
				current.WriteString(sql[i : i+len(dollarTag)])
				i += len(dollarTag) - 1
				dollarTag = ""
			} else {
				current.WriteByte(ch)
			}
			continue
		}
		if ch == '\'' {
			current.WriteByte(ch)
			singleQuoted = !singleQuoted
			continue
		}
		if !singleQuoted && ch == '$' {
			if end := strings.IndexByte(sql[i+1:], '$'); end >= 0 {
				tag := sql[i : i+end+2]
				current.WriteString(tag)
				i += len(tag) - 1
				dollarTag = tag
				continue
			}
		}
		if ch == ';' && !singleQuoted && dollarTag == "" {
			statements = append(statements, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}
	if strings.TrimSpace(current.String()) != "" {
		statements = append(statements, current.String())
	}
	return statements
}
