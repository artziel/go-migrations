package GoMigrations

type TestMigration struct {
	Migration
}

type TestBridge struct {
	MigrationBridge
}

/**
 * Compare expected data with parameter
 */
func (pt *SQLParserTest) excuteTest() (bool, []string) {

	errors := []string{}

	return len(errors) == 0, errors
}

/**
 * Row Parser Test struct
 */
type SQLParserTest struct {
	input       string
	expected    string
	errExpected []error
}

/**
 * Check if an error is in expected errors list
 */
func (rt *SQLParserTest) IsErrorExpected(e error) bool {
	if rt.errExpected == nil {
		return false
	}

	for _, err := range rt.errExpected {
		if err == e {
			return true
		}
	}

	return false
}
