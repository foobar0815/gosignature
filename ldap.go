package main

import (
	"time"

	"gopkg.in/ldap.v3"
)

func ldapConnect(ldapServer, ldapBind, ldapPassword string) (*ldap.Conn, error) {
	ldap.DefaultTimeout = 1 * time.Second

	conn, err := ldap.Dial("tcp", ldapServer)
	if err != nil {
		return nil, err
	}

	err = conn.Bind(ldapBind, ldapPassword)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func ldapSearch(conn *ldap.Conn, ldapBaseDN, ldapFilter, userName string, fieldMapping map[string]string) (*ldap.SearchResult, error) {
	attributes := []string{}
	for _, v := range fieldMapping {
		if v != "" {
			attributes = append(attributes, v[1:])
		}
	}

	searchRequest := ldap.NewSearchRequest(
		ldapBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"("+ldapFilter+"(sAMAccountName="+userName+"))",
		attributes,
		nil,
	)

	search, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	return search, nil

}

func ldapSearchToHash(searchResult *ldap.SearchResult) map[string]string {
	entryAsHash := make(map[string]string)
	for _, entry := range searchResult.Entries {
		for _, attribute := range entry.Attributes {
			entryAsHash[attribute.Name] = entry.GetAttributeValue(attribute.Name)
		}
	}

	return entryAsHash

}

func ldapFakeEntry() map[string]string {
	fakeentry := map[string]string{
		"givenName":                "Holly",
		"sn":                       "Kearney",
		"initials":                 "HK",
		"title":                    "Ressortleiterin",
		"telephoneNumber":          "+49 89 3176-0",
		"facsimileTelephoneNumber": "+49 89 3176-1000",
		"mobile":                   "+49 171 1234567",
		"mail":                     "holly@contoso.com",
		"postalCode":               "80807",
		"l":                        "München",
		"streetAddress":            "Walter-Gropius-Straße 5",
		"department":               "Vertrieb und Marketing",
		"company":                  "Contoso GmbH",
		"wWWHomePage":              "www.contoso.com",
		"co":                       "Deutschland",
	}

	return fakeentry
}
