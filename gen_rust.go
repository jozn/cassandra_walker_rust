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

		// this columns can not be not set so do not check for them
		if col.IsPartition || col.IsClustering {
			T = `
		// partition key and clustering key always must be present
		columns.push("{{.ColumnName}}");
        values.push(self.{{.ColumnName}}.clone().into());
`
		} else { // regular coulmn could be not set
			switch col.TypeRust {
			case "String", "&str":
				T = `
		if !self.{{.ColumnNameRust}}.is_empty() {
            columns.push("{{.ColumnName}}");
            values.push(self.{{.ColumnName}}.clone().into());
       	}
`
				// type blob in cassandra: cdrs driver somehow corrupt vec<u8> when into() is called
				// it produces a vec<u8> of much bigger size with different values. This code is a
				// work around this bug/implementaion
			case "Vec<u8>":
				T = `
		if !self.{{.ColumnNameRust}}.is_empty() {
            let val = Value{
                body: self.{{.ColumnName}}.clone(),
                value_type: ValueType::Normal(self.{{.ColumnName}}.len() as i32)
            };

            columns.push("{{.ColumnName}}");
            values.push(val);
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

		}
		fnStr := rawTemplateOutput(T, col)
		fnsOut = append(fnsOut, fnStr)
	}
	out := strings.Join(fnsOut, "")
	return out
}

// Pirmary Getters
type _RustGetter_ struct {
	fnName    string
	paramName string
	paramType string
	callName  string
}

// This produces Getter by primary key [whether this functionality was worth the effort is a good question]
func (table *TableOut) GetRustPrimaryGetter() string {
	//============== Fill Array ===============
	arr := []_RustGetter_{}
	paramCnt := 1
	// Partion keys
	for i := 0; i < len(table.PartitionColumns); i++ {
		col := table.Columns[i]

		param := fmt.Sprintf("p%d", paramCnt)
		fnName := col.ColumnNameRust
		callName := fmt.Sprintf("%s_eq(%s)", col.ColumnNameRust, param)
		if i > 0 {
			fnName = fmt.Sprintf("_and_%s", col.ColumnNameRust)
			callName = fmt.Sprintf("and_%s_eq(%s)", col.ColumnNameRust, param)
		}

		f := _RustGetter_{
			fnName:    fnName,
			paramName: param,
			paramType: col.TypeRust,
			callName:  callName,
		}
		arr = append(arr, f)
		paramCnt += 1
	}

	// Clustering keys
	for i := 0; i < len(table.ClusterColumns); i++ {
		col := table.Columns[i]

		param := fmt.Sprintf("p%d", paramCnt)
		f := _RustGetter_{
			fnName:    fmt.Sprintf("_and_%s", col.ColumnNameRust),
			paramName: fmt.Sprintf("p%d", paramCnt),
			paramType: col.TypeRust,
			callName:  fmt.Sprintf("and_%s_eq(%s)", col.ColumnNameRust, param),
		}
		arr = append(arr, f)
		paramCnt += 1
	}

	//================ Make Str Output =================
	fnName := fmt.Sprintf("get_%s_by_", table.TableName) //todo
	fnParam := []string{}
	fnSetter := fmt.Sprintf("%s_Selector::new()", table.TableNameRust)
	for i := 0; i < len(arr); i++ {
		f := arr[i]
		fnName += f.fnName
		fnParam = append(fnParam, f.paramName+":"+f.paramType)
		fnSetter += "\n\t\t." + f.callName
	}

	//================ Build ==================
	const TPL = `
pub fn {{ .FnName }}(session: impl FCQueryExecutor, {{ .FnParam }}) -> Result<{{.Table.TableNameRust}},CWError> {
	let m = {{ .FnSetter }}
		.get_row(session)?;
	Ok(m)
}
`
	parm := struct {
		FnName   string
		FnParam  string
		FnSetter string
		Table    *TableOut
	}{
		FnName:   fnName,
		FnParam:  strings.Join(fnParam, ","),
		FnSetter: fnSetter,
		Table:    table,
	}
	out := rawTemplateOutput(TPL, parm)

	return out

	/*
		//=============== Make fn name ===============
		fnName := fmt.Sprintf("get_%s_by_", table.TableNameRust) //todo
		for i := 0; i < len(table.PartitionColumns); i++ {
			col := table.Columns[i]
			if i == 0 {
				fnName = fnName + col.ColumnNameRust
			}else {
				fnName = fnName +"_and" + col.ColumnNameRust
			}
		}
		// Clustering
		for i := 0; i < len(table.ClusterColumns); i++ {
			col := table.Columns[i]
			fnName = fnName +"_and" + col.ColumnNameRust
		}

		//=============== Make fn param ===============
		var fnParamArr []string
		paramCnt := 1
		for i := 0; i < len(table.PartitionColumns); i++ {
			col := table.Columns[i]
			s := fmt.Sprintf("p%d :%s", paramCnt, col.TypeRust)
			fnParamArr = append(fnParamArr, s)
			paramCnt+=1
		}
		for i := 0; i < len(table.ClusterColumns); i++ {
			col := table.Columns[i]
			s := fmt.Sprintf("p%d :%s", paramCnt, col.TypeRust)
			fnParamArr = append(fnParamArr, s)
			paramCnt+=1
		}
		fnParamStr := fmt.Sprintf("(%s)",strings.Join(fnParamArr,","))

		// Clustering
		for i := 0; i < len(table.ClusterColumns); i++ {
			col := table.Columns[i]
			fnName = fnName +"_and" + col.ColumnNameRust
		}
	*/

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
