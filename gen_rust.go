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

	writeOutput("xc_models.rs", buildFromTemplate("models_types.rs", gen))
	writeOutput("common.rs", buildFromTemplate("common.rs", gen))

	//writeOutput("xc_common.go", buildFromTemplate("common.tgo", gen))

	for _, t := range gen.Tables {
		fileName := fmt.Sprintf("%s.rs", t.TableName)
		writeOutput(fileName, buildFromTemplate("model.rs", t))

		t.GetRustWheresTmplOut()
	}

	if true {
		dirOut := path.Join(args.Dir, args.Package)
		e1 := exec.Command("cargo fmt", dirOut).Run()
		errLog("gofmt", e1)
	}
}

func (table *TableOut) GetRustWheresTmplOut() string {
	const FN = `
    pub fn {{ .Mod.FuncName }} (&mut self, val: {{ .Col.TypeRustBorrow }} ) ->&Self {
        let w = WhereClause{
            condition: "{{ .Mod.AndOr }} {{ .Col.ColumnNameRust }} {{ .Mod.Condition }} ?",
            args: val.into(),
        };
        self.wheres.push(w);
        self
    }
`

	const FN2 = `
    pub fn {{.Mod.FuncName}}(&mut self, val: &str) ->&Self {
        let w = WhereClause{
            condition: "OR tweet_id >= ?",
            args: val.into(),
        };
        self.wheres.push(w);
        self
    }
`
	fnsOut := []string{}

	// parse template
	tpl := template.New("fns" )
	tpl, err := tpl.Parse(FN)
	NoErr(err)

	for i:=0; i< len(table.Columns); i++ {
		col := table.Columns[i]

		for j := 0; j < len(col.WhereModifiersRust); j++ {
			wmr := col.WhereModifiersRust[j]

			parm := struct {
				Table *TableOut
				Mod WhereModifier
				Col *ColumnOut
			}{
				table, wmr, col,
			}

			buffer := bytes.NewBufferString("")
			err = tpl.Execute(buffer, parm)

			fnStr := buffer.String()
			fmt.Println(fnStr)
			fnsOut = append(fnsOut,fnStr )

		}
	}

	return strings.Join(fnsOut, "")
}



////////////////// Shared with Go generator /////////////
func writeOutput(fileName, output string) {
	dirOut := path.Join(args.Dir, args.Package)
	fmt.Println(dirOut)
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
