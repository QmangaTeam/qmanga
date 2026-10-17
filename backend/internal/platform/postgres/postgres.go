package postgres

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"qmanga/migrations"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(contextInstance context.Context, sqlString string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(contextInstance context.Context, sqlString string, arguments ...interface{}) (pgx.Rows, error)
	QueryRow(contextInstance context.Context, sqlString string, arguments ...interface{}) pgx.Row
}

type Client struct {
	pool *pgxpool.Pool
}

func New(contextInstance context.Context, databaseURL string) (*Client, error) {
	poolConfig, parseError := pgxpool.ParseConfig(databaseURL)
	if parseError != nil {
		return nil, fmt.Errorf("parsing database url: %w", parseError)
	}

	pool, poolError := pgxpool.NewWithConfig(contextInstance, poolConfig)
	if poolError != nil {
		return nil, fmt.Errorf("creating database connection pool: %w", poolError)
	}

	clientInstance := &Client{pool: pool}
	if pingError := clientInstance.Ping(contextInstance); pingError != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", pingError)
	}

	return clientInstance, nil
}

func (clientInstance *Client) Ping(contextInstance context.Context) error {
	if clientInstance == nil || clientInstance.pool == nil {
		return errors.New("postgres connection is not initialized")
	}
	return clientInstance.pool.Ping(contextInstance)
}

func (clientInstance *Client) Close() {
	if clientInstance != nil && clientInstance.pool != nil {
		clientInstance.pool.Close()
	}
}

func (clientInstance *Client) DB() DBTX {
	if clientInstance == nil {
		return nil
	}
	return clientInstance.pool
}

func (clientInstance *Client) Migrate(contextInstance context.Context) error {
	if clientInstance == nil || clientInstance.pool == nil {
		return errors.New("postgres connection is not initialized")
	}

	conn, acquireError := clientInstance.pool.Acquire(contextInstance)
	if acquireError != nil {
		return fmt.Errorf("acquiring connection for migration: %w", acquireError)
	}

	lockKey := int64(847291)
	_, lockError := conn.Exec(contextInstance, "SELECT pg_advisory_lock($1)", lockKey)
	if lockError != nil {
		conn.Release()
		return fmt.Errorf("acquiring advisory lock: %w", lockError)
	}

	defer func() {
		unlockContext, unlockCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer unlockCancel()
		_, unlockError := conn.Exec(unlockContext, "SELECT pg_advisory_unlock($1)", lockKey)
		if unlockError != nil {
			hijackedConn := conn.Hijack()
			closeContext, cancelClose := context.WithTimeout(context.Background(), time.Second)
			defer cancelClose()
			_ = hijackedConn.Close(closeContext)
			return
		}
		conn.Release()
	}()

	_, createTableError := conn.Exec(contextInstance, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if createTableError != nil {
		return fmt.Errorf("creating schema_migrations table: %w", createTableError)
	}

	entries, readDirError := fs.ReadDir(migrations.FS, ".")
	if readDirError != nil {
		return fmt.Errorf("reading migrations directory: %w", readDirError)
	}

	var migrationFiles []string
	for _, entry := range entries {
		entryName := entry.Name()
		if !entry.IsDir() && len(entryName) > 4 && entryName[len(entryName)-4:] == ".sql" {
			migrationFiles = append(migrationFiles, entryName)
		}
	}
	sort.Strings(migrationFiles)

	for _, filename := range migrationFiles {
		var alreadyApplied bool
		checkStatusError := conn.QueryRow(contextInstance, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", filename).Scan(&alreadyApplied)
		if checkStatusError != nil {
			return fmt.Errorf("checking migration status for %s: %w", filename, checkStatusError)
		}

		if alreadyApplied {
			continue
		}

		contentBytes, readFileError := fs.ReadFile(migrations.FS, filename)
		if readFileError != nil {
			return fmt.Errorf("reading migration file %s: %w", filename, readFileError)
		}

		transaction, beginTxError := conn.BeginTx(contextInstance, pgx.TxOptions{})
		if beginTxError != nil {
			return fmt.Errorf("beginning transaction for %s: %w", filename, beginTxError)
		}

		if _, execMigrationError := transaction.Exec(contextInstance, string(contentBytes)); execMigrationError != nil {
			_ = transaction.Rollback(contextInstance)
			return fmt.Errorf("executing migration %s: %w", filename, execMigrationError)
		}

		if _, recordMigrationError := transaction.Exec(contextInstance, "INSERT INTO schema_migrations (version) VALUES ($1)", filename); recordMigrationError != nil {
			_ = transaction.Rollback(contextInstance)
			return fmt.Errorf("recording migration %s: %w", filename, recordMigrationError)
		}

		if commitError := transaction.Commit(contextInstance); commitError != nil {
			return fmt.Errorf("committing migration %s: %w", filename, commitError)
		}
	}

	return nil
}
