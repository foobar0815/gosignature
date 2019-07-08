package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
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

	err = ioutil.WriteFile(filepath.FromSlash(destinationFile), input, 0644)
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
