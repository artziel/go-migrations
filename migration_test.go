package migrations

import (
	"testing"
)

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
