package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

/**
 * Query for migration execution log
 */
const migrationsQuery string = `CREATE TABLE IF NOT EXISTS {{MIGRATIONS_TABLE_NAME}} (
	version VARCHAR(14) NOT NULL,
	migration_name VARCHAR(128) NULL,
	start_time TIMESTAMP NULL DEFAULT NOW(),
	end_time TIMESTAMP NULL,
	PRIMARY KEY (version),
	UNIQUE INDEX version_UNIQUE (version ASC))
ENGINE = InnoDB;`

func NewPool(db *sql.DB) (Pool, error) {
	var pool Pool = Pool{
		tableName:  "migrations",
		db:         db,
		migrations: map[string]Migration{},
		ups:        []string{},
		downs:      []string{},
	}

	err := pool.initialize()

	return pool, err
}

/**
 * Migrations structure
 */
type Pool struct {
	db         *sql.DB
	tableName  string
	migrations map[string]Migration
	ups        []string
	downs      []string
}

func (m *Pool) transaction(fnc func(tx *sql.Tx) error) error {

	ctx := context.Background()
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	if err := fnc(tx); err != nil {
		return err
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

/**
 * Reset migration structure, clear loaded migrations map and slices
 */
func (m *Pool) Reset() {
	for k := range m.migrations {
		delete(m.migrations, k)
	}
	m.ups = []string{}
	m.downs = []string{}
}

/**
 * Transverse a Migrations slice and add no duplicated migrations. If duplicated
 * the function Migrations.Reset() is executed and return an error
 */
func (m *Pool) AddMigrations(migrations []Migration) error {

	versions := map[string]bool{}

	for _, migration := range migrations {
		if _, found := versions[migration.Version]; found {
			m.Reset()
			return errors.New("found duplicated migration version \"" + migration.Version + "\"")
		} else {
			if _, found := m.migrations[migration.Version]; !found {
				migration.Status = StatusDown
				m.migrations[migration.Version] = migration
				if !m.IsVersionUp(migration.Version) {
					m.downs = append(m.downs, migration.Version)
				}
			} else {
				migration.Status = StatusUp
				m.migrations[migration.Version] = migration
			}
			versions[migration.Version] = true
		}
	}
	return nil
}

/**
 * Check if a migration version is already up
 */
func (m *Pool) IsVersionUp(version string) bool {
	for _, m := range m.ups {
		if version == m {
			return true
		}
	}

	return false
}

/**
 * Initialize the Migrations structure.
 *
 * Read executed migrations and add then into Migrations structure
 *
 * If the migrations table is missing, this function create de table to keep an
 * execution log.
 */
func (m *Pool) initialize() error {

	if _, err := m.db.Exec(
		strings.Replace(migrationsQuery, "{{MIGRATIONS_TABLE_NAME}}", m.tableName, 1),
	); err != nil {
		return err
	}

	rows, err := m.db.Query("SELECT version, migration_name, start_time, end_time FROM " + m.tableName + " ORDER BY version ASC")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		migration := Migration{}
		err := rows.Scan(&migration.Version, &migration.Name, &migration.StartTime, &migration.EndTime)
		if err != nil {
			return err
		} else {
			migration.Status = StatusUp
			m.migrations[migration.Version] = migration
			m.ups = append(m.ups, migration.Version)
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return err
}

func (m *Pool) GetMigrationsByStatus(status string) []Migration {

	migrations := []Migration{}

	keys := make([]string, 0, len(m.migrations))
	for k := range m.migrations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if m.migrations[k].Status == status {
			migrations = append(migrations, m.migrations[k])
		}
	}

	return migrations
}

func (m *Pool) GetMigrationStatus(migration *Migration) (string, error) {

	query := "SELECT version, start_time, end_time FROM " + m.tableName + " WHERE version = ? LIMIT 1"

	var version string

	if err := m.db.QueryRow(query, migration.Version).Scan(
		&version,
		&migration.StartTime,
		&migration.EndTime,
	); err != nil {
		migration.Status = StatusDown
		return "", err
	}

	migration.Status = StatusUp

	return migration.Status, nil
}

func (m *Pool) GetMigrations() []Migration {

	migrations := make([]Migration, 0, len(m.migrations))
	keys := make([]string, 0, len(m.migrations))
	for k := range m.migrations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		migrations = append(migrations, m.migrations[k])
	}

	return migrations
}

func (m *Pool) LoadFolder(path string) error {

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		migration, err := FromFile(path + "/" + file.Name())
		if err != nil {
			return err
		}
		if _, found := m.migrations[migration.Version]; found {
			return errors.New("Migration version \"" + migration.Version + "\" is duplicated")
		} else {
			m.migrations[migration.Version] = migration
			if migration.Status == StatusDown {
				m.downs = append(m.downs, migration.Version)
			} else {
				m.ups = append(m.ups, migration.Version)
			}
		}
	}

	sort.Strings(m.ups)
	sort.Strings(m.downs)

	return nil
}

