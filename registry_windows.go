package main

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

func getOutlookVersion() (string, error) {
	k, err := registry.OpenKey(registry.CLASSES_ROOT, `Outlook.Application\CurVer`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	s, _, err := k.GetStringValue("")
	if err != nil {
		return "", err
	}

	ssplit := strings.Split(s, ".")
	version := ssplit[len(ssplit)-1] + ".0"

	return version, nil
}
