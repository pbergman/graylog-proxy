package x509

import (
	"runtime/debug"
	"testing"
)

func TestDN(t *testing.T) {

	dn := &DN{
		CN:           "TEST_CN",
		SERIALNUMBER: "TEST_SERIALNUMBER",
		UID:          "TEST_UID",
		SN:           []string{"TEST", "SN"},
		C:            []string{"TEST", "C"},
		L:            []string{"TEST", "L"},
		ST:           []string{"TEST", "ST"},
		STREET:       []string{"TEST", "STREET"},
		POSTALCODE:   []string{"TEST", "POSTALCODE"},
		O:            []string{"TEST", "O"},
		OU:           []string{"TEST", "OU"},
		TITLE:        []string{"TEST", "TITLE"},
		G:            []string{"TEST", "G"},
		DC:           []string{"TEST", "DC"},
	}

	name := dn.ToPkixName()
	assertString(dn.SERIALNUMBER, name.SerialNumber, t)
	assertString(dn.CN, name.CommonName, t)
	assertStringSlice(dn.O, name.Organization, t)
	assertStringSlice(dn.C, name.Country, t)
	assertStringSlice(dn.L, name.Locality, t)
	assertStringSlice(dn.OU, name.OrganizationalUnit, t)
	assertStringSlice(dn.POSTALCODE, name.PostalCode, t)
	assertStringSlice(dn.ST, name.Province, t)
	assertStringSlice(dn.STREET, name.StreetAddress, t)

	new := &DN{}
	new.ReadPkixName(name)
	assertString(dn.SERIALNUMBER, new.SERIALNUMBER, t)
	assertString(dn.UID, new.UID, t)
	assertString(dn.CN, new.CN, t)
	assertStringSlice(dn.POSTALCODE, new.POSTALCODE, t)
	assertStringSlice(dn.STREET, new.STREET, t)
	assertStringSlice(dn.ST, new.ST, t)
	assertStringSlice(dn.C, new.C, t)
	assertStringSlice(dn.TITLE, new.TITLE, t)
	assertStringSlice(dn.DC, new.DC, t)
	assertStringSlice(dn.G, new.G, t)
	assertStringSlice(dn.L, new.L, t)
	assertStringSlice(dn.O, new.O, t)
	assertStringSlice(dn.OU, new.OU, t)
	assertStringSlice(dn.SN, new.SN, t)
}

func TestDN_ReadString(t *testing.T) {

	dn := &DN{}
	dn.ReadString("CN=Test Company,SERIALNUMBER=AABB\\CC00112233,UID=1000,SN=Surname,C=NL,L=Zuid Holland,ST=Some Street 15\\, a,POSTALCODE=1111AA,O=Company,O=Test,OU=development,TITLE=test\\=>title,G=Some Given Name,G=Some Extra Given Name,DC=example,DC=com")
	assertString("Test Company", dn.CN, t)
	assertString("AABB\\CC00112233", dn.SERIALNUMBER, t)
	assertString("1000", dn.UID, t)
	assertStringSlice([]string{"Surname"}, dn.SN, t)
	assertStringSlice([]string{"NL"}, dn.C, t)
	assertStringSlice([]string{"Zuid Holland"}, dn.L, t)
	assertStringSlice([]string{"Some Street 15, a"}, dn.ST, t)
	assertStringSlice([]string{"1111AA"}, dn.POSTALCODE, t)
	assertStringSlice([]string{"Company", "Test"}, dn.O, t)
	assertStringSlice([]string{"development"}, dn.OU, t)
	assertStringSlice([]string{"test=>title"}, dn.TITLE, t)
	assertStringSlice([]string{"Some Given Name", "Some Extra Given Name"}, dn.G, t)
	assertStringSlice([]string{"example", "com"}, dn.DC, t)
}

func assertString(a, b string, t *testing.T) {
	if a != b {
		t.Log(string(debug.Stack()))
		t.Fatalf("Expected '%s' got '%s'", a, b)
	}
}

func assertStringSlice(a, b []string, t *testing.T) {
	if len(a) != len(b) {
		t.Log(string(debug.Stack()))
		t.Fatalf("length mismatch, expected %#v got %#v", a, b)
	}

	for i, c := 0, len(a); i < c; i++ {
		assertString(a[i], b[i], t)
	}
}
