package main

import (
	"bytes"
	"strings"
	"text/template"
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

	return tmpl
}
