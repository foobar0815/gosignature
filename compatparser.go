package main

import (
	"strings"

	"golang.org/x/text/encoding/charmap"
)

func compatParser(fieldMap map[string]string, placeHolder, template, ex string) string {
	template = replaceFields(fieldMap, placeHolder, template)
	template = replaceFullname(fieldMap, placeHolder, template)
	template = replaceSigntitle(fieldMap, placeHolder, template, ex)

	return template
}

func replaceFields(fieldMap map[string]string, placeHolder, template string) string {
	var err error
	for k, v := range fieldMap {
		v, err = charmap.Windows1252.NewEncoder().String(v)
		checkErrAndContinue(err)
		template = strings.Replace(template, placeHolder+strings.ToUpper(k)+placeHolder, v, -1)
	}

	return template
}

func replaceFullname(fieldMap map[string]string, placeHolder, template string) string {
	fullname := []string{}
	fnFields := [4]string{"Title", "FirstName", "Initials", "LastName"}

	for _, field := range fnFields {
		if fieldMap[field] != "" {
			fullname = append(fullname, fieldMap[field])
		}
	}

	template = strings.Replace(template, placeHolder+"FULLNAME"+placeHolder, strings.Join(fullname, " "), -1)

	return template
}

func replaceSigntitle(fieldMap map[string]string, placeHolder, template, extension string) string {
	signtitle := []string{}
	stFields := [2]string{"SignTitle1", "SignTitle2"}
	for _, field := range stFields {
		if fieldMap[field] != "" {
			signtitle = append(signtitle, fieldMap[field])
		}
	}
	newline := ""
	switch extension {
	case "txt":
		newline = "\n"
	case "rtf":
		newline = "\\line\n"
	case "htm":
		newline = "<br>"
	}
	template = strings.Replace(template, placeHolder+"SIGNTITLE"+placeHolder, strings.Join(signtitle, newline), -1)

	return template

}
