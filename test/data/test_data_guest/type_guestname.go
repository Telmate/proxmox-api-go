// Package test_data_guest provides test data.
package test_data_guest

func GuestNameLegal() []string {
	return []string{

		"B-cdE",
		"7xyz-abc",
		"zYx1",
		"Q-w-e-r",
		"8abcd-efgh-ijkl",
		"1xYz9",
		"p-qr-st-uv",
		"m-noPq-rS",
		"0Z1-2a",
		"fghij-klmno",
		"wXy-7z",
		"uVw-123",
		"9-lmnop-qrst",
		"Zabc-defg-hijk",
		"N0p-qrS",
		"3-abCDeF",
		"6gh-ijkl",
		"Yz12-3456",
		"7lmn-opqr",
		"4-5678",
		"Xy-z12",
		"a-bc-123",
		"1z-abc-de",
		"mno-pqr",
		"Z-abc-def",
		"7-yz-12",
		"T-u-vw",
		"9-0-abc",
		"a-1-b2",
		GuestNameMaxLegal()}
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
		"fghij.klmno",  // contains dot
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

func GuestNameStartIllegal() string {
	return "-" + GuestNameLegal()[0]
}
