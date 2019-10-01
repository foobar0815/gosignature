package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func contains(slice []string, searchString string) bool {
	for _, value := range slice {
		if value == searchString {
			return true
		}
	}
	return false
}

func checkErrAndExit(err error) {
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
}

func checkErrAndContinue(err error) {
	if err != nil {
		log.Print("WARNING: ", err)
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
	if runtime.GOOS == "windows" {
		return os.Getenv("username"), nil
	}

	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return user.Name, nil
}

func getDestFolder() string {
	destFolder := ""

	if runtime.GOOS == "windows" {
		destFolder = filepath.Join(os.Getenv("appdata"), "Microsoft", "Signatures")
	} else {
		destFolder, _ = os.Getwd()
	}

	return destFolder

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

func winExpandEnv(path string) string {
	re := regexp.MustCompile(`%[^\%]+%`)

	compatPath := re.ReplaceAllStringFunc(path, func(match string) string {
		match = strings.Replace(match, "%", "", -1)
		match = "${" + match + "}"
		return match
	})

	return os.ExpandEnv(compatPath)
}

func askForConfirmation(question string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", question)

		response, _ := reader.ReadString('\n')

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
