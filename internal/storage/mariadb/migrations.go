package mariadb

import (
	"database/sql"
	"fmt"
	"strings"
)

// Migration represents a single schema migration for MariaDB.
type Migration struct {
	Name string
	Func func(*sql.DB) error
}

// migrationsList is the ordered list of all MariaDB schema migrations.
// Each migration must be idempotent - safe to run multiple times.
// New migrations should be appended to the end of this list.
var migrationsList = []Migration{
	{"wisp_type_column", migrateWispTypeColumn},
	{"spec_id_column", migrateSpecIDColumn},
}

// RunMigrations executes all registered MariaDB migrations in order.
// Each migration is idempotent and checks whether its changes have
// already been applied before making modifications.
func RunMigrations(db *sql.DB) error {
	for _, m := range migrationsList {
		if err := m.Func(db); err != nil {
			return fmt.Errorf("mariadb migration %q failed: %w", m.Name, err)
		}
	}
	return nil
}

// ListMigrations returns the names of all registered migrations.
func ListMigrations() []string {
	names := make([]string, len(migrationsList))
	for i, m := range migrationsList {
		names[i] = m.Name
	}
	return names
}

// migrateWispTypeColumn adds the wisp_type column if it doesn't exist
func migrateWispTypeColumn(db *sql.DB) error {
	// Check if column exists
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = 'issues' 
		AND column_name = 'wisp_type'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("checking wisp_type column: %w", err)
	}
	if count > 0 {
		return nil // Column already exists
	}

	_, err = db.Exec("ALTER TABLE issues ADD COLUMN wisp_type VARCHAR(32) DEFAULT ''")
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("adding wisp_type column: %w", err)
	}
	return nil
}

// migrateSpecIDColumn adds the spec_id column if it doesn't exist
func migrateSpecIDColumn(db *sql.DB) error {
	// Check if column exists
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = 'issues' 
		AND column_name = 'spec_id'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("checking spec_id column: %w", err)
	}
	if count > 0 {
		return nil // Column already exists
	}

	_, err = db.Exec("ALTER TABLE issues ADD COLUMN spec_id VARCHAR(1024)")
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("adding spec_id column: %w", err)
	}
	
	// Add index for spec_id
	_, err = db.Exec("CREATE INDEX idx_issues_spec_id ON issues(spec_id)")
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") && 
		!strings.Contains(strings.ToLower(err.Error()), "already exists") {
		return fmt.Errorf("creating spec_id index: %w", err)
	}
	return nil
}

