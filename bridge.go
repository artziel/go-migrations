package migrations

type MigrationBridge struct{}

func (mb *MigrationBridge) Encript(value string) (string, error) {
	return HashPassword(value)
}
