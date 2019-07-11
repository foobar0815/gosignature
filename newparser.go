package main

import (
	"bytes"
	"strings"
	"text/template"

	"golang.org/x/text/encoding/charmap"
)

func newParser(fieldMap map[string]string, templateName, tmpl, extension string) string {
	// HACK?
	if extension == "rtf" {
		tmpl = strings.Replace(tmpl, "\\{\\{", "{{", -1)
		tmpl = strings.Replace(tmpl, "\\}\\}", "}}", -1)
	}

	buf := new(bytes.Buffer)
	t, err := template.New(templateName).Parse(tmpl)
	checkErr(err)
	err = t.Execute(buf, fieldMap)
	tmpl = buf.String()

	if extension == "rtf" {
		tmpl, _ = charmap.ISO8859_1.NewEncoder().String(tmpl)
	}

	return tmpl
}
