package migrations

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var conn *sql.DB
var lock = &sync.Mutex{}

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

func Connection() (*sql.DB, error) {
	if conn == nil {
		return nil, ErrNoOpenConnection
	}

	return conn, nil
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
