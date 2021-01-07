package main

import (
	"fmt"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/jozn/xox/snaker"
	"github.com/kr/pretty"
	"log"
	"strings"
)

func cqlTypesToGoType(sqlType string) (typ, org, def string) {
	switch strings.ToLower(sqlType) {
	case "string", "uuid", "text", "varchar":
		typ = "string"
		org = "string"
		def = `""`
	case "bool":
		typ = "bool"
		org = "bool"
		def = `false`
	case "int", "serial", "tinyint", "smallint":
		typ = "int"
		org = "int"
		def = `0`
	case "bigint":
		typ = "int"
		org = "int64"
		def = `0`
	case "json":
		typ = "string"
		org = "string"
		def = `""`
	case "bytes", "blob":
		typ = "[]byte"
		org = "[]byte"
		def = `[]byte{}`
	case "date", "time", "timestamp":
		typ = "time.Time"
		org = "time.Time"
		def = `time.Time.Now()`
	case "decimal":
		typ = "float64"
		org = "float64"
		def = `0`
	case "float":
		typ = "float32"
		org = "float32"
		def = `0`

	default:
		typ = "UNKNOWN_sqlToGo__" + typ
		def = `""`
	}
	// todo add this:
	//asci, counter, double, duration, inet, timeuuid, uuid, varint,
	return
}

// https://docs.datastax.com/en/cql-oss/3.x/cql/cql_reference/cql_data_types_c.html
func cqlTypesToRustType(sqlType string) (typ, org, def string) {
	switch strings.ToLower(sqlType) {
	case "string", "text", "varchar", "asci", "inet":
		typ = "String"
		org = "&str"
		def = `"".to_string()`
	case "bool":
		typ = "bool"
		org = "bool"
		def = `false`
	case "int", "serial", "tinyint", "smallint", "varint":
		typ = "i32"
		org = "i32"
		def = `0i32`
	case "bigint", "counter":
		typ = "i64"
		org = "i64"
		def = `0i64`
	case "json":
		typ = "String"
		org = "string"
		def = `""`
	case "bytes", "blob":
		typ = "Vec<u8>"
		org = "&Vec<u8>"
		def = `vec![]`
	case "date", "time":
		typ = "String"
		org = "&str"
		def = `"".to_string()`
	case "timestamp":
		typ = "String"
		org = "&str"
		def = `"".to_string()`
	case "double", "decimal":
		typ = "f64"
		org = "f64"
		def = `0f64`
	case "float":
		typ = "f32"
		org = "f32"
		def = `0f32`
	case "uuid", "timeuuid":
		typ = "String"
		org = "&str"
		def = `"".to_string()`

	default:
		typ = "UNKNOWN_sqlToRust__" + typ
		org = "UNKNOWN_sqlToRust__" + typ
		def = `""`
	}
	//duration,timeuuid, uuid, map, tuple, set, list
	return
}

func CamelCase(s string) string {
	return generator.CamelCase(s)
}

func PertyPrint(a interface{}) {
	//spew.Dump(a)
	fmt.Printf("%# v \n", pretty.Formatter(a))
}

func NoErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func errLog(typ string, err error) {
	if err != nil {
		log.Printf("ERROR - %s : %s ", typ, err)
	}
}

//////////////// shortname func ////////////
func shortname(typ string, scopeConflicts ...interface{}) string {
	var v string
	var ok bool

	// check short name map
	if v, ok = _shortNameTypeMap[typ]; !ok {
		// calc the short name
		u := []string{}
		for _, s := range strings.Split(strings.ToLower(snaker.CamelToSnake(typ)), "_") {
			if len(s) > 0 && s != "id" {
				u = append(u, s[:1])
			}
		}
		v = strings.Join(u, "")

		// check go reserved names
		if n, ok := _goReservedNames[v]; ok {
			v = n
		}

		// store back to short name map
		_shortNameTypeMap[typ] = v
	}

	// initial conflicts are the default imported packages from
	// xo_package.go.tpl
	conflicts := map[string]bool{
		"sql":     true,
		"driver":  true,
		"csv":     true,
		"errors":  true,
		"fmt":     true,
		"regexp":  true,
		"strings": true,
		"time":    true,
	}

	// add scopeConflicts to conflicts
	for _, c := range scopeConflicts {
		switch k := c.(type) {
		case string:
			conflicts[k] = true

		case []*CColumn:
			for _, f := range k {
				conflicts[f.ColumnName] = true
			}
			/*case []*QueryParam:
			  for _, f := range k {
			      conflicts[f.Name] = true
			  }*/

		default:
			panic("not implemented")
		}
	}

	// append suffix if conflict exists
	if _, ok := conflicts[v]; ok {
		v = v + "_sufix" //NameConflictSuffix
	}

	return v
}

// _shortNameTypeMap is the collection of Go style short names for types, mainly
// used for use with declaring a func receiver on a type.
var _shortNameTypeMap = map[string]string{
	"bool":        "b",
	"string":      "s",
	"byte":        "b",
	"rune":        "r",
	"int":         "i",
	"int16":       "i",
	"int32":       "i",
	"int64":       "i",
	"uint":        "u",
	"uint8":       "u",
	"uint16":      "u",
	"uint32":      "u",
	"uint64":      "u",
	"float32":     "f",
	"float64":     "f",
	"Slice":       "s",
	"StringSlice": "ss",
}

var _goReservedNames = map[string]string{
	"break":       "brk",
	"case":        "cs",
	"chan":        "chn",
	"const":       "cnst",
	"continue":    "cnt",
	"default":     "def",
	"defer":       "dfr",
	"else":        "els",
	"fallthrough": "flthrough",
	"for":         "fr",
	"func":        "fn",
	"go":          "goVal",
	"goto":        "gt",
	"if":          "ifVal",
	"import":      "imp",
	"interface":   "iface",
	"map":         "mp",
	"package":     "pkg",
	"range":       "rnge",
	"return":      "ret",
	"select":      "slct",
	"struct":      "strct",
	"switch":      "swtch",
	"type":        "typ",
	"var":         "vr",

	// go types
	"error":      "e",
	"bool":       "b",
	"string":     "str",
	"byte":       "byt",
	"rune":       "r",
	"uintptr":    "uptr",
	"int":        "i",
	"int8":       "i8",
	"int16":      "i16",
	"int32":      "i32",
	"int64":      "i64",
	"uint":       "u",
	"uint8":      "u8",
	"uint16":     "u16",
	"uint32":     "u32",
	"uint64":     "u64",
	"float32":    "z",
	"float64":    "f",
	"complex64":  "c",
	"complex128": "c128",
}

