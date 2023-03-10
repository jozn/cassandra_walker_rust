package main

import (
	"fmt"
	"strings"
)

type GenOut struct {
	TablesExtracted   []*CTable
	Tables            []*TableOut
	KeyspacesDescribe []string
	Package           string
}

// A table in cassandra
type CTable struct {
	TableName string
	Keyspace  string
	Columns   []*CColumn
	//PartitionColumns []*CColumn
	//ClusterColumns   []*CColumn
}

type TableOut struct {
	CTable
	Columns          []*ColumnOut
	PartitionColumns []*ColumnOut
	ClusterColumns   []*ColumnOut
	TableShortName   string
	TableNameGo      string
	TableSchemeOut   string
	Comment          string
	OutColParams     string
	PrefixHidden     string  //hide ex: Table_Selector in docs
	GenOut           *GenOut //we need this for package refrencing, could be done better, but good

	TableNameRust string
}

// A column in cassandra
type CColumn struct {
	ColumnName   string
	Kind         string
	Position     int
	TypeCql      string
	IsPartition  bool
	IsClustering bool
	IsRegular    bool //regular column types
}

type ColumnOut struct {
	CColumn
	ColumnNameGO   string
	OutNameShorted string
	TypeGo         string
	TypeGoOriginal string
	TypeDefaultGo  string
	WhereModifiers []WhereModifier

	ColumnNameRust        string
	TypeRust              string
	TypeRustBorrow        string // remove? or something for owenership
	TypeDefaultRust       string
	WhereModifiersRust    []WhereModifier
	WhereInsModifiersRust []WhereModifierIns
}

type WhereModifier struct {
	Suffix    string
	Prefix    string
	Condition string
	AndOr     string
	FuncName  string
}

type WhereModifierIns struct {
	Suffix   string
	Prefix   string
	AndOr    string
	FuncName string
}

func setTableParams(gen *GenOut) {
	for _, table := range gen.TablesExtracted {
		t := &TableOut{
			CTable:         *table,
			Comment:        fmt.Sprintf("table: %s", table.TableName),
			TableShortName: shortname(table.TableName),
			TableSchemeOut: table.Keyspace + "." + table.TableName,
			TableNameGo:    CamelCase(table.TableName),
			PrefixHidden:   "",
			GenOut:         gen,
			// Rust
			TableNameRust: CamelCase(table.TableName),
		}
		if args.Minimize {
			t.PrefixHidden = "__"
		}
		var outColParams = ""
		for _, col := range table.Columns {
			typGo, typOrg, defGo := cqlTypesToGoType(col.TypeCql)
			typRs, typOrgRs, defRs := cqlTypesToRustType(col.TypeCql)
			c := &ColumnOut{
				CColumn:        *col,
				ColumnNameGO:   CamelCase(col.ColumnName),
				TypeGo:         typGo,
				TypeGoOriginal: typOrg,
				TypeDefaultGo:  defGo,
				// Rust
				ColumnNameRust:  col.ColumnName,
				TypeRust:        typRs,
				TypeRustBorrow:  typOrgRs,
				TypeDefaultRust: defRs,
			}
			c.OutNameShorted = fmt.Sprintf(" %s.%s", t.TableShortName, c.ColumnNameGO)
			t.Columns = append(t.Columns, c)
			if c.IsPartition {
				t.PartitionColumns = append(t.PartitionColumns, c)
			}

			if col.IsClustering {
				t.ClusterColumns = append(t.ClusterColumns, c)
			}

			outColParams += c.OutNameShorted + "," //fmt.Sprintf(" %s.%s,", t.TableShortName, c.ColumnNameGO)
			c.WhereModifiers = c.GetModifiers()
			c.WhereModifiersRust = c.GetModifiersRust()
			c.WhereInsModifiersRust = c.GetRustModifiersIns()
		}

		t.OutColParams = outColParams[:len(outColParams)-1]
		gen.Tables = append(gen.Tables, t)
	}
}

func (c *ColumnOut) GetModifiersRust() (res []WhereModifier) {
	add := func(m WhereModifier) {
		if len(m.AndOr) > 0 {
			m.FuncName = m.Prefix + "_" + c.ColumnNameRust + m.Suffix
		} else {
			m.FuncName = c.ColumnNameRust + m.Suffix
		}
		res = append(res, m)
	}
	eqAdd := func(filter, andOr string) {
		//sufix := filter + andOr
		add(WhereModifier{"_eq" + filter, strings.ToLower(andOr), "=", andOr, ""})
	}

	notEqs := func(filter, andOr string) {
		sufix := filter
		pre := strings.ToLower(andOr)
		add(WhereModifier{"_lt" + sufix, pre, "<", andOr, ""})
		add(WhereModifier{"_le" + sufix, pre, "<=", andOr, ""})
		add(WhereModifier{"_gt" + sufix, pre, ">", andOr, ""})
		add(WhereModifier{"_ge" + sufix, pre, ">=", andOr, ""})
	}
	const filter = "_filtering"
	for _, andOr := range []string{"", "AND", "OR"} {
		// todo
		if c.TypeRust == "i32" || c.TypeRust == "i64" ||
			c.TypeRust == "f32" || c.TypeRust == "f64" {

			filter := "_filtering"
			if c.IsPartition {
				eqAdd("", andOr)
				notEqs(filter, andOr)
			}
			if c.IsClustering {
				eqAdd("", andOr)
				notEqs("", andOr)
			}
			if c.IsRegular {
				eqAdd(filter, andOr)
				notEqs(filter, andOr)
			}
		}

		if c.TypeRust == "String" {
			if c.IsPartition {
				eqAdd("", andOr)
			}
			if c.IsClustering {
				eqAdd("", andOr)
			}
			if c.IsRegular {
				eqAdd(filter, andOr)
			}
		}
	}

	return
}

