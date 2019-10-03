# An open source reimplementation of Kristof Zerbe's OutlookSignature

## Build

```console
go build -o gosignature.exe
```

To reduce binary size use:

```console
go build -ldflags="-s -w" gosignature.exe
```

Use [upx](https://upx.github.io) to reduce binary size even more:

```console
upx gosignature.exe
```

## Status

```ini
DatabaseConnection=
```

Not supported. LDAP only for now.

```ini
UserSelect=
```

Not supported. LDAP only for now.

```ini
LDAPBaseObjectDN=
```

List of LDAP servers, ports and base DNs (ldapserver1:port1/basedn1,ldapserver2:port2/basedn2,...)

```ini
LDAPReaderAccountName=
```

Distinguished Name of an LDAP account with read rights.

```ini
LDAPReaderAccountPassword=
```

Password of the LDAP reader account.

```ini
LDAPUserFieldname=
```

LDAP user field name (defaults to "sAMAccountName").

```ini
LDAPFilter=
```

LDAP filter (defaults to ""&(objectCategory=person)(objectClass=user)")

```ini
TemplateFolder=
```

Template directory (defaults to "Vorlagen").

```ini
EMailAccount=
```

The generated signature is set as default signature for this Outlook profile (otherwise the default profile is used).

```ini
SetForAllEMailAccounts=
```

The generated signature is set as default signature for all Outlook profiles (**0**/1).

```ini
AppDataPath=
```

Destination directory (defaults to "%appdata"\Microsoft\Signarures" on Windows and the current working directory on any other OS)

```ini
NoNewMessageSignature=
```

Disable setting the default new message signature (**0**/1)

```ini
NoReplyMessageSignature=
```

Disable setting the default reply message signature (**0**/1)

```ini
FixedSignType=
FixedSignTypeForDN1=
[...]
```

Name of the template for the new message signature. (Use "FixedSignTypeForDN1 ... n" to generate diffrent signatures for each base DN.)

```ini
FixedSignTypeReply=
FixedSignTypeReplyForDN1=
[...]
```

Name of the template for the reply message signature. (Use "FixedSignTypeReplyForDN1 ... n" to generate diffrent signatures for each base DN.)

```ini
FixedSignTypeNoMobile=
```

Not supported.

```ini
FixedSignTypeReplyNoMobile=
```

Not supported.

```ini
TargetSignType=
```

Target name of the new message signature (otherwise the name of the template is used).

```ini
TargetSignTypeReply=
```

Target name of the reply message signature (otherwise the name of the template is used).

```ini
PlaceholderSymbol=
```

Symbol which marks variables in templates (defaults to "@").

```ini
LogFile=
```

Not supported.

```ini
EmptySignatureFolder=
```

Empty the destination directory before generating signatures (**0**/1). Use "-force" to suppress the confirmation message!

## Notes

* Supports GIF, PNG and JPEG.

* Supports per user images (copies *username*.(gif|png|jpg) to portrait.(gif|png|jpg)!).

* The plain text template should be UTF-8 encoded!

* Settings for "-ini" are relative to the program's base directory.

* Features a new and more powerful template parser based on [go's template package](https://golang.org/pkg/text/template/) ("-newparser", fixed delimiter: "[[ ... ]]", have a look at the examples!).

* Relying on a GPO for setting the default signature on a company wide level should be preferred over the software's approach to manipulate the user's registry.