func (m *Pool) HasDownMigrations() bool {
	return len(m.downs) > 0
}

func (m *Pool) HasUpMigrations() bool {
	return len(m.ups) > 0
}

func (m *Pool) migrate(migration *Migration, bridge interface{}) error {
	if migration.Status != StatusDown {
		return errors.New("migration " + migration.Version + " >> Migration already Up")
	}

	if err := m.transaction(func(tx *sql.Tx) error {
		start := time.Now()

		inQuery := "INSERT INTO " + m.tableName + "(version, migration_name, start_time)VALUES(?,?,?);"
		if _, err := tx.Exec(inQuery, migration.Version, migration.Name, start); err != nil {
			return errors.New("migration " + migration.Version + " >> " + err.Error())
		}

		upQuery, err := parseTplSQL(migration.Up, bridge)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(upQuery); err != nil {
			return errors.New("migration " + migration.Version + " >> " + err.Error())
		}

		updateQuery := "UPDATE " + m.tableName + " SET end_time = ? WHERE version = ?;"
		end := time.Now()
		if _, err := tx.Exec(updateQuery, end, migration.Version); err != nil {
			return errors.New("migration " + migration.Version + " >> " + err.Error())
		}

		migration.Status = StatusUp
		migration.StartTime = new(sql.NullTime)
		migration.StartTime.Time = start

		migration.EndTime = new(sql.NullTime)
		migration.EndTime.Time = end

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (m *Pool) Migrate(bridge interface{}, pre func(*Migration), post func(*Migration)) error {

	for _, version := range m.downs {
		migration := m.migrations[version]
		if pre != nil {
			pre(&migration)
		}
		if err := m.migrate(&migration, bridge); err != nil {
			return err
		}
		if post != nil {
			post(&migration)
		}
		m.migrations[migration.Version] = migration
		m.ups = append(m.ups, migration.Version)

	}
	m.downs = []string{}

	return nil
}

func (m *Pool) rollback(migration *Migration, bridge interface{}) error {

	if migration.Status != StatusUp {
		return errors.New("migration " + migration.Version + " >> Migration already Down")
	}
	downQuery, err := parseTplSQL(migration.Down, bridge)
	if err != nil {
		return err
	}

	if err := m.transaction(func(tx *sql.Tx) error {
		if downQuery != "" {
			fmt.Printf("Down Query executed: %v\n", downQuery)
			if _, err := m.db.Exec(downQuery); err != nil {
				return errors.New("migration " + migration.Version + " >> " + err.Error())
			}
		}

		delQuery := "DELETE FROM " + m.tableName + " WHERE version = ?"
		if _, err := m.db.Exec(delQuery, migration.Version); err != nil {
			return errors.New("migration " + migration.Version + " >> " + err.Error())
		}

		migration.Status = StatusDown

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (m *Pool) Rollback(bridge interface{}, pre func(*Migration), post func(*Migration)) error {

	if len(m.ups) > 0 {
		last := m.ups[len(m.ups)-1]
		migration := m.migrations[last]

		if pre != nil {
			pre(&migration)
		}
		err := m.rollback(&migration, bridge)
		if err != nil {
			return err
		} else {
			m.ups = m.ups[:len(m.ups)-1]
			m.downs = append(m.downs, last)
			m.migrations[migration.Version] = migration
		}
		if post != nil {
			post(&migration)
		}
	}

	return nil
}