func (c *ColumnOut) GetRustModifiersIns() (res []WhereModifierIns) {
	add := func(m WhereModifierIns) {
		if len(m.AndOr) > 0 {
			m.FuncName = m.Prefix + "_" + c.ColumnNameRust + m.Suffix
		} else {
			m.FuncName = c.ColumnNameRust + m.Suffix
		}
		res = append(res, m)
	}
	inAdd := func(filter, andOr string) {
		add(WhereModifierIns{"_in" + filter, strings.ToLower(andOr), andOr, ""})
	}

	const filter = "_filtering"

	for _, andOr := range []string{"", "AND", "OR"} {
		if c.TypeRust == "i32" || c.TypeRust == "i64" ||
			c.TypeRust == "f32" || c.TypeRust == "f64" {
			if c.IsPartition {
				inAdd("", andOr)
			}
			if c.IsClustering {
				inAdd("", andOr)
			}
			if c.IsRegular {
				inAdd(filter, andOr)
			}
		}
		if c.TypeGo == "string" {
			if c.IsPartition {
				inAdd("", andOr)
			}
			if c.IsClustering {
				inAdd("", andOr)
			}
			if c.IsRegular {
				inAdd(filter, andOr)
			}
		}
	}

	return
}

// todo add suffix 'Go' + change in templates
func (c *ColumnOut) GetModifiers() (res []WhereModifier) {
	add := func(m WhereModifier) {
		if len(m.AndOr) > 0 {
			m.FuncName = m.AndOr + "_" + c.ColumnNameGO + m.Suffix
		} else {
			m.FuncName = c.ColumnNameGO + m.Suffix
		}
		res = append(res, m)
	}
	eqAdd := func(filter, andOr string) {
		//sufix := filter + andOr
		add(WhereModifier{"_Eq" + filter, andOr, "=", andOr, ""})
	}

	notEqs := func(filter, andOr string) {
		sufix := filter //+ andOr
		and := andOr
		add(WhereModifier{"_LT" + sufix, and, "<", andOr, ""})
		add(WhereModifier{"_LE" + sufix, and, "<=", andOr, ""})
		add(WhereModifier{"_GT" + sufix, and, ">", andOr, ""})
		add(WhereModifier{"_GE" + sufix, and, ">=", andOr, ""})
	}
	const filter = "_FILTERING"
	for _, andOr := range []string{"", "And", "Or"} {
		if c.TypeGo == "int" || c.TypeGo == "int64" {
			filter := "_Filtering"
			if c.IsPartition {
				eqAdd("", andOr)
				notEqs(filter, andOr)
			}
			if c.IsClustering {
				eqAdd("", andOr)
				notEqs("", andOr)
			}
			if c.IsRegular {
				eqAdd(filter, andOr)
				notEqs(filter, andOr)
			}
		}
		if c.TypeGo == "string" {
			if c.IsPartition {
				eqAdd("", andOr)
			}
			if c.IsClustering {
				eqAdd("", andOr)
			}
			if c.IsRegular {
				eqAdd(filter, andOr)
			}
		}
	}

	return
}

// not used yet; utlizing this a better alternative to current implemention.
func (c *ColumnOut) GetModifiersIns() (res []WhereModifierIns) {
	add := func(m WhereModifierIns) {
		if len(m.AndOr) > 0 {
			m.FuncName = m.AndOr + "_" + c.ColumnNameGO + m.Suffix
		} else {
			m.FuncName = c.ColumnNameGO + m.Suffix
		}
		res = append(res, m)
	}
	inAdd := func(filter, andOr string) {
		add(WhereModifierIns{"_In" + filter, andOr, andOr, ""})
	}

	const filter = "_FILTERING"

	for _, andOr := range []string{"", "And", "Or"} {
		if c.TypeGo == "int" {
			if c.IsPartition {
				inAdd("", andOr)
			}
			if c.IsClustering {
				inAdd("", andOr)
			}
			if c.IsRegular {
				inAdd(filter, andOr)
			}
		}
		if c.TypeGo == "string" {
			if c.IsPartition {
				inAdd("", andOr)
			}
			if c.IsClustering {
				inAdd("", andOr)
			}
			if c.IsRegular {
				inAdd(filter, andOr)
			}
		}
	}

	return
}

func (table *TableOut) ColumnNamesParams() string {
	var arr []string
	for _, t := range table.Columns {
		arr = append(arr, t.ColumnName)
	}
	return strings.Join(arr, ",")
}

func (col *ColumnOut) RustBorrowSign() string {
	o := ""
	switch col.TypeRust {
	case "String", "Vec<u8>":
		o = "&"
	case "i32", "i64", "f64", "f32", "&str":
		o = ""
	}
	return o
}

func (col *CColumn) IsNumber() bool {
	nums := []string{"int", "serial", "tinyint", "smallint", "bigint",
		"decimal", "float"}
	for i := 0; i < len(nums); i++ {
		if col.TypeCql == nums[i] {
			return true
		}
	}
	return false
}

// For sorting

type ColumnsSortable []*CColumn

func (a ColumnsSortable) Len() int            { return len(a) }
func (a ColumnsSortable) Swap(i, j int)       { a[i], a[j] = a[j], a[i] }
func (a ColumnsSortable) Less2(i, j int) bool { return a[i].Position > a[j].Position }
func (a ColumnsSortable) Less(i, j int) bool {
	if a[i].Position == a[j].Position {
		return a[i].ColumnName < a[j].ColumnName
	} else {
		return a[i].Position > a[j].Position
	}
}
