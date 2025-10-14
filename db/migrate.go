package db

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// migrations holds SQL migration files embedded into the binary.
//
//go:embed migrations/*.sql
var migrations embed.FS

var (
	rxUp   = regexp.MustCompile(`(?s)--\s*\+duckUp\s*(.*?)\n--\s*\+duckDown`)
	rxDown = regexp.MustCompile(`(?s)--\s*\+duckUp.*?--\s*\+duckDown\s*(.*)$`)
)

// MigrateUp applies all pending migrations. Naming ordered by integer prefix (i.e. 00001_sample.sql)
func MigrateUp(db *sqlx.DB) error {
	if err := ensureMigrationTable(db); err != nil {
		return err
	}
	applied, err := loadAppliedVersions(db)
	if err != nil {
		return err
	}
	files, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("db.MigrateUp: read migrations: %w", err)
	}
	type mf struct {
		version int
		name    string
	}
	list := []mf{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ver, ok := parseVersion(f.Name())
		if !ok {
			continue
		}
		list = append(list, mf{ver, f.Name()})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].version < list[j].version })

	for _, m := range list {
		if applied[m.version] {
			continue // already applied
		}
		content, err := fs.ReadFile(migrations, "migrations/"+m.name)
		if err != nil {
			return fmt.Errorf("db.MigrateUp: read %s: %w", m.name, err)
		}
		upSQL, err := extractUp(string(content))
		if err != nil {
			return fmt.Errorf("db.MigrateUp: parse %s: %w", m.name, err)
		}
		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("db.MigrateUp: begin %s: %w", m.name, err)
		}
		if err = execStatements(tx, upSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("db.MigrateUp: apply %s: %w", m.name, err)
		}
		if _, err = tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, m.version, time.Now()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("db.MigrateUp: record %s: %w", m.name, err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("db.MigrateUp: commit %s: %w", m.name, err)
		}
	}
	return nil
}

// MigrateDown rolls back all applied migrations in reverse order.
func MigrateDown(db *sqlx.DB) error {
	applied, err := orderedAppliedVersions(db)
	if err != nil {
		return err
	}
	for i := len(applied) - 1; i >= 0; i-- {
		ver := applied[i]
		name, err := findFileByVersion(ver)
		if err != nil {
			return err
		}
		content, err := fs.ReadFile(migrations, "migrations/"+name)
		if err != nil {
			return fmt.Errorf("db.MigrateDown: read %s: %w", name, err)
		}
		downSQL, err := extractDown(string(content))
		if err != nil {
			return fmt.Errorf("db.MigrateDown: parse %s: %w", name, err)
		}
		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("db.MigrateDown: begin %s: %w", name, err)
		}
		if err = execStatements(tx, downSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("db.MigrateDown: apply %s: %w", name, err)
		}
		if _, err = tx.Exec(`DELETE FROM schema_migrations WHERE version = ?`, ver); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("db.MigrateDown: delete version %d: %w", ver, err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("db.MigrateDown: commit %s: %w", name, err)
		}
	}
	return nil
}

func ensureMigrationTable(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMP NOT NULL)`)
	if err != nil {
		return fmt.Errorf("db.ensureMigrationTable: %w", err)
	}
	return nil
}

func loadAppliedVersions(db *sqlx.DB) (map[int]bool, error) {
	rows, err := db.Queryx(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("db.loadAppliedVersions: %w", err)
	}
	defer rows.Close()
	out := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func orderedAppliedVersions(db *sqlx.DB) ([]int, error) {
	rows, err := db.Queryx(`SELECT version FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, fmt.Errorf("db.orderedAppliedVersions: %w", err)
	}
	defer rows.Close()
	var vs []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	return vs, rows.Err()
}

func parseVersion(name string) (int, bool) {
	// Expect prefix like 0001_...
	for i, r := range name {
		if r < '0' || r > '9' { // stop at first non-digit
			if i == 0 {
				return 0, false
			}
			var v int
			_, err := fmt.Sscanf(name[:i], "%d", &v)
			return v, err == nil
		}
	}
	return 0, false
}

func findFileByVersion(ver int) (string, error) {
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return "", err
	}
	prefix := fmt.Sprintf("%04d", ver)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			return e.Name(), nil
		}
	}
	return "", errors.New("migration file for version not found")
}

func extractUp(content string) (string, error) {
	m := rxUp.FindStringSubmatch(content)
	if len(m) < 2 {
		return "", errors.New("no Up section found")
	}
	return strings.TrimSpace(m[1]), nil
}

func extractDown(content string) (string, error) {
	m := rxDown.FindStringSubmatch(content)
	if len(m) < 2 {
		return "", errors.New("no Down section found")
	}
	return strings.TrimSpace(m[1]), nil
}

// execStatements splits on semicolons and executes non-empty statements.
func execStatements(tx *sqlx.Tx, sqlBlob string) error {
	for _, stmt := range strings.Split(sqlBlob, ";") {
		s := strings.TrimSpace(stmt)
		if s == "" {
			continue
		}
		if _, err := tx.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
