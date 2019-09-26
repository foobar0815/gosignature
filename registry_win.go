// +build windows

package main

import (
	"log"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func setSignature(signature, replysignature, profile string, setforall, nonew, noreply int) error {
	if nonew != 1 || noreply != 1 {
		outlookVersion, _ := getOutlookVersion()
		log.Println(outlookVersion)
		if outlookVersion > 14 {
			if profile != "" {
				if nonew == 0 {
					setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+profile+`\9375CFF0413111d3B88A00104B2A6676\00000002`, "New Signature", signature)
				}
				if noreply == 0 {
					setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+profile+`\9375CFF0413111d3B88A00104B2A6676\00000002`, "Reply-Forward Signature", replysignature)
				}
			} else if setforall == 1 {
				outlookProfiles, _ := getOutlookProfiles(strconv.Itoa(outlookVersion) + ".0")
				log.Println(outlookProfiles)
				for _, op := range outlookProfiles {
					if nonew == 0 {
						setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+op+`\9375CFF0413111d3B88A00104B2A6676\00000002`, "New Signature", signature)
					}
					if noreply == 0 {
						setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+op+`\9375CFF0413111d3B88A00104B2A6676\00000002`, "Reply-Forward Signature", replysignature)
					}
				}
			} else {
				defaultProfile, _ := getOutlookDefaultProfile(strconv.Itoa(outlookVersion) + ".0")
				log.Println(defaultProfile)
				if nonew == 0 {
					setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+defaultProfile+`\9375CFF0413111d3B88A00104B2A6676\00000002`, "New Signature", signature)
				}
				if noreply == 0 {
					setHkcuString(`Software\Microsoft\Office\`+strconv.Itoa(outlookVersion)+`.0\Outlook\Profiles\`+defaultProfile+`\9375CFF0413111d3B88A00104B2A6676\00000002`, "Reply-Forward Signature", replysignature)
				}
			}
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