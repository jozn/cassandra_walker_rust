package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

func buildRust(gen *GenOut) {
	//buildGo(gen) // temp

	writeOutput("xc_models.rs", buildFromTemplate("rust/models_types.rs", gen))
	writeOutput("common.rs", buildFromTemplate("rust/common.rs", gen))
	writeOutput("mod.rs", buildFromTemplate("rust/mod.rs", gen))

	//writeOutput("xc_common.go", buildFromTemplate("common.tgo", gen))

	for _, t := range gen.Tables {
		fileName := fmt.Sprintf("%s.rs", t.TableName)
		writeOutput(fileName, buildFromTemplate("rust/model.rs", t))

		t.GetRustWheresTmplOut()
	}

	if true {
		dirOut := strings.Replace(args.Dir, "src/", "", -1)
		e1 := os.Chdir(dirOut)
		e1 = exec.Command("cargo fmt").Run()
		errLog("cargo fmt", e1)
	}
}

func (table *TableOut) GetRustWheresTmplOut() string {
	const TPL = `
    pub fn {{ .Mod.FuncName }} (&mut self, val: {{ .Col.TypeRustBorrow }} ) -> &mut Self {
        let w = WhereClause{
            condition: "{{ .Mod.AndOr }} {{ .Col.ColumnNameRust }} {{ .Mod.Condition }} ?".to_string(),
            args: val.into(),
        };
        self.wheres.push(w);
        self
    }
`

	fnsOut := []string{}

	// parse template
	tpl := template.New("fns")
	tpl, err := tpl.Parse(TPL)
	NoErr(err)

	for i := 0; i < len(table.Columns); i++ {
		col := table.Columns[i]

		for j := 0; j < len(col.WhereModifiersRust); j++ {
			wmr := col.WhereModifiersRust[j]

			parm := struct {
				Table *TableOut
				Mod   WhereModifier
				Col   *ColumnOut
			}{
				table, wmr, col,
			}

			buffer := bytes.NewBufferString("")
			err = tpl.Execute(buffer, parm)

			fnStr := buffer.String()
			//fmt.Println(fnStr)
			fnsOut = append(fnsOut, fnStr)
		}
	}

	return strings.Join(fnsOut, "")
}

func (table *TableOut) GetRustWhereInsTmplOut() string {
	const TPL = `
    pub fn {{ .Mod.FuncName }} (&mut self, val: Vec<{{ .Col.TypeRustBorrow }}> ) -> &mut Self {
		let len = val.len();
        if len == 0 {
            return self
        }

        let mut marks = "?,".repeat(len);
        marks.remove(marks.len()-1);
        let w = WhereClause{
			condition: format!("{{ .Mod.AndOr }} {{ .Col.ColumnNameRust }} IN ({})", marks),
            args: val.into(),
        };
        self.wheres.push(w);
        self
    }
`
	fnsOut := []string{}

	// parse template
	tpl := template.New("fns")
	tpl, err := tpl.Parse(TPL)
	NoErr(err)

	for i := 0; i < len(table.Columns); i++ {
		col := table.Columns[i]

		for j := 0; j < len(col.WhereInsModifiersRust); j++ {
			wmr := col.WhereInsModifiersRust[j]

			parm := struct {
				Table *TableOut
				Mod   WhereModifierIns
				Col   *ColumnOut
			}{
				table, wmr, col,
			}

			buffer := bytes.NewBufferString("")
			err = tpl.Execute(buffer, parm)

			fnStr := buffer.String()
			//fmt.Println(fnStr)
			fnsOut = append(fnsOut, fnStr)
		}
	}

	return strings.Join(fnsOut, "")
}

// Updater
func (table *TableOut) GetRustUpdaterFnsOut() string {
	const TPL = `
    pub fn update_{{ .Col.ColumnNameRust }}(&mut self, val: {{ .Col.TypeRustBorrow }}) -> &mut Self {
        self.updates.insert("{{ .Col.ColumnName }} = ?", val.into());
        self
    }
`

	const TPL_BLOB = `
    pub fn update_{{ .Col.ColumnNameRust }}(&mut self, val: {{ .Col.TypeRustBorrow }}) -> &mut Self {
        self.updates.insert("{{ .Col.ColumnName }} = ?", Blob::new(val.clone()).into());
        self
    }
`
	fnsOut := []string{}

	for i := 0; i < len(table.Columns); i++ {
		col := table.Columns[i]

		parm := struct {
			Table *TableOut
			Col   *ColumnOut
		}{
			table, col,
		}

		var fnStr string

		// Due to cdrs lib limitation we should treat blob differently
		if col.TypeCql == "blob" {
			fnStr = rawTemplateOutput(TPL_BLOB, parm)
		} else {
			fnStr = rawTemplateOutput(TPL, parm)
		}

		//fmt.Println(fnStr)
		fnsOut = append(fnsOut, fnStr)
	}

	return strings.Join(fnsOut, "")
}

