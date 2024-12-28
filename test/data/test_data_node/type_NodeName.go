package test_data_node

import "strings"

func NodeName_Error_Characters() []string {
	characters := strings.Split("`_~!@#$%^&*()=+{}[]|\\;:'\"<,>.?/", "")
	for i := range characters {
		characters[i] = characters[i] + "node"
	}
	return characters
}

func NodeName_Max_Legal() string {
	return "abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
}

func NodeName_Max_Illegal() string {
	return NodeName_Max_Legal() + "A"
}

func NodeName_Numeric_Illegal() []string {
	return []string{
		"1",
		"12",
		"123",
		"12-34",
		"1234567-8901234",
	}
}

func nodeName_Legals() []string {
	return []string{
		"node1",
		"nOde-1",
		"nOde1"}
}

func NodeName_Legals() []string {
	return append(nodeName_Legals(), NodeName_Max_Legal())
}

func NodeName_StartHyphens() []string {
	legals := nodeName_Legals()
	hyphen := make([]string, len(legals))
	for i := range legals {
		hyphen[i] = "-" + legals[i]
	}
	return hyphen
}

func NodeName_EndHyphens() []string {
	legals := nodeName_Legals()
	hyphen := make([]string, len(legals))
	for i := range legals {
		hyphen[i] = legals[i] + "-"
	}
	return hyphen
}
