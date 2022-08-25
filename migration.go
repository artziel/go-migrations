package migrations

import (
	"bytes"
	"database/sql"
	"html/template"
	"regexp"
	"strings"
)

const (
	StatusUp   string = "UP"
	StatusDown string = "DOWN"
)

type Migration struct {
	Version   string
	Name      string
	Up        string
	Down      string
	Status    string
	StartTime *sql.NullTime
	EndTime   *sql.NullTime
}

func parseTplSQL(sql string, bridge interface{}) (string, error) {
	buf := &bytes.Buffer{}

	tmpl, err := template.New("").Parse(sql)

	if err != nil {
		return "", err
	}

	err = tmpl.Execute(buf, &bridge)

	return buf.String(), err
}

func (m *Migration) getVersion(body string) string {
	re := regexp.MustCompile(`(?i)\-- Version: (\d.*)\n`)
	match := re.FindStringSubmatch(body)

	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func (m *Migration) getName(body string) string {
	re := regexp.MustCompile(`(?i)\-- Name: (.*)\n`)
	match := re.FindStringSubmatch(body)

	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func (m *Migration) getUpQuery(body string) string {
	re := regexp.MustCompile(`(?i)\-- up start(?s)(.*)\-- up end`)
	match := re.FindStringSubmatch(body)

	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func (m *Migration) getDownQuery(body string) string {
	re := regexp.MustCompile(`(?i)\-- down start(?s)(.*)\-- down end`)
	match := re.FindStringSubmatch(body)

	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}
