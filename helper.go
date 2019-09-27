package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
)

func contains(slice []string, searchString string) bool {
	for _, value := range slice {
		if value == searchString {
			return true
		}
	}
	return false
}

func checkErr(err error) {
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
}

func copyFile(sourceFile, destinationFile string) error {
	input, err := ioutil.ReadFile(filepath.FromSlash(sourceFile))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.FromSlash(destinationFile), input, os.ModePerm)
	if err != nil {
		return err
	}

	return err
}

func getUsername() (string, error) {
	username := ""
	user, err := user.Current()

	if runtime.GOOS == "windows" {
		username = os.Getenv("username")
	} else {
		username = user.Name
	}

	return username, err

}

func getDestFolder() string {
	destFolder := ""
	err := *new(error)

	if runtime.GOOS == "windows" {
		destFolder = filepath.Join(os.Getenv("appdata"), "Microsoft", "Signatures")
	} else {
		destFolder, err = os.Getwd()
		checkErr(err)
	}

	return destFolder

}

func escapeUnicode(s string) string {
	convertedString := ""

	for _, r := range s {
		if r > 127 {
			convertedString += "\\u" + strconv.Itoa(int(r)) + "?"
		} else {
			convertedString += string(r)
		}
	}

	return convertedString
}

func removeContents(dir string) error {
	items, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}

	for _, item := range items {
		err = os.RemoveAll(item)
		if err != nil {
			return err
		}
	}

	return nil
}
