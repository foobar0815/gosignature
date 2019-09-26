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
	newparser := flag.Bool("newparser", false, "use the new template parser")
	flag.Parse()

	cfg, err := readConfig(filepath.Join(programPath, *configFile))
	checkErr(err)

	templateNames := make(map[string]map[string]string)
	templateNames["signature"] = make(map[string]string)
	templateNames["signatureReply"] = make(map[string]string)
	templateNames["signature"]["src"] = cfg.Section("Main").Key("FixedSignType").String()
	templateNames["signatureReply"]["src"] = cfg.Section("Main").Key("FixedSignTypeReply").String()

	userName := ""
	ldapEntry := make(map[string]string)
	if !(*testmode) && cfg.Section("Main").Key("LDAPBaseObjectDN").String() != "" {
		userName, err = getUsername()
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
				templateNames["signature"]["src"] = cfg.Section("Main").Key("FixedSignTypeForDN" + strconv.Itoa(i)).MustString(templateNames["signature"]["src"])
				templateNames["signatureReply"]["src"] = cfg.Section("Main").Key("FixedSignTypeReplyForDN" + strconv.Itoa(i)).MustString(templateNames["signatureReply"]["src"])
				break
			} else if i == len(ldapConnStrings) {
				checkErr(errors.New("user not found"))
			}
		}
	} else {
		userName = "hkearny"
		ldapEntry = ldapFakeEntry()
	}

	fieldMap := mapFields(ldapEntry, cfg.Section("FieldMapping").KeysHash())

	if fieldMap["SignType"] != "" {
		templateNames["signature"]["src"] = fieldMap["SignType"]
	}
	if fieldMap["SignTypeReply"] != "" {
		templateNames["signatureReply"]["src"] = fieldMap["SignTypeReply"]
	}

	if templateNames["signature"]["src"] != "" || templateNames["signatureReply"]["src"] != "" {
		extensions := [3]string{"txt", "htm", "rtf"}
		generated := []string{}
		templateFolder := filepath.Join(programPath, cfg.Section("Main").Key("TemplateFolder").MustString("Vorlagen"))
		destFolder := getDestFolder()
		err = prepareFolder(destFolder)
		checkErr(err)
		templateNames["signature"]["dst"] = cfg.Section("Main").Key("TargetSignType").MustString(templateNames["signature"]["src"])
		templateNames["signatureReply"]["dst"] = cfg.Section("Main").Key("TargetSignTypeReply").MustString(templateNames["signatureReply"]["src"])
		for _, templateName := range templateNames {
			if templateName["src"] != "" && !contains(generated, templateName["dst"]) {
				copyImages(templateFolder, templateName["src"], templateName["dst"], userName, destFolder)
				for _, ex := range extensions {
					signature, err := readTemplate(filepath.Join(templateFolder, templateName["src"]+"."+ex))
					checkErr(err)
					if *newparser {
						signature = newParser(fieldMap, templateName["src"], signature, ex)
					} else {
						signature = compatParser(fieldMap,
							cfg.Section("Main").Key("PlaceholderSymbol").MustString("@"),
							signature,
							ex)
					}
					if ex == "rtf" || ex == "txt" {
						signature, _ = charmap.Windows1252.NewEncoder().String(signature)
					}
					err = writeSignature(destFolder, templateName["dst"], ex, signature)
					checkErr(err)
					generated = append(generated, templateName["dst"])
				}
			}
		}
		setSignature(templateNames["signature"]["dst"],
			templateNames["signatureReply"]["dst"],
			cfg.Section("Main").Key("EMailAccount").MustString(""),
			cfg.Section("Main").Key("SetForAllEMailAccounts").MustInt(0),
			cfg.Section("Main").Key("NoNewMessageSignature").MustInt(0),
			cfg.Section("Main").Key("NoReplyMessageSignature").MustInt(0))
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

func prepareFolder(destFolder string) error {
	err := os.MkdirAll(destFolder, os.ModePerm)

	return err
}

func writeSignature(destFolder, templateName, extension, content string) error {
	fileName := templateName + "." + extension

	err := ioutil.WriteFile(filepath.Join(destFolder, fileName), []byte(content), os.ModePerm)

	return err
}

func mapFields(ldapEntry, fieldMapping map[string]string) map[string]string {
	m := make(map[string]string)
	for k, v := range fieldMapping {
		if len(v) > 0 {
			m[k] = ldapEntry[v[1:]]
		}
	}

	return m
}

func copyImages(templateFolder, srcName, dstName, userName, destFolder string) {
	extensions := [3]string{"gif", "jpg", "png"}

	for _, extension := range extensions {
		images, _ := filepath.Glob(filepath.Join(templateFolder, srcName+"*."+extension))

		for _, image := range images {
			copyFile(image, filepath.Join(destFolder, dstName+filepath.Ext(image)))
		}
	}

	for _, extension := range extensions {
		images, _ := filepath.Glob(filepath.Join(templateFolder, userName+"."+extension))

		for _, image := range images {
			copyFile(image, filepath.Join(destFolder, "portrait"+filepath.Ext(image)))
		}
	}

}
