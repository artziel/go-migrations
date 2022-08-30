package migrations

import (
	"bytes"
	"database/sql"
	"html/template"
	"os"
	"regexp"
	"strings"
)

const (
	StatusUp   string = "UP"
	StatusDown string = "DOWN"
)

func parseTplSQL(sql string, bridge interface{}) (string, error) {
	buf := &bytes.Buffer{}

	tmpl, err := template.New("").Parse(sql)

	if err != nil {
		return "", err
	}

	err = tmpl.Execute(buf, &bridge)

	return buf.String(), err
}

func FromString(body string) Migration {

	m := Migration{}

	version := regexp.MustCompile(`(?i)\-- Version: (\d.*)\n`).FindStringSubmatch(body)
	if len(version) > 1 {
		m.Version = strings.TrimSpace(version[1])
	}

	name := regexp.MustCompile(`(?i)\-- Name: (.*)\n`).FindStringSubmatch(body)

	if len(name) > 1 {
		m.Name = strings.TrimSpace(name[1])
	}

	up := regexp.MustCompile(`(?i)\-- up start(?s)(.*)\-- up end`).FindStringSubmatch(body)
	if len(up) > 1 {
		m.Up = strings.TrimSpace(up[1])
	}

	down := regexp.MustCompile(`(?i)\-- down start(?s)(.*)\-- down end`).FindStringSubmatch(body)

	if len(down) > 1 {
		m.Down = strings.TrimSpace(down[1])
	}

	return m
}

func FromFile(fileName string) (Migration, error) {
	var m Migration
	content, err := os.ReadFile(fileName)
	if err != nil {
		return m, err
	}

	m = FromString(string(content))

	return m, nil
}

type Migration struct {
	Version   string
	Name      string
	Up        string
	Down      string
	Status    string
	StartTime *sql.NullTime
	EndTime   *sql.NullTime
}
