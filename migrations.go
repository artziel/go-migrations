package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

func MigrationFromString(body string) Migration {

	migration := Migration{}

	migration.Up = migration.getUpQuery(body)
	migration.Down = migration.getDownQuery(body)
	migration.Version = migration.getVersion(body)
	migration.Name = migration.getName(body)

	return migration
}

func MigrationFromFile(fileName string) (Migration, error) {
	migration := Migration{}
	content, err := os.ReadFile(fileName)
	if err != nil {
		return migration, err
	}

	migration.Up = migration.getUpQuery(string(content))
	migration.Down = migration.getDownQuery(string(content))
	migration.Version = migration.getVersion(string(content))
	migration.Name = migration.getName(string(content))

	return migration, nil
}

/**
 * Query for migration execution log
 */
const migrationsQuery string = `
CREATE TABLE IF NOT EXISTS {{TABLE_PREFIX}}migrations (
	version VARCHAR(14) NOT NULL,
	migration_name VARCHAR(128) NULL,
	start_time TIMESTAMP NULL DEFAULT NOW(),
	end_time TIMESTAMP NULL,
	PRIMARY KEY (version),
	UNIQUE INDEX version_UNIQUE (version ASC))
ENGINE = InnoDB;
`

/**
 * Migrations structure
 */
type Migrations struct {
	db          *sql.DB
	tablePrefix string
	migrations  map[string]Migration
	ups         []string
	downs       []string
}

func (m *Migrations) Transaction(fnc func(tx *sql.Tx) error) error {

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
func (m *Migrations) Reset() {
	for k := range m.migrations {
		delete(m.migrations, k)
	}
	m.ups = []string{}
	m.downs = []string{}
}

/**
 * Add a migration to Migrations structure
 */
func (m *Migrations) addMigration(migration Migration) {

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
}

/**
 * Transverse a Migrations slice and add no duplicated migrations. If duplicated
 * the function Migrations.Reset() is executed and return an error
 */
func (m *Migrations) AddMigrations(migrations []Migration) error {

	versions := map[string]bool{}

	for _, migration := range migrations {
		if _, found := versions[migration.Version]; found {
			m.Reset()
			return errors.New("found duplicated migration version \"" + migration.Version + "\"")
		} else {
			m.addMigration(migration)
			versions[migration.Version] = true
		}
	}
	return nil
}

/**
 * Check if a migration version is already up
 */
func (m *Migrations) IsVersionUp(version string) bool {
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
func (m *Migrations) Initialize(db *sql.DB, tblprefix string) (*Migrations, error) {
	m.migrations = map[string]Migration{}
	m.ups = []string{}
	m.downs = []string{}

	query := ""
	if tblprefix != "" {
		query = strings.Replace(migrationsQuery, "{{TABLE_PREFIX}}", tblprefix+"_", 1)
		m.tablePrefix = tblprefix
	} else {
		query = strings.Replace(migrationsQuery, "{{TABLE_PREFIX}}", "", 1)
	}

	if _, err := db.Exec(query); err != nil {
		panic(err)
	}
	m.db = db

	queryUps := "SELECT version, migration_name, start_time, end_time FROM " + m.tablePrefix + "_" + "migrations ORDER BY version ASC"
	rows, err := m.db.Query(queryUps)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		migration := Migration{}
		err := rows.Scan(&migration.Version, &migration.Name, &migration.StartTime, &migration.EndTime)
		if err != nil {
			log.Fatal(err)
		} else {
			migration.Status = StatusUp
			m.migrations[migration.Version] = migration
			m.ups = append(m.ups, migration.Version)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return m, err
}

func (m *Migrations) GetMigrationsByStatus(status string) []Migration {

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

func (m *Migrations) GetMigrationStatus(migration *Migration) (string, error) {

	query := `SELECT version, start_time, end_time FROM migrations WHERE version = ? LIMIT 1`

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

func (m *Migrations) GetMigrations() []Migration {

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

func (m *Migrations) LoadFolder(path string) error {

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		migration, err := MigrationFromFile(path + "/" + file.Name())
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

func (m *Migrations) HasDownMigrations() bool {
	return len(m.downs) > 0
}

func (m *Migrations) HasUpMigrations() bool {
	return len(m.ups) > 0
}

func (m *Migrations) migrate(migration *Migration, bridge interface{}) error {
	if migration.Status != StatusDown {
		return errors.New("migration " + migration.Version + " >> Migration already Up")
	}

	if err := m.Transaction(func(tx *sql.Tx) error {
		start := time.Now()

		inQuery := "INSERT INTO " + m.tablePrefix + "_" + "migrations(version, migration_name, start_time)VALUES(?,?,?);"
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

		updateQuery := "UPDATE " + m.tablePrefix + "_" + "migrations SET end_time = ? WHERE version = ?;"
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

func (m *Migrations) Migrate(bridge interface{}, pre func(*Migration), post func(*Migration)) error {

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

func (m *Migrations) rollback(migration *Migration, bridge interface{}) error {

	if migration.Status != StatusUp {
		return errors.New("migration " + migration.Version + " >> Migration already Down")
	}
	downQuery, err := parseTplSQL(migration.Down, bridge)
	if err != nil {
		return err
	}

	if err := m.Transaction(func(tx *sql.Tx) error {
		if downQuery != "" {
			fmt.Printf("Down Query executed: %v\n", downQuery)
			if _, err := m.db.Exec(downQuery); err != nil {
				return errors.New("migration " + migration.Version + " >> " + err.Error())
			}
		}

		delQuery := "DELETE FROM " + m.tablePrefix + "_" + "migrations WHERE version = ?"
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

func (m *Migrations) Rollback(bridge interface{}, pre func(*Migration), post func(*Migration)) error {

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
