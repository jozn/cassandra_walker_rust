package main

import "text/template"

func NewTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"ms_to_slice": ms_to_slice,
	}
}

func ms_to_slice(typ ...string) []string {
	return typ
}
