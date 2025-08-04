// Package test_data_guest provides test data.
package test_data_guest

func GuestNameLegal() []string {
	return append(guestNameLegal(), GuestNameMaxLegal())
}

func guestNameLegal() []string {
	return []string{
		"0Z1-2a",
		"1xYz9",
		"1z-abc-de",
		"3-abCDeF",
		"4-5678",
		"6gh-ijkl",
		"7.yz-12",
		"7lmn-p.op.qr",
		"7xyz-abc",
		"8abcd-efgh.ijkl",
		"9-0.abc",
		"9-lmnop.qrst",
		"B-cdE",
		"N0p-qrS",
		"Q-w-e-r",
		"T-u-vw",
		"Xy-z12",
		"Yz12-3456",
		"Z-abc-def",
		"Zabc-defg-hijk",
		"a-1-b2",
		"a-bc.123",
		"a.b.c-c.e-a.f",
		"fghij-klmno",
		"m-noPq-rS",
		"mno-pqr",
		"p-qr.st.uv",
		"uVw-123",
		"wXy.7z",
		"zYx1",
	}
}

func GuestNameCharacterIllegal() []string {
	return []string{
		"a bc",         // contains space
		"B_cd",         // contains underscore
		"7xyz@abc",     // contains @
		"zYx1#",        // contains #
		"8abcd$efgh",   // contains $
		"1xYz9*",       // contains *
		"p-qr=st",      // contains =
		"m+noPq-rS",    // contains +
		"0Z1~2a",       // contains ~
		"wXy?7z",       // contains ?
		"uVw^123",      // contains ^
		"9{lmnop}qrst", // contains {}
		"Zabc|defg",    // contains |
		"N0p\\qrS",     // contains backslash
		"3-abCD!eF",    // contains !
		"Yz12%3456",    // contains %
		"7lmn&opqr",    // contains &
		"a,bc,123",     // contains commas
		"1z;abc;de",    // contains semicolons
		"mno: pqr",     // contains colon & space
		"Z<abc>def",    // contains < and >
		"T/u\\v",       // contains slash & backslash
	}
}

func GuestNameEmpty() string {
	return ""
}

func GuestNameMaxIllegal() string {
	return GuestNameMaxLegal() + "x"
}

func GuestNameMaxLegal() string {
	return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-ab"
}

func GuestNameEndIllegal() []string {
	names := guestNameLegal()
	namesDot := make([]string, len(names))
	namesHyphen := make([]string, len(names))
	for i := range names {
		namesDot[i] = names[i] + "."
		namesHyphen[i] = names[i] + "-"
	}
	return append(namesDot, namesHyphen...)
}

func GuestNameStartIllegal() []string {
	names := guestNameLegal()
	namesDot := make([]string, len(names))
	namesHyphen := make([]string, len(names))
	for i := range names {
		namesDot[i] = "." + names[i]
		namesHyphen[i] = "-" + names[i]
	}
	return append(namesDot, namesHyphen...)
}
