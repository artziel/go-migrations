package migrations

import (
	"testing"
)

const testMigrationSQL = `
-- ------------------------------------------------------------
-- Version: 20211223120557
-- Name: Migration Name
-- ------------------------------------------------------------
-- ------------------------------------------------------------
-- Up Start

INSERT INTO catalogs (id, tag) VALUES ( 1, 'MASTER' );

-- Up End
-- ------------------------------------------------------------
-- ------------------------------------------------------------
-- Down Start

DELETE FROM catalogs;

-- Down End
-- ------------------------------------------------------------
`

type errTests struct {
	migration       Migration
	expectedUp      string
	expectedDown    string
	expectedName    string
	expectedVersion string
}

type testBridge struct {
	String string
}

func (t testBridge) DoSomething() string {
	return "SomeString"
}

func TestCustomBridgeMessage(t *testing.T) {

	tests := []errTests{
		{migration: Migration{}, expectedUp: "", expectedDown: "", expectedName: "", expectedVersion: ""},
	}

	for i, test := range tests {
		if test.expectedUp != test.migration.Up {
			t.Errorf("Test %d: Unexpected value result of parsing Migration.Up\n", i)
		}
		if test.expectedDown != test.migration.Down {
			t.Errorf("Test %d: Unexpected value result of parsing Migration.Down\n", i)
		}
		if test.expectedName != test.migration.Name {
			t.Errorf("Test %d: Unexpected value result of parsing Migration.Name\n", i)
		}
		if test.expectedVersion != test.migration.Version {
			t.Errorf("Test %d: Unexpected value result of parsing Migration.Version\n", i)
		}
	}
}

func TestParseTplSQL(t *testing.T) {

	sql, err := parseTplSQL("INSERT INTO table (id, username, password)VALUES(1, 'sample@sample.com', '{{ .DoSomething }}')", testBridge{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else {
		expected := "INSERT INTO table (id, username, password)VALUES(1, 'sample@sample.com', 'SomeString')"
		if sql != expected {
			t.Errorf("Returned SQL dont match:\ngot  %s\nwant %s", sql, expected)
		}
	}

	sql, err = parseTplSQL("INSERT INTO table (id, username, password)VALUES(1, 'sample@sample.com', '{{ .GenerateError }}')", testBridge{})
	if err == nil && sql == "" {
		t.Errorf("Expected error result got No Error")
	}
}

func TestMigration(t *testing.T) {

	m := FromString(testMigrationSQL)

	expected := "Migration Name"
	if result := m.getName(testMigrationSQL); result != expected {
		t.Errorf("getName return unexpected value:\ngot  %s\nwant %s", result, expected)
	}

	expected = ""
	if result := m.getName(""); result != expected {
		t.Errorf("getName return unexpected value:\ngot  %s\nwant %s", result, expected)
	}

	expected = "20211223120557"
	if result := m.getVersion(testMigrationSQL); result != expected {
		t.Errorf("getVersion return unexpected value:\ngot  %s\nwant %s", result, expected)
	}

	expected = ""
	if result := m.getVersion(""); result != expected {
		t.Errorf("getVersion return unexpected value:\ngot  %s\nwant %s", result, expected)
	}

	expected = "INSERT INTO catalogs (id, tag) VALUES ( 1, 'MASTER' );"
	if result := m.getUpQuery(testMigrationSQL); result != expected {
		t.Errorf("getUpQuery return unexpected value:\ngot  %s\nwant %s", result, expected)
	}

	expected = ""
	if result := m.getUpQuery(""); result != expected {
		t.Errorf("getUpQuery return unexpected value:\ngot  %s\nwant %s", result, expected)
	}

	expected = "DELETE FROM catalogs;"
	if result := m.getDownQuery(testMigrationSQL); result != expected {
		t.Errorf("getDownQuery return unexpected value:\ngot  %s\nwant %s", result, expected)
	}

	expected = ""
	if result := m.getDownQuery(""); result != expected {
		t.Errorf("getDownQuery return unexpected value:\ngot  %s\nwant %s", result, expected)
	}
}
