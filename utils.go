package main

import (
	"fmt"
	"github.com/golang/protobuf/protoc-gen-go/generator"
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
		typ = "string"
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
