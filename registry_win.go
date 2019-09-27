// +build windows

package main

import (
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func setSignature(signature, style, profile string, setforall int) error {
	name := ""
	if style == "new" {
		name = "New Signature"
	} else if style == "reply" {
		name = "Reply-Forward Signature"
	}
	outlookVersion, _ := getOutlookVersion()
	if outlookVersion > 14 {
		if profile != "" {
			setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+profile+`\9375CFF0413111d3B88A00104B2A6676\00000002`, name, signature)
		} else if setforall == 1 {
			outlookProfiles, _ := getOutlookProfiles(strconv.Itoa(outlookVersion) + ".0")
			for _, op := range outlookProfiles {
				setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+op+`\9375CFF0413111d3B88A00104B2A6676\00000002`, name, signature)
			}
		} else {
			defaultProfile, _ := getOutlookDefaultProfile(strconv.Itoa(outlookVersion) + ".0")
			setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+defaultProfile+`\9375CFF0413111d3B88A00104B2A6676\00000002`, name, signature)
		}
	}

	return nil
}

func setHkcuString(key, name, data string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	err = k.SetStringValue(name, data)
	if err != nil {
		return err
	}

	return nil

}

func getOutlookVersion() (int, error) {
	k, err := registry.OpenKey(registry.CLASSES_ROOT, `Outlook.Application\CurVer`, registry.QUERY_VALUE)
	if err != nil {
		return 0, err
	}
	defer k.Close()

	s, _, err := k.GetStringValue("")
	if err != nil {
		return 0, err
	}

	ssplit := strings.Split(s, ".")
	version, err := strconv.Atoi(ssplit[len(ssplit)-1])
	if err != nil {
		return 0, err
	}

	return version, nil
}

func getOutlookDefaultProfile(outlookVersion string) (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Office\`+outlookVersion+`\Outlook`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	profile, _, err := k.GetStringValue("DefaultProfile")
	if err != nil {
		return "", err
	}

	return profile, nil
}

func getOutlookProfiles(outlookVersion string) ([]string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Office\`+outlookVersion+`\Outlook\Profiles`, registry.READ)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	keystats, err := k.Stat()
	if err != nil {
		return nil, err
	}

	subkeycount := int(keystats.SubKeyCount)

	profiles, err := k.ReadSubKeyNames(subkeycount)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}
