package command

import "github.com/pbergman/app"

func NewDnCommand() app.CommandInterface {
	return &app.Command{
		Name:  "dn",
		Short: "Information about the DN argument",
		Long: `
The distinguished name (DN) uniquely identifies an entity in an X.509 certificate that
are used when creating the ca (certificate authority), server and client certificate.

The syntax used for the required commands is a key value with a equals sign as delimiter
(and with escaped comma for the value) and comma delimiter for the attributes.

Example: 'CN=logger.example.com,C=Netherlands,L=NL,O=Foo\, Bar'

The following attributes are supported:

    CN               // common name
    SN               // surname
    SERIALNUMBER     // certificate serial number
    C                // country
    L                // locality
    ST               // state or province
    STREET           // street
    POSTALCODE       // postal code
    O                // organization name
    OU               // organizational unit
    TITLE            // title
    G                // given name
    UID              // userid
    DC               // domain component

`,
	}
}
