package pgsql

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TimeStr for code generation purposes
//type TimeStr string

// JSONStr for code generation purposes
type JSONStr string

// ToGo takes a sql type (string) and returns a go type (string)
// for use in templating. If no conversion is available an error is returned
func ToGo(typeStr string) (string, error) {
	switch typeStr {
	case "smallint":
		return "int8", nil
	case "integer":
		return "int", nil
	case "bigint":
		return "int64", nil
	case "decimal", "numeric", "double", "double precision":
		return "float64", nil
	case "real":
		return "float32", nil
	case "smallserial":
		return "uint8", nil
	case "serial":
		return "uint32", nil
	case "bigserial":
		return "uint64", nil
	case "boolean":
		return "bool", nil
	case "jsonb":
		return "pgsql.JSONStr", nil
	}

	var strTypeRegex = regexp.MustCompile("character varying(\\d+)|varchar(\\d+)|character(\\d+)|char(\\d+)|text")
	var timeTypeRegex = regexp.MustCompile("((timestamp|time)( \\([0..6]\\))? (with|without) time zone)|date")

	switch {
	case strTypeRegex.MatchString(typeStr):
		return "string", nil
	case timeTypeRegex.MatchString(typeStr):
		return "time.Time", nil
	}

	return "", errors.New("Unknown type")
}

//TimeOnly - strip the error from a time.Time, error tuple
func TimeOnly(t time.Time, err error) time.Time {
	return t
}

//DefaultTestValue returns a string value suitable
// for use in templating test values
func DefaultTestValue(typeStr string, index int) string {
	var bval = "true"
	if index == 0 {
		bval = "false"
	}

	t := time.Now()

	switch typeStr {
	case "int8", "int", "int64", "uint8", "uint", "uint64":
		return fmt.Sprintf("%d", index)
	case "float64", "float32":
		return fmt.Sprintf("%f", float32(index))
	case "boolean":
		return bval
	case "string":
		return fmt.Sprintf("\"test %d\"", index)
	case "time.Time":
		return fmt.Sprintf("pgsql.TimeOnly(time.Parse(time.RFC3339,\"%s\"))", t.Format(time.RFC3339))
	case "pgsql.JSONStr":
		jstr, _ := json.Marshal(struct {
			ID   int
			Name string
		}{
			ID:   123,
			Name: "Hello, World",
		})
		return fmt.Sprintf("`%s`", string(jstr))
	}

	return ""
}

//NewUUID returns a uuid as urn
func NewUUID() string {
	uid := uuid.New()

	return uid.URN()
}

// CreateTestStruct produces an anonymous struct with the non-defaulted columns in Column
// assigned values
func CreateTestStruct(columns []*Column, tableConstraints []*TableConstraints) string {
	var types bytes.Buffer
	var values bytes.Buffer

	for i, column := range columns {
		if column.Default == "" {
			t, _ := ToGo(column.Type)
			types.WriteString(fmt.Sprintf("%s %s\n", strings.Title(column.Name), t))

			if t == "string" && IsPrimaryKey(column, tableConstraints) {
				values.WriteString(fmt.Sprintf("%s: \"%s\",\n", strings.Title(column.Name), NewUUID()))
			} else {
				values.WriteString(fmt.Sprintf("%s: %s,\n", strings.Title(column.Name), DefaultTestValue(t, i)))
			}
		}

	}

	return fmt.Sprintf("struct {\n%s}{\n%s}", string(types.Bytes()), string(values.Bytes()))
}
