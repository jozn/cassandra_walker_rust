package main

import (
	"strings"
	"text/template"

	"github.com/jozn/xox/snaker"
)

func NewTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"ms_to_slice": ms_to_slice,
	}
}

