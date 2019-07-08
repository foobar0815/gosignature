package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"gopkg.in/ini.v1"
)

func main() {
	exe, err := os.Executable()
	checkErr(err)
	programPath := filepath.Dir(exe)

	configFile := flag.String("ini", "OutlookSignature.ini", "use alternative configuration file")
	testmode := flag.Bool("testmode", false, "run in test mode")
	flag.Parse()

	cfg, err := readConfig(filepath.Join(programPath, *configFile))
	checkErr(err)

	templateNames := make(map[string]string)
	if cfg.Section("Main").Key("FixedSignType").String() != "" {
		templateNames["signature"] = cfg.Section("Main").Key("FixedSignType").String()
	}
	if cfg.Section("Main").Key("FixedSignTypeReply").String() != "" {
		templateNames["signatureReply"] = cfg.Section("Main").Key("FixedSignTypeReply").String()
	}

	ldapEntry := make(map[string]string)
	if !(*testmode) && cfg.Section("Main").Key("LDAPBaseObjectDN").String() != "" {
		userName, err := getUsername()
		checkErr(err)
		ldapConnStrings := strings.Split(cfg.Section("Main").Key("LDAPBaseObjectDN").String(), ";")
		for i := 1; i <= len(ldapConnStrings); i++ {
			ldapServer := strings.Split(ldapConnStrings[i-1], "/")[0]
			ldapBaseDN := strings.Split(ldapConnStrings[i-1], "/")[1]

			conn, err := ldapConnect(ldapServer, cfg.Section("Main").Key("LDAPReaderAccountName").String(), cfg.Section("Main").Key("LDAPReaderAccountPassword").String())
			checkErr(err)

			ldapSearchresult, err := ldapSearch(conn,
				ldapBaseDN,
				cfg.Section("Main").Key("LDAPFilter").MustString("&(objectCategory=person)(objectClass=user)"),
				userName,
				cfg.Section("FieldMapping").KeysHash(),
			)
			checkErr(err)
			ldapEntry = ldapSearchToHash(ldapSearchresult)

			if len(ldapEntry) > 0 {
				if cfg.Section("Main").Key("FixedSignTypeForDN"+strconv.Itoa(i)).String() != "" {
					templateNames["signature"] = cfg.Section("Main").Key("FixedSignTypeForDN" + strconv.Itoa(i)).String()
				}
				if cfg.Section("Main").Key("FixedSignTypeReplyForDN"+strconv.Itoa(i)).String() != "" {
					templateNames["signature"] = cfg.Section("Main").Key("FixedSignTypeReplyForDN" + strconv.Itoa(i)).String()
				}
				break
			} else if i == len(ldapConnStrings) {
				checkErr(errors.New("user not found"))
			}
		}
	} else {
		ldapEntry = ldapFakeEntry()
	}

	if ldapEntry[cfg.Section("FieldMapping").Key("SignType").String()] != "" {
		templateNames["signature"] = ldapEntry[cfg.Section("FieldMapping").Key("SignType").String()]
	}
	if ldapEntry[cfg.Section("FieldMapping").Key("SignTypeReply").String()] != "" {
		templateNames["signatureReply"] = ldapEntry[cfg.Section("FieldMapping").Key("SignTypeReply").String()]
	}

	extensions := [3]string{"txt", "htm", "rtf"}
	generated := []string{}
	templateFolder := filepath.Join(programPath, cfg.Section("Main").Key("TemplateFolder").MustString("Vorlagen"))
	destFolder := getDestFolder()
	if len(templateNames) > 0 {
		for _, templateName := range templateNames {
			if !contains(generated, templateName) {
				copyFile(filepath.Join(templateFolder, templateName+".jpg"), filepath.Join(destFolder, templateName+".jpg"))
				for _, ex := range extensions {
					signature, err := readTemplate(filepath.Join(templateFolder, templateName+"."+ex))
					checkErr(err)
					signature = generateSignature(ldapEntry,
						cfg.Section("FieldMapping").KeysHash(),
						cfg.Section("Main").Key("PlaceholderSymbol").MustString("@"),
						signature)
					signature = replaceFullname(ldapEntry,
						cfg.Section("FieldMapping").KeysHash(),
						cfg.Section("Main").Key("PlaceholderSymbol").MustString("@"),
						signature)
					signature = replaceSigntitle(ldapEntry,
						cfg.Section("FieldMapping").KeysHash(),
						cfg.Section("Main").Key("PlaceholderSymbol").MustString("@"),
						signature, ex)
					err = writeSignature(destFolder, templateName, ex, signature)
					checkErr(err)
					generated = append(generated, templateName)
				}
			}
		}
	}
}

func readConfig(configFile string) (*ini.File, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		SpaceBeforeInlineComment: true,
	}, configFile)

	return cfg, err
}

func readTemplate(template string) (string, error) {

	read, err := ioutil.ReadFile(template)

	return string(read), err
}

func writeSignature(destFolder, templateName, extension, content string) error {
	fileName := templateName + "." + extension
	err := ioutil.WriteFile(filepath.Join(destFolder, fileName), []byte(content), 0644)

	return err
}

func generateSignature(ldapEntry, fieldMapping map[string]string, placeHolder, template string) string {
	for k, v := range fieldMapping {
		if v != "" {
			field := strings.ToUpper(placeHolder + k + placeHolder)
			template = strings.Replace(template, field, ldapEntry[v], -1)
		}
	}

	return template

}

func replaceFullname(ldapEntry, fieldMapping map[string]string, placeHolder, template string) string {
	fullname := []string{}
	fnFields := [4]string{"Title", "FirstName", "Initials", "LastName"}

	for _, field := range fnFields {
		if ldapEntry[fieldMapping[field]] != "" {
			fullname = append(fullname, ldapEntry[fieldMapping[field]])
		}
	}

	template = strings.Replace(template, placeHolder+"FULLNAME"+placeHolder, strings.Join(fullname, " "), -1)

	return template
}

func replaceSigntitle(ldapEntry, fieldMapping map[string]string, placeHolder, template, extension string) string {
	signtitle := []string{}
	stFields := [2]string{"SignTitle1", "SignTitle2"}
	for _, field := range stFields {
		if ldapEntry[fieldMapping[field]] != "" {
			signtitle = append(signtitle, ldapEntry[fieldMapping[field]])
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

	if extension == "rtf" {
		template, _ = charmap.ISO8859_1.NewEncoder().String(template)
	}

	return template

}
