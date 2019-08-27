package public

import (
	"fmt"
	"pggen/pgsql"
	"reflect"
	"time"
)

// Member models the table public.member
type Member struct {
	pgSQL     *pgsql.PgSQL
	Id        int
	Firstname string
	Lastname  string
	Email     string
	Password  string
}

// MemberPrimaryKey models the primary key for the table public.member
type MemberPrimaryKey struct {
	Id int
}

// NewMember instantiates and returns a Member struct
func NewMember(pgSQL *pgsql.PgSQL) *Member {
	s := new(Member)
	s.pgSQL = pgSQL

	return s
}

// Create inserts a Member record into the public.member table
// using the values of the interface argument as an initializer
func (member *Member) Create(s interface{}) (*MemberPrimaryKey, error) {
	if err := member.pgSQL.Db.Ping(); err != nil {
		return nil, err
	}

	insertStmt := fmt.Sprintf("%s %s", pgsql.InsertClause(s, "member"), "returning id")

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

	row := member.pgSQL.Db.QueryRow(insertStmt, ifs...)
	pk := new(MemberPrimaryKey)
	err := row.Scan(&pk.Id)

	return pk, err
}

// Read selects the  public.member row keyed by  MemberPrimaryKey and returns a *Member, error tuple
func (member *Member) Read(pk *MemberPrimaryKey) (*Member, error) {
	if err := member.pgSQL.Db.Ping(); err != nil {
		return nil, err
	}

	selectStmt := "select id, firstname, lastname, email, password from member where id = $1"

	row := member.pgSQL.Db.QueryRow(selectStmt, pk.Id)

	err := row.Scan(&member.Id, &member.Firstname, &member.Lastname, &member.Email, &member.Password)

	return member, err
}

// Update upates the row of the public.member table represented by the Member argument
func (member *Member) Update(s *Member) error {
	if err := member.pgSQL.Db.Ping(); err != nil {
		return err
	}

	updateStmt := "update member set firstname = $1, lastname = $2, email = $3, password = $4 where id = $1"
	_, err := member.pgSQL.Db.Exec(updateStmt, s.Firstname, s.Lastname, s.Email, s.Password, s.Id)

	return err
}

// Delete removes the Member row from the database
func (member *Member) Delete(pk *MemberPrimaryKey) error {
	if err := member.pgSQL.Db.Ping(); err != nil {
		return err
	}

	deleteStmt := "delete from member  where id = $1"
	_, err := member.pgSQL.Db.Exec(deleteStmt, pk.Id)

	return err
}
