package public_test

import (
	"fmt"
	"pggen/pgsql"
	. "pggen/public"
	"reflect"
	"testing"
)

type siteDbConnection struct {
	PgSQL *pgsql.PgSQL
}

var siteconn siteDbConnection

func siteSetup(t *testing.T) {
	fmt.Println("Running setup")
	if siteconn.PgSQL == nil {
		connectionStr := "database=touch user=touch_dbuser password=verified_touch_user_001 host=192.168.0.107"
		pg, err := pgsql.NewPgSQL(connectionStr)
		if err != nil {
			t.Fatalf("\nPgSQL error during setup: %s\n", err)
		}

		siteconn.PgSQL = pg
	}
}

func TestPublicSite(t *testing.T) {
	siteSetup(t)

	site := NewSite(siteconn.PgSQL)

	s := struct {
		Domain   string
		Memberid int
		Role     string
	}{
		Domain:   "urn:uuid:42c41dce-3318-4aa4-a125-3dd06f961bfe",
		Memberid: 1,
		Role:     "test 2",
	}

	pk, err := site.Create(s)

	if err != nil {
		t.Fatalf("\nError from Create row for %s\n%s\n", "site", err)
	}

	returnedVal, err := site.Read(pk)
	if err != nil {
		t.Fatalf("\nError from Read row for %s\n%s\n", "site", err)
	}

	if !reflect.DeepEqual(returnedVal, site) {
		t.Errorf("Failed equivalency for returnedVal and %s", "site")
	}

	err = site.Delete(pk)
	if err != nil {
		t.Fatalf("\nError from Delete row for %s\n%s\n", "site", err)
	}

}
