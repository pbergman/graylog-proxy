package x509

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"strings"
)

var (
	// A mapping between the DN property name and the ASN1
	// ObjectIdentifier that it represent.
	iod = map[string][]int{
		"CN":           []int{2, 5, 4, 3},                       // common name
		"SN":           []int{2, 5, 4, 4},                       // surname
		"SERIALNUMBER": []int{2, 5, 4, 5},                       // Certificate serial number
		"C":            []int{2, 5, 4, 6},                       // country
		"L":            []int{2, 5, 4, 7},                       // locality
		"ST":           []int{2, 5, 4, 8},                       // state or province
		"STREET":       []int{2, 5, 4, 9},                       // street
		"POSTALCODE":   []int{2, 5, 4, 17},                      // postal code
		"O":            []int{2, 5, 4, 10},                      // organization name
		"OU":           []int{2, 5, 4, 11},                      // organizational unit
		"TITLE":        []int{2, 5, 4, 12},                      // title
		"G":            []int{2, 5, 4, 42},                      // given name
		"UID":          []int{0, 9, 2342, 19200300, 100, 1, 1},  // userid
		"DC":           []int{0, 9, 2342, 19200300, 100, 1, 25}, // domain component
	}
)

// DN represents an X.509 distinguished name. It can read or copy values to
// a pkix.Name object that can be used for creating a certificate. The main
// deference between a pkix.Name and DN is that the DN can also read from a
// string and write to string (ReadString and String methods)
type DN struct {
	CN           string
	SERIALNUMBER string
	UID          string
	SN           []string
	C            []string
	L            []string
	STREET       []string
	POSTALCODE   []string
	ST           []string
	O            []string
	OU           []string
	TITLE        []string
	G            []string
	DC           []string
}

func (d *DN) setValue(name, value string) {
	switch name {
	case "CN":
		d.CN = value
	case "UID":
		d.UID = value
	case "SERIALNUMBER":
		d.SERIALNUMBER = value
	case "SN":
		d.SN = append(d.SN, value)
	case "C":
		d.C = append(d.C, value)
	case "L":
		d.L = append(d.L, value)
	case "ST":
		d.ST = append(d.ST, value)
	case "STREET":
		d.STREET = append(d.STREET, value)
	case "O":
		d.O = append(d.O, value)
	case "OU":
		d.OU = append(d.OU, value)
	case "TITLE":
		d.TITLE = append(d.TITLE, value)
	case "POSTALCODE":
		d.POSTALCODE = append(d.POSTALCODE, value)
	case "G":
		d.G = append(d.G, value)
	case "DC":
		d.DC = append(d.DC, value)
	}
}

func (d DN) getIodByName(o asn1.ObjectIdentifier) string {
	for name, id := range iod {
		if o.Equal(id) {
			return name
		}
	}
	return ""
}

func (d DN) addToRDNSequence(sequence *pkix.RDNSequence, values []string, id asn1.ObjectIdentifier) {
	if s := len(values); s > 0 {
		for c := 0; c < s; c++ {
			*sequence = append(*sequence, pkix.RelativeDistinguishedNameSET{{id, values[c]}})
		}
	}
}

func (d DN) addToExtraName(name *pkix.Name, values []string, id asn1.ObjectIdentifier) {
	if s := len(values); s > 0 {
		for c := 0; c < s; c++ {
			name.ExtraNames = append(name.ExtraNames, pkix.AttributeTypeAndValue{id, values[c]})
		}
	}
}

// ReadPkixName will copy the values from a pkix.Name to this instance
func (d *DN) ReadPkixName(n pkix.Name) {
	for _, sec := range n.ToRDNSequence() {
		for _, set := range sec {
			d.setValue(d.getIodByName(set.Type), set.Value.(string))
		}
	}
}

func (d DN) ToPkixName() (n pkix.Name) {
	name := make(pkix.RDNSequence, 0)
	if len(d.CN) > 0 {
		d.addToRDNSequence(&name, []string{d.CN}, iod["CN"])
	}
	if len(d.SERIALNUMBER) > 0 {
		d.addToRDNSequence(&name, []string{d.SERIALNUMBER}, iod["SERIALNUMBER"])
	}
	d.addToRDNSequence(&name, d.C, iod["C"])
	d.addToRDNSequence(&name, d.POSTALCODE, iod["POSTALCODE"])
	d.addToRDNSequence(&name, d.L, iod["L"])
	d.addToRDNSequence(&name, d.ST, iod["ST"])
	d.addToRDNSequence(&name, d.STREET, iod["STREET"])
	d.addToRDNSequence(&name, d.O, iod["O"])
	d.addToRDNSequence(&name, d.OU, iod["OU"])
	d.addToRDNSequence(&name, d.G, iod["G"])
	// extra atributes and not directly supported by
	if len(d.UID) > 0 {
		d.addToRDNSequence(&name, []string{d.UID}, iod["UID"])
	}
	d.addToExtraName(&n, d.SN, iod["SN"])
	d.addToExtraName(&n, d.G, iod["G"])
	d.addToExtraName(&n, d.TITLE, iod["TITLE"])
	d.addToExtraName(&n, d.DC, iod["DC"])
	n.FillFromRDNSequence(&name)
	return n
}

func (d *DN) ReadString(str string) {
	name, buf, escape := "", "", false
	for _, value := range str {
		if !escape {
			switch value {
			case '\\':
				escape = true
			case '=':
				name, buf = strings.ToUpper(buf), ""
				continue
			case ',':
				d.setValue(name, buf)
				buf = ""
				continue
			default:
				buf += string(value)
			}
		} else {
			// check for a slash in the value and not escape
			if value != ',' && value != '=' {
				buf += "\\"
			}
			buf += string(value)
			escape = false
		}

	}
	if len(buf) > 0 {
		d.setValue(name, buf)
	}
}
