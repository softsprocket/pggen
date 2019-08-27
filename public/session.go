package public

import (
	"fmt"
	"pggen/pgsql"
	"reflect"
	"time"
)

// Session models the table public.session
type Session struct {
	pgSQL   *pgsql.PgSQL
	Id      string
	Created time.Time
	Updated time.Time
	Store   pgsql.JSONStr
}

// SessionPrimaryKey models the primary key for the table public.session
type SessionPrimaryKey struct {
	Id string
}

// NewSession instantiates and returns a Session struct
func NewSession(pgSQL *pgsql.PgSQL) *Session {
	s := new(Session)
	s.pgSQL = pgSQL

	return s
}

// Create inserts a Session record into the public.session table
// using the values of the interface argument as an initializer
func (session *Session) Create(s interface{}) (*SessionPrimaryKey, error) {
	if err := session.pgSQL.Db.Ping(); err != nil {
		return nil, err
	}

	insertStmt := fmt.Sprintf("%s %s", pgsql.InsertClause(s, "session"), "returning id")

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

	row := session.pgSQL.Db.QueryRow(insertStmt, ifs...)
	pk := new(SessionPrimaryKey)
	err := row.Scan(&pk.Id)

	return pk, err
}

// Read selects the  public.session row keyed by  SessionPrimaryKey and returns a *Session, error tuple
func (session *Session) Read(pk *SessionPrimaryKey) (*Session, error) {
	if err := session.pgSQL.Db.Ping(); err != nil {
		return nil, err
	}

	selectStmt := "select id, created, updated, store from session where id = $1"

	row := session.pgSQL.Db.QueryRow(selectStmt, pk.Id)

	err := row.Scan(&session.Id, &session.Created, &session.Updated, &session.Store)

	return session, err
}

// Update upates the row of the public.session table represented by the Session argument
func (session *Session) Update(s *Session) error {
	if err := session.pgSQL.Db.Ping(); err != nil {
		return err
	}

	updateStmt := "update session set created = $1, updated = $2, store = $3 where id = $1"
	_, err := session.pgSQL.Db.Exec(updateStmt, s.Created, s.Updated, s.Store, s.Id)

	return err
}

// Delete removes the Session row from the database
func (session *Session) Delete(pk *SessionPrimaryKey) error {
	if err := session.pgSQL.Db.Ping(); err != nil {
		return err
	}

	deleteStmt := "delete from session  where id = $1"
	_, err := session.pgSQL.Db.Exec(deleteStmt, pk.Id)

	return err
}
