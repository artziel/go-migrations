package GoMigrations

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
