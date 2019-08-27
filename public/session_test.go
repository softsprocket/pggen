package public_test

import (
	"fmt"
	"pggen/pgsql"
	. "pggen/public"
	"reflect"
	"testing"
	"time"
)

type sessionDbConnection struct {
	PgSQL *pgsql.PgSQL
}

var sessionconn sessionDbConnection

func sessionSetup(t *testing.T) {
	fmt.Println("Running setup")
	if sessionconn.PgSQL == nil {
		connectionStr := "database=touch user=touch_dbuser password=verified_touch_user_001 host=192.168.0.107"
		pg, err := pgsql.NewPgSQL(connectionStr)
		if err != nil {
			t.Fatalf("\nPgSQL error during setup: %s\n", err)
		}

		sessionconn.PgSQL = pg
	}
}

func TestPublicSession(t *testing.T) {
	sessionSetup(t)

	session := NewSession(sessionconn.PgSQL)

	s := struct {
		Id      string
		Created time.Time
		Updated time.Time
		Store   pgsql.JSONStr
	}{
		Id:      "urn:uuid:e86a949b-020e-4170-a705-aeddb0b242fc",
		Created: pgsql.TimeOnly(time.Parse(time.RFC3339, "2019-08-27T10:21:19-07:00")),
		Updated: pgsql.TimeOnly(time.Parse(time.RFC3339, "2019-08-27T10:21:19-07:00")),
		Store:   `{"ID":123,"Name":"Hello, World"}`,
	}

	pk, err := session.Create(s)

	if err != nil {
		t.Fatalf("\nError from Create row for %s\n%s\n", "session", err)
	}

	returnedVal, err := session.Read(pk)
	if err != nil {
		t.Fatalf("\nError from Read row for %s\n%s\n", "session", err)
	}

	if !reflect.DeepEqual(returnedVal, session) {
		t.Errorf("Failed equivalency for returnedVal and %s", "session")
	}

	err = session.Delete(pk)
	if err != nil {
		t.Fatalf("\nError from Delete row for %s\n%s\n", "session", err)
	}

}
