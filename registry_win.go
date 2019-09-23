// +build windows

package main

import (
	"log"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func setSignature() error {
	outlookVersion, _ := getOutlookVersion()
	log.Println(outlookVersion)
	if outlookVersion > 14 {
		outlookProfiles, _ := getOutlookProfiles(strconv.Itoa(outlookVersion) + ".0")
		log.Println(outlookProfiles)
		defaultProfile, _ := getOutlookDefaultProfile(strconv.Itoa(outlookVersion) + ".0")
		log.Println(defaultProfile)

		k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+defaultProfile+`\9375CFF0413111d3B88A00104B2A6676\00000002`, registry.SET_VALUE)
		if err != nil {
			return err
		}
		defer k.Close()

		strings := [2]string{"New Signature", "Reply-Forward Signature"}

		for _, string := range strings {
			err = k.SetStringValue(string, "foo")
			if err != nil {
				return err
			}
		}
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
