// +build windows

package main

import (
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func setSignature(signature, style, profile string, setforall int) error {
	outlookVersion, _ := getOutlookVersion()

	if outlookVersion > 12 {

		name := ""
		if style == "new" {
			name = "New Signature"
		} else if style == "reply" {
			name = "Reply-Forward Signature"
		}

		profiles := []string{}
		if profile != "" {
			profiles = append(profiles, profile)
		} else if setforall == 1 {
			outlookProfiles, _ := getOutlookProfiles(outlookVersion)
			for _, op := range outlookProfiles {
				profiles = append(profiles, op)
			}
		} else {
			defaultProfile, _ := getOutlookDefaultProfile(outlookVersion)
			profiles = append(profiles, defaultProfile)
		}

		for _, p := range profiles {
			subkeys, err := getAccountSubkeys(outlookVersion, p)
			if err != nil {
				return err
			}
			for _, sk := range subkeys {
				setSignatureValue(outlookVersion, sk, name, signature)
			}
		}
	}

	return nil
}

func getAccountSubkeys(outlookVersion int, profile string) ([]string, error) {
	key := ""
	if outlookVersion > 14 {
		key = `Software\Microsoft\Office\` + strconv.Itoa(outlookVersion) + `.0\Outlook\Profiles\` + profile + `\9375CFF0413111d3B88A00104B2A6676`
	} else if outlookVersion > 12 {
		key = `Software\Microsoft\Windows NT\CurrentVersion\Windows Messaging Subsystem\Profiles\` + profile + `\9375CFF0413111d3B88A00104B2A6676`
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.READ)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	keystats, err := k.Stat()
	if err != nil {
		return nil, err
	}

	subkeycount := int(keystats.SubKeyCount)

	subkeys, err := k.ReadSubKeyNames(subkeycount)
	if err != nil {
		return nil, err
	}

	accountSubkeys := []string{}
	for _, subkey := range subkeys {
		sk, err := registry.OpenKey(registry.CURRENT_USER, key+`\`+subkey, registry.READ)
		if err != nil {
			return nil, err
		}
		defer sk.Close()

		mapiprovider, _, err := sk.GetIntegerValue("MAPI Provider")

		if mapiprovider != 2 && mapiprovider != 4 {
			accountSubkeys = append(accountSubkeys, key+`\`+subkey)
		}
	}

	return accountSubkeys, nil

}

func setSignatureValue(outlookVersion int, key, name, value string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if outlookVersion > 14 {
		err = k.SetStringValue(name, value)
		if err != nil {
			return err
		}
	} else if outlookVersion > 12 {
		msValue := ""
		for _, char := range msValue {
			msValue += string(char) + "\x00"
		}
		msValue += "\x00\x00"

		err = k.SetBinaryValue(name, []byte(msValue))
		if err != nil {
			return err
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

func getOutlookDefaultProfile(outlookVersion int) (string, error) {
	key := ""
	if outlookVersion > 14 {
		key = `Software\Microsoft\Office\` + strconv.Itoa(outlookVersion) + `.0\Outlook`
	} else if outlookVersion > 12 {
		key = `Software\Microsoft\Windows NT\CurrentVersion\Windows Messaging Subsystem\Profiles`
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.QUERY_VALUE)
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

func getOutlookProfiles(outlookVersion int) ([]string, error) {
	key := ""
	if outlookVersion > 14 {
		key = `Software\Microsoft\Office\` + strconv.Itoa(outlookVersion) + `.0\Outlook\Profiles`
	} else if outlookVersion > 12 {
		key = `Software\Microsoft\Windows NT\CurrentVersion\Windows Messaging Subsystem\Profiles`
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.READ)
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
