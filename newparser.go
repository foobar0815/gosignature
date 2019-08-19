package main

import (
	"bytes"
	"text/template"
)

func newParser(fieldMap map[string]string, templateName, tmpl, extension string) string {
	buf := new(bytes.Buffer)
	t, err := template.New(templateName).Delims("[[", "]]").Parse(tmpl)
	checkErr(err)
	err = t.Execute(buf, fieldMap)
	tmpl = buf.String()

	return tmpl
}
