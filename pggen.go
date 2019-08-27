package main

import (
	"fmt"
	"io/ioutil"
	"locker/vault"
	"log"
	"os"
	"path/filepath"
	"pggen/pgsql"
	"strings"
	"text/template"
)

type args struct {
	Vault            string
	Key              string
	Filename         string
	OutputPath       string
	ConnectionString string
	PackageRoot      string
}

func help() {
	fmt.Println("\npggen <[-v path_to_vault -k vault_key -f encrypted_filename] | [-c connection_string]> [-o outputPath] [-p packageRoot] ")
	fmt.Println("If a connection string to a pgsql db is not included (-c option) then the -v -k -f combination is expected")
	fmt.Println("where encrypted_filename is a vault file containing a encrypted connection string.")
	fmt.Println("\npggen -h")
	fmt.Println("Prints this help message and exits the program.")

}

func nextArg(args []string, i int, expectation string) string {
	if len(args) <= (i + 1) {
		help()
		panic("pggen:" + expectation)
	}

	return args[i+1]
}

func parseArgs() (args, bool) {
	if len(os.Args) == 1 {
		help()
		os.Exit(-1)
	}

	a := args{}
	oa := os.Args[1:]

	for i := 0; i < len(oa); i++ {
		switch oa[i] {
		case "-v":
			a.Vault = nextArg(oa, i, "argument -v (vault location expected)")
			i++
		case "-k":
			a.Key = nextArg(oa, i, "arguments -k (key string expected)")
			i++
		case "-f":
			a.Filename = nextArg(oa, i, "arguments -f (filename expected")
		case "-o":
			a.OutputPath = nextArg(oa, i, "arguments -o (output path expected)")
			i++
		case "-c":
			a.ConnectionString = nextArg(oa, i, "arguments -c (connection string expected)")
		case "-p":
			a.PackageRoot = nextArg(oa, i, "arguments -p (packageRoot string expected)")
		case "-h":
			help()
			os.Exit(-1)
		}
	}

	useConnectionString := false
	if len(a.ConnectionString) == 0 {
		if len(a.Vault) == 0 || len(a.Key) == 0 || len(a.Filename) == 0 {
			help()
			os.Exit(-1)
		}
	} else {
		useConnectionString = true
	}

	return a, useConnectionString
}

func main() {
	args, useConnectionString := parseArgs()

	var connectionStr string

	if useConnectionString {
		connectionStr = args.ConnectionString
	} else {
		vault := vault.Vault{
			Path: args.Vault,
		}

		connectionStr = strings.Trim(string(vault.DecryptFromFile(args.Filename, args.Key)), "\n\t ")
	}

	fmt.Printf("Loaded connection string, %s\n", connectionStr)

	pg, err := pgsql.NewPgSQL(connectionStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "PgSQL error during setup: %s\n", err)
		return
	}

	tables, err := pg.GetTables()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get tables: %s\n", err)
		return
	}

	for _, table := range tables {
		fmt.Printf("%s.%s\n", table.Schema, table.Name)

		columns, err := pg.GetColumns(table)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to get Columns: %s\n", err)
			return
		}

		tableConstraints, err := pg.GetTableConstraints(table)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to get TableConstraints: %s\n", err)
			return
		}

		for _, column := range columns {
			fmt.Printf("\t%s %s default = \"%s\"", column.Name, column.Type, column.Default)
			tc := getColumnConstraints(tableConstraints, column.Name)

			for _, ea := range tc {
				fmt.Printf("\t%s", ea.ConstraintType)

			}

			fmt.Println()
		}

	}

	tmpl, err := ioutil.ReadFile("templates/table.tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read templates/table.tmpl: %s\n", err)
		return
	}

	testsTmpl, err := ioutil.ReadFile("templates/tests.tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read templates/tests.tmpl: %s\n", err)
		return
	}

	funcs := template.FuncMap{
		"title":                  strings.Title,
		"togo":                   togo,
		"onlyOne":                onlyOne,
		"first":                  first,
		"isPrimaryKey":           pgsql.IsPrimaryKey,
		"returnKeyClause":        pgsql.ReturnKeyClause,
		"primaryKeyFunctionArgs": pgsql.PrimaryKeyFunctionArgs,
		"createTestStruct":       pgsql.CreateTestStruct,
		"inc": func(i int) int {
			return i + 1
		},
	}

	for _, table := range tables {
		dir := filepath.Join(args.OutputPath, table.Schema)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create dir %s : %s\n", dir, err)
			os.Exit(-1)
		}

		filename := filepath.Join(dir, table.Name) + ".go"
		file, err := os.Create(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s : %s\n", filename, err)
			os.Exit(-1)
		}

		defer file.Close()

		testFilename := filepath.Join(dir, table.Name) + "_test.go"
		testFile, err := os.Create(testFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s : %s\n", testFilename, err)
			os.Exit(-1)
		}

		defer testFile.Close()

		tableTmpl, err := template.New("Table").Funcs(funcs).Parse(string(tmpl))
		if err != nil {
			log.Fatal(err)
		}

		testTmpl, err := template.New("Test").Funcs(funcs).Parse(string(testsTmpl))
		if err != nil {
			log.Fatal(err)
		}

		imp := []string{
			"pggen/pgsql",
			"fmt",
			"reflect",
			"time",
		}

		dat := struct {
			Schema             string
			Name               string
			Columns            []*pgsql.Column
			Imports            []string
			TestImports        []string
			Constraints        []*pgsql.TableConstraints
			ConnectionString   string
			PackageRoot        string
			PrimaryKeyNames    []string
			NonPrimaryKeyNames []string
		}{
			Schema:           table.Schema,
			Name:             table.Name,
			Imports:          imp,
			PackageRoot:      args.PackageRoot,
			ConnectionString: connectionStr,
		}

		columns, err := pg.GetColumns(table)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to get Columns: %s\n", err)
			return
		}

		for _, column := range columns {
			gotype, err := pgsql.ToGo(column.Type)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to get type for %s: %s\n", column.Type, err)
				return
			}
			dat.Columns = append(dat.Columns, column)
			if gotype == "time.Time" {
				contains := false

				for _, val := range dat.Imports {
					if val == "time" {
						contains = true
						break
					}
				}

				if !contains {
					dat.Imports = append(dat.Imports, "time")
				}

				contains = false
				for _, val := range dat.TestImports {
					if val == "time" {
						contains = true
						break
					}
				}

				if !contains {
					dat.TestImports = append(dat.TestImports, "time")
				}
			}

		}

		tableConstraints, err := pg.GetTableConstraints(table)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to get TableConstraints: %s\n", err)
			return
		}

		dat.Constraints = tableConstraints

		dat.PrimaryKeyNames = pgsql.PrimaryKeyNames(columns, tableConstraints)
		dat.NonPrimaryKeyNames = pgsql.NonPrimaryKeyNames(columns, tableConstraints)

		if err := tableTmpl.Execute(file, dat); err != nil {
			log.Fatal(err)
		}

		if err := testTmpl.Execute(testFile, dat); err != nil {
			log.Fatal(err)
		}

	}

	fmt.Println("\nDone.")
}

func getColumnConstraints(tableConstraints []*pgsql.TableConstraints, columnName string) []*pgsql.TableConstraints {
	tc := []*pgsql.TableConstraints{}

	for _, ea := range tableConstraints {
		if ea.ColumnName == columnName {
			tc = append(tc, ea)
		}
	}

	return tc
}

func togo(t string) string {
	tp, _ := pgsql.ToGo(t)

	return tp
}

func onlyOne(a []string) bool {
	return len(a) == 1
}

func first(a []string) string {
	return a[0]
}
