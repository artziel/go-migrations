package migrations

import "errors"

var (
	ErrNoOpenConnection           = errors.New("no open connection found")
	ErrDuplicatedMigrationVersion = errors.New("duplicated migration version")
	ErrMigrationAlreadyUp         = errors.New("migration already up")
	ErrMigrationAlreadyDown       = errors.New("migration already down")
)
