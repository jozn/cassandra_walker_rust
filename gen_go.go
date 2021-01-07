package main

import (
	"fmt"
	"os/exec"
	"path"
)

func buildGo(gen *GenOut) {
	writeOutput("xc_models.go", buildFromTemplate("go/models_types.tgo", gen))
	writeOutput("xc_common.go", buildFromTemplate("go/common.tgo", gen))

	for _, t := range gen.Tables {
		fileName := fmt.Sprintf("%s.go", t.TableName)
		writeOutput(fileName, buildFromTemplate("go/model.tgo", t))
	}

	if true {
		dirOut := path.Join(args.Dir, args.Package)
		e1 := exec.Command("gofmt", "-w", dirOut).Run()
		e2 := exec.Command("goimports", "-w", dirOut).Run()
		errLog("gofmt", e1)
		errLog("goimports", e2)
	}
}
