package test_data_group

import "strings"

// 1000 valid charaters
func GroupName_Max_Legal() string {
	return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMN"
}

// 1001 valid charaters
func GroupName_Max_Illegal() string {
	return GroupName_Max_Legal() + "A"
}

// Has all the legal runes for th GroupName type.
func GroupName_Legal() []string {
	legalRunes := strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_", "")
	legalStrings := []string{
		"group1",
		"grOup-1",
		"Group_1",
		GroupName_Max_Legal(),
	}
	return append(legalRunes, legalStrings[:]...)
}

// Has all the legal runes for th GroupName type.
func GroupName_Illegal() []string {
	illegalRunes := strings.Split("`~!@#$%^&*()=+{}[]|\\;:'\"<,>.?/", "")
	illegalSrings := []string{
		"",
		GroupName_Max_Illegal(),
	}
	return append(illegalRunes, illegalSrings[:]...)
}

// map of user mappings
func UserMap() []interface{} {
	return []interface{}{
		map[string]interface{}{"userid": "user1@pve", "groups": ""},
		map[string]interface{}{"userid": "user2@pve", "groups": "group1"},
		map[string]interface{}{"userid": "user3@pve", "groups": "group1"},
		map[string]interface{}{"userid": "user4@pve", "groups": "group1,group2"},
		map[string]interface{}{"userid": "user5@pve", "groups": "group1,group2,group3"},
		map[string]interface{}{"userid": "user6@pve", "groups": "group2,group3"},
		map[string]interface{}{"userid": "user7@pve"},
		map[string]interface{}{"groups": "group1,group2,group3"},
	}
}
