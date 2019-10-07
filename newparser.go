package main

import (
	"bytes"
	"text/template"
)

func newParser(fieldMap map[string]string, templateName, tmpl string) (string, error) {
	buf := new(bytes.Buffer)
	t, err := template.New(templateName).Delims("[[", "]]").Parse(tmpl)
	if err != nil {
		return "", err
	}
	err = t.Execute(buf, fieldMap)
	if err != nil {
		return "", err
	}
	tmpl = buf.String()

	return tmpl, nil

}
