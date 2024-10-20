package test_data_tag

// illegal character
func Tag_Character_Illegal() []string {
	return append([]string{
		`Tag!`,
		`#InvalidTag`,
		`$MoneyTag`,
		`Tag@`,
		`Tag with space`,
		`Tag&`,
		`Tag*Name`,
		`Tag#Name`,
		`Tag(Name)`,
		`Tag/Name`,
		`Tag|Name`,
		`Tag[Name]`,
		`Tag{Name}`,
		`Tag=Name`,
		`Tag+Name`,
		`Tag'Name`,
		`Tag~Name`,
		`Name<Tag`,
		`tag.with.dot`,
		`tag,with,comma`,
		`tag:name`,
		`Tag?`,
		`Tag[Bracket]`,
		`Tag{Name}`,
		`!InvalidTag`,
		`-StartWithDashTag`,
	}, Tag_Illegal())
}

func Tag_Illegal() string {
	return "!@^$^&$^&"
}

func Tag_Empty() string {
	return ""
}

func Tag_Max_Illegal() string {
	return Tag_Max_Legal() + "A"
}

func Tag_Max_Legal() string {
	return "abcdefghijklmnopqrstuvqxyz0123456789_abcdefghijklmnopqrstuvqxyz0123456789_abcdefghijklmnopqrstuvqxyz0123456789_abcdefghijklm"
}

func Tag_Legal() []string {
	return append([]string{
		`tag1`,
		`tag2`,
		`tag3`,
		`my_tag`,
		`important_tag`,
		`tech`,
		`science`,
		`art`,
		`music`,
		`coding`,
		`programming`,
		`python`,
		`72d1109e_97f6_41e7_96cc_18a8b7dc19dc`,
		`dash-tag`,
		`TagName`,
	}, Tag_Max_Legal())
}
