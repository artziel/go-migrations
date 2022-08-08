package GoMigrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySqlSettings struct {
	Username         string
	Password         string
	Host             string
	Port             string
	Database         string
	MaxAllowedPacket int
	ConnMaxLifetime  time.Duration
	MaxOpenConns     int
	MaxIdleConns     int
}

func setMySqlDefaults(cnf *MySqlSettings) {
	if cnf.MaxIdleConns == 0 {
		cnf.MaxIdleConns = 10
	}
	if cnf.MaxOpenConns == 0 {
		cnf.MaxOpenConns = 10
	}
	if cnf.ConnMaxLifetime == 0 {
		cnf.ConnMaxLifetime = time.Minute * 3
	}
	if cnf.MaxAllowedPacket == 0 {
		cnf.MaxAllowedPacket = 12194304
	}
}

var ErrNoOpenConnection = errors.New("no open connection found")

var conn *sql.DB
var lock = &sync.Mutex{}

func Connection() (*sql.DB, error) {
	if conn == nil {
		return nil, ErrNoOpenConnection
	}

	return conn, nil
}

func Transaction(db *sql.DB, transaction func() error) error {

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	if err := transaction(); err != nil {
		return err
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func OpenMySql(cnf MySqlSettings) (*sql.DB, error) {

	if conn != nil {
		return conn, nil
	}

	lock.Lock()
	defer lock.Unlock()

	setMySqlDefaults(&cnf)

	dsn := fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%v?tls=skip-verify&autocommit=true&multiStatements=true&parseTime=true&maxAllowedPacket=%v",
		cnf.Username,
		cnf.Password,
		cnf.Host,
		cnf.Port,
		cnf.Database,
		cnf.MaxAllowedPacket,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(cnf.ConnMaxLifetime)
	db.SetMaxOpenConns(cnf.MaxOpenConns)
	db.SetMaxIdleConns(cnf.MaxIdleConns)

	conn = db

	return conn, nil
}

func Close() {
	conn.Close()
	conn = nil
}
