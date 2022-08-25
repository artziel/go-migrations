package migrations

import "errors"

var ErrDuplicatedMigrationVersion = errors.New("duplicated migration version")
