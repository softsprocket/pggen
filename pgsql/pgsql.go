package pgsql

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	// github.com/lib/pq is imported to initalize the postgres driver
	_ "github.com/lib/pq"
)

// PgSQL is a wrapper around a postgres sql.DB
type PgSQL struct {
	Db *sql.DB
}

// NewPgSQL opens a connection to a postgres database using the connection string
func NewPgSQL(connectionString string) (*PgSQL, error) {
	Db, err := sql.Open("postgres", connectionString)

	pg := new(PgSQL)
	pg.Db = Db

	return pg, err
}

// Count returns row count
func (pg *PgSQL) Count(tableName string) (int64, error) {

	query := "select count(*) from " + tableName

	row := pg.Db.QueryRow(query)
	var n int64
	err := row.Scan(&n)

	return n, err
}

// WhereClause accepts a struct in the form of name = value
// where name is a table column name and value is the value
// desired in the where clause
// the return value is in the form " where col1 = val1 and col2 = val2"
func WhereClause(s interface{}) string {
	e := reflect.TypeOf(s)
	var b bytes.Buffer
	sep := " where "
	for i := 0; i < e.NumField(); i++ {
		nam := e.Field(i).Name
		typ := e.Field(i).Type
		val := reflect.ValueOf(s).Field(i)

		if typ.Kind() == reflect.Int32 || typ.Kind() == reflect.Int64 || typ.Kind() == reflect.Float32 || typ.Kind() == reflect.Float64 {
			b.WriteString(fmt.Sprintf("%v %v=%v", sep, nam, val))
		} else if typ.Kind() == reflect.Struct {
			b.WriteString(fmt.Sprintf("%v %v='%s'", sep, nam, val.MethodByName("Format").Call([]reflect.Value{reflect.ValueOf(time.RFC3339)})))
		} else {
			b.WriteString(fmt.Sprintf("%v %v='%v'", sep, nam, val))
		}
		sep = " and "
	}

	return string(b.Bytes())
}

// SelectClause accepts a struct representing a table and returns
// a select clause using its fields in the form of
// "select field1, field2 ..."
func SelectClause(s interface{}) string {
	e := reflect.TypeOf(s)
	var b bytes.Buffer
	sep := "select"
	for i := 0; i < e.NumField(); i++ {
		nam := e.Field(i).Name
		b.WriteString(fmt.Sprintf("%v %v", sep, nam))
		sep = ","
	}

	return string(b.Bytes())

}

func getType(s interface{}) string {
	t := reflect.TypeOf(s)

	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}

	return t.Name()
}

// InsertClause returns a insert clause based on the fields in s
// in the form of "insert into name (field1, field2) values ({0}, {1})"
func InsertClause(s interface{}, name string) string {
	e := reflect.TypeOf(s)
	var b bytes.Buffer
	sep := "insert into"
	b.WriteString(fmt.Sprintf("%v %v ", sep, name))
	sep = "("
	for i := 0; i < e.NumField(); i++ {
		name = e.Field(i).Name
		b.WriteString(fmt.Sprintf("%v %v", sep, name))
		sep = ","
	}

	sep = ") values ("
	for i := 0; i < e.NumField(); i++ {
		b.WriteString(fmt.Sprintf("%v $%d", sep, i+1))
		sep = ","
	}

	b.WriteString(")")

	return string(b.Bytes())

}

// FieldArguments returns a string in the form "variable.Field1, variable.Field2 or "&variable.Field1, &variable.Field2" if pointers is true
func FieldArguments(variable string, s interface{}, pointers bool) string {
	var b bytes.Buffer
	e := reflect.TypeOf(s)
	sep := ""
	amp := ""
	if pointers {
		amp = "&"
	}
	for i := 0; i < e.NumField(); i++ {
		name := e.Field(i).Name
		b.WriteString(fmt.Sprintf("%v %s%v", sep, amp, name))
		sep = ","
	}

	return string(b.Bytes())
}

//IsPrimaryKey searches through table constraints to see if a column exists and if it is part of a primary key.
func IsPrimaryKey(c *Column, tableConstraints []*TableConstraints) bool {

	for _, ea := range tableConstraints {
		if ea.ColumnName == c.Name && ea.ConstraintType == "PRIMARY KEY" {
			return true
		}
	}

	return false
}

//ReturnKeyClause accepts a column and the constraints from its table and
// returns a string in the form "returning key1, key2"
func ReturnKeyClause(columns []*Column, tableConstraints []*TableConstraints) string {
	var b bytes.Buffer
	sep := "returning"
	for _, column := range columns {
		if IsPrimaryKey(column, tableConstraints) {
			b.WriteString(fmt.Sprintf("%v %v", sep, column.Name))
			sep = ","
		}
	}
	return string(b.Bytes())
}

//PrimaryKeyWhereClause accepts a column and the constraints from its table and
// returns a string in the form "where key1 = $1 and key2 = $2"
func PrimaryKeyWhereClause(columns []*Column, tableConstraints []*TableConstraints) string {
	var b bytes.Buffer
	sep := "where"

	for i, column := range columns {
		if IsPrimaryKey(column, tableConstraints) {
			b.WriteString(fmt.Sprintf("%v %v = $%d", sep, column.Name, i+1))
			sep = " and"
		}
	}

	return string(b.Bytes())
}

// PrimaryKeyNames returns a list of the names of the primary key components
// drawn from the arguments
func PrimaryKeyNames(columns []*Column, tableConstraints []*TableConstraints) []string {
	primaryKeyNames := make([]string, 0)
	for _, column := range columns {
		if IsPrimaryKey(column, tableConstraints) {
			primaryKeyNames = append(primaryKeyNames, column.Name)
		}
	}

	return primaryKeyNames
}

