// Package mariadb implements the storage interface using MariaDB.
//
// MariaDB provides a MySQL-compatible relational database that replaces SQLite
// for issue storage. This backend connects to a MariaDB server via the MySQL protocol.
//
// Key differences from SQLite backend:
//   - Connects to MariaDB server via MySQL protocol
//   - Multi-writer support via server mode
//   - No local file-based storage
package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	// Import MySQL driver for MariaDB connections
	_ "github.com/go-sql-driver/mysql"

	"github.com/steveyegge/beads/internal/storage"
)

// MariaDBStore implements the Storage interface using MariaDB
type MariaDBStore struct {
	db       *sql.DB
	dbName   string       // Database name
	closed   atomic.Bool  // Tracks whether Close() has been called
	connStr  string       // Connection string for reconnection
	mu       sync.RWMutex // Protects concurrent access
	readOnly bool         // True if opened in read-only mode
}

// Config holds MariaDB database configuration
type Config struct {
	Host     string // Server host (default: 127.0.0.1)
	Port     int    // Server port (default: 3306)
	User     string // MySQL user (default: root)
	Password string // MySQL password (default: empty, can be set via BEADS_MARIADB_PASSWORD)
	Database string // Database name (default: beads)
	ReadOnly bool   // Open in read-only mode (skip schema init)
}

// DefaultPort is the default MariaDB port
const DefaultPort = 3306

// Server retry configuration.
// go-sql-driver/mysql doesn't have built-in retry. We add retry for transient
// connection errors (stale pool connections, brief network issues, server restarts).
const serverRetryMaxElapsed = 30 * time.Second

func newServerRetryBackoff() backoff.BackOff {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = serverRetryMaxElapsed
	return bo
}

// isRetryableError returns true if the error is a transient connection error
// that should be retried in server mode.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// MySQL driver transient errors
	if strings.Contains(errStr, "driver: bad connection") {
		return true
	}
	if strings.Contains(errStr, "invalid connection") {
		return true
	}
	// Network transient errors (brief blips, not persistent failures)
	if strings.Contains(errStr, "broken pipe") {
		return true
	}
	if strings.Contains(errStr, "connection reset") {
		return true
	}
	// Don't retry "connection refused" - that means server is down
	return false
}

// withRetry executes an operation with retry for transient errors.
func (s *MariaDBStore) withRetry(ctx context.Context, op func() error) error {
	bo := newServerRetryBackoff()
	return backoff.Retry(func() error {
		err := op()
		if err != nil && isRetryableError(err) {
			return err // Retryable - backoff will retry
		}
		if err != nil {
			return backoff.Permanent(err) // Non-retryable - stop immediately
		}
		return nil
	}, backoff.WithContext(bo, ctx))
}

// New creates a new MariaDB storage backend
func New(ctx context.Context, cfg *Config) (*MariaDBStore, error) {
	// Default values
	if cfg.Database == "" {
		cfg.Database = "beads"
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultPort
	}
	if cfg.User == "" {
		cfg.User = "root"
	}
	// Check environment variable for password (more secure than command-line)
	if cfg.Password == "" {
		cfg.Password = os.Getenv("BEADS_MARIADB_PASSWORD")
	}

	// Connect to MariaDB server via MySQL protocol
	db, connStr, err := openServerConnection(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Test connection
	pingCtx := ctx
	if pingCtx == nil {
		pingCtx = context.Background()
	}
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping MariaDB database: %w", err)
	}

	store := &MariaDBStore{
		db:       db,
		dbName:   cfg.Database,
		connStr:  connStr,
		readOnly: cfg.ReadOnly,
	}

	// Initialize schema (idempotent)
	if !cfg.ReadOnly {
		if err := store.initSchema(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize schema: %w", err)
		}
	}

	return store, nil
}

// openServerConnection opens a connection to a MariaDB server via MySQL protocol
func openServerConnection(ctx context.Context, cfg *Config) (*sql.DB, string, error) {
	// DSN format: user:password@tcp(host:port)/database?parseTime=true
	// parseTime=true tells the MySQL driver to parse DATETIME/TIMESTAMP to time.Time
	var connStr string
	if cfg.Password != "" {
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	} else {
		connStr = fmt.Sprintf("%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.User, cfg.Host, cfg.Port, cfg.Database)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open MariaDB server connection: %w", err)
	}

	// Server mode supports multi-writer, configure reasonable pool size
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ensure database exists (may need to create it)
	// First connect without database to create it
	var initConnStr string
	if cfg.Password != "" {
		initConnStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true",
			cfg.User, cfg.Password, cfg.Host, cfg.Port)
	} else {
		initConnStr = fmt.Sprintf("%s@tcp(%s:%d)/?parseTime=true",
			cfg.User, cfg.Host, cfg.Port)
	}
	initDB, err := sql.Open("mysql", initConnStr)
	if err != nil {
		_ = db.Close()
		return nil, "", fmt.Errorf("failed to open init connection: %w", err)
	}
	defer func() { _ = initDB.Close() }()

	_, err = initDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", cfg.Database))
	if err != nil {
		// MariaDB may return error 1007 even with IF NOT EXISTS - ignore if database already exists
		errLower := strings.ToLower(err.Error())
		if !strings.Contains(errLower, "database exists") && !strings.Contains(errLower, "1007") {
			_ = db.Close()
			// Check for connection refused - server likely not running
			if strings.Contains(errLower, "connection refused") || strings.Contains(errLower, "connect: connection refused") {
				return nil, "", fmt.Errorf("failed to connect to MariaDB server at %s:%d: %w\n\nThe MariaDB server may not be running. Try:\n  sudo systemctl start mariadb    # On systemd systems\n  brew services start mariadb     # On macOS with Homebrew",
					cfg.Host, cfg.Port, err)
			}
			return nil, "", fmt.Errorf("failed to create database: %w", err)
		}
		// Database already exists - that's fine, continue
	}

	return db, connStr, nil
}

