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

type signatureDefinition struct {
	templateName  string
	signatureName string
	style         string
	nodefault     int
}

type ldapConnectionProfile struct {
	server   string
	userdn   string
	password string
}

type ldapSearchCriteria struct {
	basedn    string
	filter    string
	userfield string
	fieldmap  map[string]string
}

func main() {
	exe, err := os.Executable()
	checkErrAndExit(err)
	programPath := filepath.Dir(exe)

	configFile := flag.String("ini", "OutlookSignature.ini", "use alternative configuration file")
	testmode := flag.Bool("testmode", false, "run in test mode")
	newparser := flag.Bool("newparser", false, "use the new template parser")
	force := flag.Bool("force", false, "empty destination directory without confirmation")
	userName := flag.String("username", "", "generate signature for another user")
	flag.Parse()

	cfg, err := readConfig(filepath.Join(programPath, *configFile))
	checkErrAndExit(err)

	signatureDefintions := []*signatureDefinition{}

	sd := new(signatureDefinition)
	sd.templateName = cfg.Section("Main").Key("FixedSignType").String()
	sd.style = "new"
	sd.nodefault = cfg.Section("Main").Key("NoNewMessageSignature").MustInt(0)
	signatureDefintions = append(signatureDefintions, sd)

	sd = new(signatureDefinition)
	sd.templateName = cfg.Section("Main").Key("FixedSignTypeReply").String()
	sd.style = "reply"
	sd.nodefault = cfg.Section("Main").Key("NoReplyMessageSignature").MustInt(0)
	signatureDefintions = append(signatureDefintions, sd)

	sd = new(signatureDefinition)
	sd.templateName = cfg.Section("Main").Key("FixedSignTypeNoMobile").String()
	sd.style = "new"
	sd.nodefault = 1
	sd.signatureName = sd.templateName
	signatureDefintions = append(signatureDefintions, sd)

	sd = new(signatureDefinition)
	sd.templateName = cfg.Section("Main").Key("FixedSignTypeReplyNoMobile").String()
	sd.style = "reply"
	sd.nodefault = 1
	sd.signatureName = sd.templateName
	signatureDefintions = append(signatureDefintions, sd)

	ldapEntry := make(map[string]string)
	if !(*testmode) && cfg.Section("Main").Key("LDAPBaseObjectDN").String() != "" {

		if *userName == "" {
			*userName, err = getUsername()
			checkErrAndExit(err)
		}

		lcp := new(ldapConnectionProfile)
		lcp.userdn = cfg.Section("Main").Key("LDAPReaderAccountName").String()
		lcp.password = cfg.Section("Main").Key("LDAPReaderAccountPassword").String()

		lsc := new(ldapSearchCriteria)
		lsc.fieldmap = cfg.Section("FieldMapping").KeysHash()
		lsc.filter = cfg.Section("Main").Key("LDAPFilter").MustString("&(objectCategory=person)(objectClass=user)")
		lsc.userfield = cfg.Section("Main").Key("LDAPUserFieldname").MustString("sAMAccountName")

		ldapConnStrings := strings.Split(cfg.Section("Main").Key("LDAPBaseObjectDN").String(), ";")
		for i := 1; i <= len(ldapConnStrings); i++ {
			lcp.server = strings.Split(ldapConnStrings[i-1], "/")[0]
			lsc.basedn = strings.Split(ldapConnStrings[i-1], "/")[1]

			conn, err := ldapConnect(lcp)
			checkErrAndExit(err)

			ldapSearchresult, err := ldapSearch(conn, lsc, *userName)
			conn.Close()
			checkErrAndExit(err)
			ldapEntry = ldapSearchToHash(ldapSearchresult)

			if len(ldapEntry) > 0 {
				signatureDefintions[0].templateName = cfg.Section("Main").Key("FixedSignTypeForDN" + strconv.Itoa(i)).MustString(signatureDefintions[0].templateName)
				signatureDefintions[1].templateName = cfg.Section("Main").Key("FixedSignTypeReplyForDN" + strconv.Itoa(i)).MustString(signatureDefintions[1].templateName)
				break
			} else if i == len(ldapConnStrings) {
				checkErrAndExit(errors.New("user not found"))
			}
		}
	} else {
		ldapEntry = ldapFakeEntry()
		*userName = ldapEntry["sAMAccountName"]
	}

	fieldMap := mapFields(ldapEntry, cfg.Section("FieldMapping").KeysHash())

	if fieldMap["SignType"] != "" {
		signatureDefintions[0].templateName = fieldMap["SignType"]
	}
	if fieldMap["SignTypeReply"] != "" {
		signatureDefintions[1].templateName = fieldMap["SignTypeReply"]
	}

	extensions := [3]string{"txt", "htm", "rtf"}
	generated := []string{}
	templateFolder := filepath.Join(programPath, cfg.Section("Main").Key("TemplateFolder").MustString("Vorlagen"))
	destFolder := ""
	if cfg.Section("Main").Key("AppDataPath").String() != "" {
		destFolder = winExpandEnv(cfg.Section("Main").Key("AppDataPath").String())
	} else {
		destFolder = getDestFolder()
	}
	err = prepareFolder(destFolder)
	checkErrAndExit(err)
	if cfg.Section("Main").Key("EmptySignatureFolder").MustInt(0) == 1 && (*force || askForConfirmation("Do you really want to empty the destination directory ("+destFolder+")?")) {
		removeContents(destFolder)
	}
	signatureDefintions[0].signatureName = cfg.Section("Main").Key("TargetSignType").MustString(signatureDefintions[0].templateName)
	signatureDefintions[1].signatureName = cfg.Section("Main").Key("TargetSignTypeReply").MustString(signatureDefintions[1].templateName)
	for _, sd := range signatureDefintions {
		if sd.templateName != "" {
			if !contains(generated, sd.signatureName) {
				copyImages(templateFolder, sd.templateName, sd.signatureName, *userName, destFolder)
				for _, ex := range extensions {
					signature, err := readTemplate(filepath.Join(templateFolder, sd.templateName+"."+ex))
					checkErrAndExit(err)
					if *newparser {
						signature, err = newParser(fieldMap, sd.templateName, signature)
						checkErrAndExit(err)
					} else {
						signature = compatParser(fieldMap,
							cfg.Section("Main").Key("PlaceholderSymbol").MustString("@"),
							signature,
							ex)
					}
					if ex == "rtf" || ex == "txt" {
						signature, err = charmap.Windows1252.NewEncoder().String(signature)
						checkErrAndContinue(err)
					}
					err = writeSignature(destFolder, sd.signatureName, ex, signature)
					checkErrAndExit(err)
					generated = append(generated, sd.signatureName)
				}
			}
			if sd.nodefault == 0 {
				setSignature(sd.signatureName,
					sd.style,
					winExpandEnv(cfg.Section("Main").Key("EMailAccount").MustString("")),
					cfg.Section("Main").Key("SetForAllEMailAccounts").MustInt(0))
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
			copyFile(image, filepath.Join(destFolder, strings.ReplaceAll(filepath.Base(image), srcName, dstName)))
		}
	}

	for _, extension := range extensions {
		images, _ := filepath.Glob(filepath.Join(templateFolder, userName+"."+extension))

		for _, image := range images {
			copyFile(image, filepath.Join(destFolder, "portrait"+filepath.Ext(image)))
		}
	}

}
