# Golang Migrations
Artziel Narvaiza <artziel@gmail.com>

### Features
- TBD

### Dependencies
- github.com/go-sql-driver/mysql
- golang.org/x/crypto

Get the package
```bash
go get github.com/artziel/go-migrations
```

Use example:
```golang
package main

import (
	"fmt"

	Migrations "github.com/artziel/go-migrations"
)

const Sample1Migration = `
-- ------------------------------------------------------------
-- Version: 20211223120556
-- Name: Products Catalogs
-- ------------------------------------------------------------
-- ------------------------------------------------------------
-- Up Start

CREATE TABLE catalogs (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  created DATETIME NOT NULL DEFAULT NOW(),
  modified DATETIME NOT NULL DEFAULT NOW(),
  status ENUM('AVAILABLE', 'UNAVAILABLE') NOT NULL DEFAULT 'AVAILABLE',
  convertion_factor DECIMAL(10,4) NOT NULL DEFAULT 1,
  tag VARCHAR(45) NOT NULL,
  name VARCHAR(64) NOT NULL,
  description VARCHAR(256) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  UNIQUE INDEX id_UNIQUE (id ASC),
  UNIQUE INDEX name_UNIQUE (name ASC))
ENGINE = InnoDB;

-- Up End
-- ------------------------------------------------------------
-- ------------------------------------------------------------
-- Down Start

DROP TABLE catalogs;dddd

-- Down End
-- ------------------------------------------------------------
`

const Sample2Migration = `
-- ------------------------------------------------------------
-- Version: 20211223120557
-- Name: Catalog Insert
-- ------------------------------------------------------------
-- ------------------------------------------------------------
-- Up Start

INSERT INTO catalogs (id, tag, name, description)
VALUES ( 1, 'MASTER', 'Catálogo Maestro', 'Este catálogo incluye todos los productos' );

INSERT INTO catalogs (id, tag, name, description)
VALUES ( 2, 'MASTER 2', 'Catálogo Maestro 2', '{{ .Encript "123456" }}' );

-- Up End
-- ------------------------------------------------------------
-- ------------------------------------------------------------
-- Down Start

DELETE FROM catalogs;

-- Down End
-- ------------------------------------------------------------
`

func migrate(ms *Migrations.Migrations) {
	fmt.Println("Before Migration -------------------------------------------")
	for i, m := range ms.GetMigrations() {
		fmt.Printf("%v) %v - %v - %v\n", i, m.Version, m.Name, m.Status)
	}
	bridge := Migrations.MigrationBridge{}
	if err := ms.Migrate(&bridge, nil, nil); err != nil {
		panic(err)
	}

	fmt.Println("\nAfter Migration --------------------------------------------")
	for i, m := range ms.GetMigrations() {
		fmt.Printf("%v) %v - %v - %v\n", i, m.Version, m.Name, m.Status)
	}
}

func rollback(ms *Migrations.Migrations) {
	if err := ms.Rollback(nil, nil, nil); err != nil {
		panic(err)
	}

	fmt.Println("\nAfter Rollback --------------------------------------------")
	for i, m := range ms.GetMigrations() {
		fmt.Printf("%v) %v - %v - %v\n", i, m.Version, m.Name, m.Status)
	}
}

func main() {
	conn, err := Migrations.OpenMySql(Migrations.MySqlSettings{
		Username: "admin",
		Password: "admin",
		Host:     "localhost",
		Port:     "3306",
		Database: "go_catalogs",
	})
	if err != nil {
		panic(err)
	}

	ms := Migrations.Migrations{}
	ms.Initialize(conn, "alus")

	m1 := Migrations.MigrationFromString(Sample1Migration)
	m2 := Migrations.MigrationFromString(Sample2Migration)

	if err := ms.AddMigrations([]Migrations.Migration{m1, m2}); err != nil {
		panic(err)
	}

	migrate(&ms)
	rollback(&ms)

	conn.Close()
}
```
