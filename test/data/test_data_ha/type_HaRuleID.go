package test_data_ha

func HaRuleID_Legal() []string {
	return append(haRuleID_Legal(), HaRuleID_MaxLegal())
}

func haRuleID_Legal() []string {
	return []string{
		"aZ1-2a",
		"BxYz9",
		"cz-abc-de",
		"D-abCDeF",
		"e-5678",
		"Fgh-ijkl",
		"gyz-12-",
		"Hlmn-popqr",
		"Ixyz-abc",
		"jabcd-efghijkl",
		"K-0abc_",
		"l-lmnopqrst",
		"M-cdE",
		"n0p-qrS",
		"O-w-e-r",
		"p-u-vw",
		"Qy-z12",
		"rz12-3456",
		"S-abc-def",
		"tabc-defg-hijk",
		"U-1-b2",
		"v-bc123",
		"Wbc-ce-af",
		"xghij-klmno",
		"Y-noPq-rS",
		"zno-pqr",
	}
}

func HaRuleID_CharacterIllegal() []string {
	return []string{
		"a bc",         // contains space
		"B.cd",         // contains .
		"cxyz@abc",     // contains @
		"DYx1#",        // contains #
		"eabcd$efgh",   // contains $
		"FxYz9*",       // contains *
		"g-qr=st",      // contains =
		"H+noPq-rS",    // contains +
		"IZ1~2a",       // contains ~
		"jXy?7z",       // contains ?
		"KVw^123",      // contains ^
		"l{lmnop}qrst", // contains {}
		"Mabc|defg",    // contains |
		"n0p\\qrS",     // contains backslash
		"o-abCD!eF",    // contains !
		"Pz12%3456",    // contains %
		"qlmn&opqr",    // contains &
		"R,bc,123",     // contains commas
		"sz;abc;de",    // contains semicolons
		"Tno: pqr",     // contains colon & space
		"u<abc>def",    // contains < and >
		"V/u\\v",       // contains slash & backslash
	}
}

func HaRuleID_MinLength() []string {
	return []string{"", "a"}
}

func HaRuleID_MaxIllegal() string {
	return HaRuleID_MaxLegal() + "x"
}

// 128
func HaRuleID_MaxLegal() string {
	return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-ab"
}

func HaRuleID_StartIllegal() []string {
	names := haRuleID_Legal()
	prefix := []string{"-", "_", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	results := make([]string, len(prefix))
	for i := range prefix {
		results[i] = prefix[i] + names[i]
	}
	return results
}
