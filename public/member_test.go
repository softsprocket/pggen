package public_test

import (
	"fmt"
	"pggen/pgsql"
	. "pggen/public"
	"reflect"
	"testing"
)

type memberDbConnection struct {
	PgSQL *pgsql.PgSQL
}

var memberconn memberDbConnection

func memberSetup(t *testing.T) {
	fmt.Println("Running setup")
	if memberconn.PgSQL == nil {
		connectionStr := "database=touch user=touch_dbuser password=verified_touch_user_001 host=192.168.0.107"
		pg, err := pgsql.NewPgSQL(connectionStr)
		if err != nil {
			t.Fatalf("\nPgSQL error during setup: %s\n", err)
		}

		memberconn.PgSQL = pg
	}
}

func TestPublicMember(t *testing.T) {
	memberSetup(t)

	member := NewMember(memberconn.PgSQL)

	s := struct {
		Firstname string
		Lastname  string
		Email     string
		Password  string
	}{
		Firstname: "test 1",
		Lastname:  "test 2",
		Email:     "test 3",
		Password:  "test 4",
	}

	pk, err := member.Create(s)

	if err != nil {
		t.Fatalf("\nError from Create row for %s\n%s\n", "member", err)
	}

	returnedVal, err := member.Read(pk)
	if err != nil {
		t.Fatalf("\nError from Read row for %s\n%s\n", "member", err)
	}

	if !reflect.DeepEqual(returnedVal, member) {
		t.Errorf("Failed equivalency for returnedVal and %s", "member")
	}

	err = member.Delete(pk)
	if err != nil {
		t.Fatalf("\nError from Delete row for %s\n%s\n", "member", err)
	}

}
