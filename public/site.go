package public

import (
	"fmt"
	"pggen/pgsql"
	"reflect"
	"time"
)

// Site models the table public.site
type Site struct {
	pgSQL    *pgsql.PgSQL
	Domain   string
	Memberid int
	Role     string
}

// SitePrimaryKey models the primary key for the table public.site
type SitePrimaryKey struct {
	Domain   string
	Memberid int
}

// NewSite instantiates and returns a Site struct
func NewSite(pgSQL *pgsql.PgSQL) *Site {
	s := new(Site)
	s.pgSQL = pgSQL

	return s
}

// Create inserts a Site record into the public.site table
// using the values of the interface argument as an initializer
func (site *Site) Create(s interface{}) (*SitePrimaryKey, error) {
	if err := site.pgSQL.Db.Ping(); err != nil {
		return nil, err
	}

	insertStmt := fmt.Sprintf("%s %s", pgsql.InsertClause(s, "site"), "returning domain, memberid")

	v := reflect.ValueOf(s)
	n := v.NumField()
	ifs := make([]interface{}, n)

	for i := 0; i < n; i++ {
		switch v.Field(i).Kind() {
		case reflect.Int8, reflect.Int, reflect.Int64:
			ifs[i] = v.Field(i).Int()
		case reflect.Uint8, reflect.Uint, reflect.Uint64:
			ifs[i] = v.Field(i).Uint()
		case reflect.Float64, reflect.Float32:
			ifs[i] = v.Field(i).Float()
		case reflect.Bool:
			ifs[i] = v.Field(i).Bool()
		case reflect.String:
			ifs[i] = v.Field(i).String()
		case reflect.Struct:
			ifs[i] = v.Field(i).MethodByName("Format").Call([]reflect.Value{reflect.ValueOf(time.RFC3339)})[0].String()
		}
	}

	row := site.pgSQL.Db.QueryRow(insertStmt, ifs...)
	pk := new(SitePrimaryKey)
	err := row.Scan(&pk.Domain, &pk.Memberid)

	return pk, err
}

// Read selects the  public.site row keyed by  SitePrimaryKey and returns a *Site, error tuple
func (site *Site) Read(pk *SitePrimaryKey) (*Site, error) {
	if err := site.pgSQL.Db.Ping(); err != nil {
		return nil, err
	}

	selectStmt := "select domain, memberid, role from site where domain = $1 and memberid = $2"

	row := site.pgSQL.Db.QueryRow(selectStmt, pk.Domain, pk.Memberid)

	err := row.Scan(&site.Domain, &site.Memberid, &site.Role)

	return site, err
}

// Update upates the row of the public.site table represented by the Site argument
func (site *Site) Update(s *Site) error {
	if err := site.pgSQL.Db.Ping(); err != nil {
		return err
	}

	updateStmt := "update site set role = $1 where domain = $1 and memberid = $2"
	_, err := site.pgSQL.Db.Exec(updateStmt, s.Role, s.Domain, s.Memberid)

	return err
}

// Delete removes the Site row from the database
func (site *Site) Delete(pk *SitePrimaryKey) error {
	if err := site.pgSQL.Db.Ping(); err != nil {
		return err
	}

	deleteStmt := "delete from site  where domain = $1 and memberid = $2"
	_, err := site.pgSQL.Db.Exec(deleteStmt, pk.Domain, pk.Memberid)

	return err
}