// NonPrimaryKeyNames returns a list of the names of the non-primary key components
// drawn from the arguments
func NonPrimaryKeyNames(columns []*Column, tableConstraints []*TableConstraints) []string {
	nonPrimaryKeyNames := make([]string, 0)
	for _, column := range columns {
		if !IsPrimaryKey(column, tableConstraints) {
			nonPrimaryKeyNames = append(nonPrimaryKeyNames, column.Name)
		}
	}

	return nonPrimaryKeyNames
}

//PrimaryKeyFunctionArgs accepts a column, the constraints from its table, the PrimaryKey variable name
//and a bool indication if a pointer is required and returns a string in the form "varname.key1, varname.key2" or "&varname.key1, &varname.key2"
func PrimaryKeyFunctionArgs(columns []*Column, tableConstraints []*TableConstraints, varname string, isPointer bool) string {
	var b bytes.Buffer
	sep := ""

	pointer := ""
	if isPointer {
		pointer = "&"
	}

	for _, column := range columns {
		if IsPrimaryKey(column, tableConstraints) {
			b.WriteString(fmt.Sprintf("%v%s%s.%v", sep, pointer, varname, strings.Title(column.Name)))
			sep = ", "
		}
	}
	return string(b.Bytes())
}

// Column models a postgres table's columns
type Column struct {
	Name     string
	Default  string
	Nullable bool
	Type     string
}

// Table models a postgres table
type Table struct {
	Schema string
	Name   string
}

// TableConstraints models a postgres tables constraints
type TableConstraints struct {
	ColumnName          string
	ConstraintType      string
	IsDeferrable        bool
	isInitiallyDeferred bool
}

// GetTables returns an array of Table structs
func (pg *PgSQL) GetTables() ([]*Table, error) {
	if err := pg.Db.Ping(); err != nil {
		return nil, err
	}

	query := "select table_schema, table_name " +
		"from information_schema.tables " +
		"where table_schema not in ('pg_catalog', 'information_schema')"

	rows, err := pg.Db.Query(query)

	if err != nil {
		return nil, err
	}

	tables := []*Table{}

	for rows.Next() {
		t := new(Table)

		err := rows.Scan(&t.Schema, &t.Name)

		if err != nil {
			return nil, err
		}

		tables = append(tables, t)
	}

	return tables, nil
}

func nullableToString(n sql.NullString) string {
	if n.Valid {
		return n.String
	}

	return ""
}

func nullableToBool(n sql.NullString) bool {
	if n.Valid && n.String == "YES" {
		return true
	}

	return false
}

// GetColumns returns a array of Columns as defined in information_schema columns
func (pg *PgSQL) GetColumns(table *Table) ([]*Column, error) {
	if err := pg.Db.Ping(); err != nil {
		return nil, err
	}

	query := "select column_name, column_default, is_nullable, data_type from information_schema.columns where table_schema = $1 and table_name = $2"
	rows, err := pg.Db.Query(query, table.Schema, table.Name)

	if err != nil {
		return nil, err
	}

	columns := []*Column{}

	for rows.Next() {
		c := new(Column)
		tmp := struct {
			Name     sql.NullString
			Default  sql.NullString
			Nullable sql.NullString
			Type     sql.NullString
		}{}

		err := rows.Scan(&tmp.Name, &tmp.Default, &tmp.Nullable, &tmp.Type)
		if err != nil {
			return nil, err
		}

		c.Name = nullableToString(tmp.Name)
		c.Default = nullableToString(tmp.Default)
		c.Nullable = nullableToBool(tmp.Nullable)
		c.Type = nullableToString(tmp.Type)

		columns = append(columns, c)
	}

	return columns, nil
}

// GetTableConstraints returns values from information_schema.table_constraints for the table passed ass an argument
func (pg *PgSQL) GetTableConstraints(table *Table) ([]*TableConstraints, error) {
	if err := pg.Db.Ping(); err != nil {
		return nil, err
	}

	query := "select tc.constraint_type, tc.initially_deferred, tc.is_deferrable, cu.column_name from information_schema.table_constraints tc " +
		"join information_schema.constraint_column_usage cu on cu.Table_schema = tc.table_schema and cu.table_name = tc.table_name and cu.constraint_name = tc.constraint_name " +
		"where tc.table_schema = $1 and tc.table_name = $2"

	rows, err := pg.Db.Query(query, table.Schema, table.Name)

	if err != nil {
		return nil, err
	}

	tableConstraints := []*TableConstraints{}

	for rows.Next() {
		tc := new(TableConstraints)

		tmp := struct {
			ColumnName          sql.NullString
			ConstraintType      sql.NullString
			IsDeferrable        sql.NullString
			IsInitiallyDeferred sql.NullString
		}{}

		err := rows.Scan(&tmp.ConstraintType, &tmp.IsInitiallyDeferred, &tmp.IsDeferrable, &tmp.ColumnName)
		if err != nil {
			return nil, err
		}

		tc.ColumnName = nullableToString(tmp.ColumnName)
		tc.ConstraintType = nullableToString(tmp.ConstraintType)
		tc.IsDeferrable = nullableToBool(tmp.IsDeferrable)
		tc.isInitiallyDeferred = nullableToBool(tmp.IsInitiallyDeferred)

		tableConstraints = append(tableConstraints, tc)
	}

	return tableConstraints, nil
}
