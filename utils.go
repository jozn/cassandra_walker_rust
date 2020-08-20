package main

import (
	"fmt"
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
	return
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
