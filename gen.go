package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"text/template"

	bind "github.com/jozn/cassandra-walker/template_bind"
)

func build(gen *GenOut) {
	writeOutput("xc_models.go", buildFromTemplate("models_types.tpl.go", gen))
	writeOutput("xc_common.go", buildFromTemplate("common.tpl.go", gen))

	for _, t := range gen.Tables {
		fileName := fmt.Sprintf("%s.go", t.TableName)
		writeOutput(fileName, buildFromTemplate("model.tpl.go", t))
	}

	if true {
		dirOut := path.Join(args.Dir, args.Package)
		e1 := exec.Command("gofmt", "-w", dirOut).Run()
		e2 := exec.Command("goimports", "-w", dirOut).Run()
		errLog("gofmt", e1)
		errLog("goimports", e2)
	}
}

func writeOutput(fileName, output string) {
	dirOut := path.Join(args.Dir, args.Package)
	err := os.MkdirAll(dirOut, os.ModePerm)
	NoErr(err)
	file := path.Join(dirOut, fileName)

	err = ioutil.WriteFile(file, []byte(output), os.ModePerm)
	NoErr(err)
}

func buildFromTemplate(tplName string, gen interface{}) string {
	tpl := template.New("" + tplName)
	tpl.Funcs(NewTemplateFuncs())

	tplGoInterface, err := bind.Asset(tplName)
	NoErr(err)
	tpl, err = tpl.Parse(string(tplGoInterface))
	NoErr(err)

	buffer := bytes.NewBufferString("")
	err = tpl.Execute(buffer, gen)
	NoErr(err)

	return buffer.String()
}