// initSchema creates all tables if they don't exist
func initSchemaOnDB(ctx context.Context, db *sql.DB) error {
	// Execute schema creation - split into individual statements
	// because MySQL/MariaDB doesn't support multiple statements in one Exec
	for _, stmt := range splitStatements(schema) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		// Skip pure comment-only statements, but execute statements that start with comments
		if isOnlyComments(stmt) {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to create schema: %w\nStatement: %s", err, truncateForError(stmt))
		}
	}

	// Insert default config values
	for _, stmt := range splitStatements(defaultConfig) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if isOnlyComments(stmt) {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to insert default config: %w", err)
		}
	}

	// Apply index migrations for existing databases.
	// CREATE TABLE IF NOT EXISTS won't add new indexes to existing tables.
	indexMigrations := []string{
		"CREATE INDEX idx_issues_issue_type ON issues(issue_type)",
	}
	for _, migration := range indexMigrations {
		_, err := db.ExecContext(ctx, migration)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") &&
			!strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return fmt.Errorf("failed to apply index migration: %w", err)
		}
	}

	// Remove FK constraint on depends_on_id to allow external references.
	// See SQLite migration 025_remove_depends_on_fk.go for design context.
	// This is idempotent - DROP FOREIGN KEY fails silently if constraint doesn't exist.
	_, err := db.ExecContext(ctx, "ALTER TABLE dependencies DROP FOREIGN KEY fk_dep_depends_on")
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "can't drop") &&
		!strings.Contains(strings.ToLower(err.Error()), "doesn't exist") &&
		!strings.Contains(strings.ToLower(err.Error()), "check that it exists") &&
		!strings.Contains(strings.ToLower(err.Error()), "was not found") {
		return fmt.Errorf("failed to drop fk_dep_depends_on: %w", err)
	}

	// Create views
	if _, err := db.ExecContext(ctx, readyIssuesView); err != nil {
		return fmt.Errorf("failed to create ready_issues view: %w", err)
	}
	if _, err := db.ExecContext(ctx, blockedIssuesView); err != nil {
		return fmt.Errorf("failed to create blocked_issues view: %w", err)
	}

	// Run schema migrations for existing databases
	if err := RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run mariadb migrations: %w", err)
	}

	return nil
}

func (s *MariaDBStore) initSchema(ctx context.Context) error {
	return initSchemaOnDB(ctx, s.db)
}

// splitStatements splits a SQL script into individual statements
func splitStatements(script string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(script); i++ {
		c := script[i]

		if inString {
			current.WriteByte(c)
			if c == stringChar && (i == 0 || script[i-1] != '\\') {
				inString = false
			}
			continue
		}

		if c == '\'' || c == '"' || c == '`' {
			inString = true
			stringChar = c
			current.WriteByte(c)
			continue
		}

		if c == ';' {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			continue
		}

		current.WriteByte(c)
	}

	// Handle last statement without semicolon
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

// truncateForError truncates a string for use in error messages
func truncateForError(s string) string {
	if len(s) > 100 {
		return s[:100] + "..."
	}
	return s
}

// isOnlyComments returns true if the statement contains only SQL comments
func isOnlyComments(stmt string) bool {
	lines := strings.Split(stmt, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		// Found a non-comment, non-empty line
		return false
	}
	return true
}

// Close closes the database connection
func (s *MariaDBStore) Close() error {
	s.closed.Store(true)
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	if s.db != nil {
		if cerr := s.db.Close(); cerr != nil {
			if !errors.Is(cerr, context.Canceled) {
				err = errors.Join(err, cerr)
			}
		}
	}
	s.db = nil
	return err
}

// Path returns the database name (for daemon validation compatibility)
func (s *MariaDBStore) Path() string {
	return s.dbName
}

// IsClosed returns true if Close() has been called
func (s *MariaDBStore) IsClosed() bool {
	return s.closed.Load()
}

// UnderlyingDB returns the underlying *sql.DB connection
func (s *MariaDBStore) UnderlyingDB() *sql.DB {
	return s.db
}

// UnderlyingConn returns a connection from the pool
func (s *MariaDBStore) UnderlyingConn(ctx context.Context) (*sql.Conn, error) {
	return s.db.Conn(ctx)
}

// Ensure MariaDBStore implements storage.Storage
var _ storage.Storage = (*MariaDBStore)(nil)