// Selectors
func (table *TableOut) GetRustSelectorOrders() string {
	const TPL = `
    pub fn order_by_{{ .Col.ColumnNameRust }}_asc(&mut self) -> &mut Self {
		self.order_by.push("{{ .Col.ColumnName }} ASC");
        self
    }

	pub fn order_by_{{ .Col.ColumnNameRust }}_desc(&mut self) -> &mut Self {
		self.order_by.push("{{ .Col.ColumnName }} DESC");
        self
    }
`
	fnsOut := []string{}

	for i := 0; i < len(table.Columns); i++ {
		col := table.Columns[i]
		if col.IsClustering { //&& col.IsNumber()
			parm := struct {
				Table *TableOut
				Col   *ColumnOut
			}{
				table, col,
			}

			fnStr := rawTemplateOutput(TPL, parm)
			//fmt.Println(fnStr)
			fnsOut = append(fnsOut, fnStr)
		}
	}

	return strings.Join(fnsOut, "")
}

// Models (save, delete, update)

func (table *TableOut) GetRustModelSavePartial() string {
	fnsOut := []string{}

	for i := 0; i < len(table.Columns); i++ {
		col := table.Columns[i]

		T := ""
		switch col.TypeRust {
		case "String", "&str", "Vec<u8>":
			T = `
		if !self.{{.ColumnNameRust}}.is_empty() {
            columns.push("{{.ColumnName}}");
            values.push(self.{{.ColumnName}}.clone().into());
       	}
`
		default:
			T = `
		if self.{{.ColumnNameRust}} != {{.TypeDefaultRust}} {
            columns.push("{{.ColumnName}}");
            values.push(self.{{.ColumnName}}.clone().into());
       	}
`
		}
		fnStr := rawTemplateOutput(T, col)
		fnsOut = append(fnsOut, fnStr)
	}
	out := strings.Join(fnsOut, "")
	return out
}

// Utils - not used
func eachColumn(table *TableOut, tpl string) string {
	fnsOut := []string{}

	for i := 0; i < len(table.Columns); i++ {
		col := table.Columns[i]

		parm := struct {
			Table *TableOut
			Col   *ColumnOut
		}{
			Table: table,
			Col:   col,
		}

		fnStr := rawTemplateOutput(tpl, parm)
		//fmt.Println(fnStr)
		fnsOut = append(fnsOut, fnStr)
	}

	return strings.Join(fnsOut, "")
}

func rawTemplateOutput(templ string, data interface{}) string {
	tpl := template.New("fns")
	tpl, err := tpl.Parse(templ)
	NoErr(err)

	buffer := bytes.NewBufferString("")
	err = tpl.Execute(buffer, data)
	NoErr(err)
	outPut := buffer.String()
	return outPut
}

////////////////// Shared with Go generator /////////////
func writeOutput(fileName, output string) {
	dirOut := path.Join(args.Dir, args.Package)
	//fmt.Println(dirOut)
	err := os.MkdirAll(dirOut, os.ModePerm)
	NoErr(err)
	file := path.Join(dirOut, fileName)

	err = ioutil.WriteFile(file, []byte(output), os.ModePerm)
	NoErr(err)
}

func buildFromTemplate(tplName string, gen interface{}) string {
	tpl := template.New("" + tplName)
	tpl.Funcs(NewTemplateFuncs())

	tplGoInterface, err := Asset("templates/" + tplName) // Asset form bind_template
	NoErr(err)
	tpl, err = tpl.Parse(string(tplGoInterface))
	NoErr(err)

	buffer := bytes.NewBufferString("")
	err = tpl.Execute(buffer, gen)
	NoErr(err)

	return buffer.String()
}
