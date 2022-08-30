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

DROP TABLE catalogs;

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

func migrate(ms *Migrations.Pool) {
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

func rollback(ms *Migrations.Pool) {
	if err := ms.Rollback(nil, nil, nil); err != nil {
		panic(err)
	}

	fmt.Println("\nAfter Rollback --------------------------------------------")
	for i, m := range ms.GetMigrations() {
		fmt.Printf("%v) %v - %v - %v\n", i, m.Version, m.Name, m.Status)
	}
}

func main() {
	db, err := Migrations.OpenMySql(Migrations.MySqlSettings{
		Username: "admin",
		Password: "admin",
		Host:     "localhost",
		Port:     "3306",
		Database: "go_catalogs",
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if pool, err := Migrations.NewPool(db); err != nil {

		m1 := Migrations.FromString(Sample1Migration)
		m2 := Migrations.FromString(Sample2Migration)

		if err := pool.AddMigrations([]Migrations.Migration{m1, m2}); err != nil {
			panic(err)
		}

		migrate(&pool)
		rollback(&pool)
	}
}
