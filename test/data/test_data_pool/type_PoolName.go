package test_data_pool

import "strings"

func PoolName_Error_Characters() []string {
	// '/' is legal now (nested pools like Hardware/Test), so it must not be listed here.
	return strings.Split("`~!@#$%^&*()=+{}[]|\\;:'\"<,>.?", "")
}

// 1024 valid charaters
func PoolName_Max_Legal() string {
	return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_"
}

// 1025 valid charaters
func PoolName_Max_Illegal() string {
	return PoolName_Max_Legal() + "A"
}

func PoolName_Legal() string {
	return "pool-1"
}

// Has all the legal runes for th PoolName type.
func PoolName_Legals() []string {
	legalRunes := strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_", "")
	legalStrings := []string{
		"pool1",
		"pOol-1",
		"PooL_1",
		PoolName_Legal(),
		PoolName_Max_Legal(),
	}
	// Nested pool ids (slash-separated), e.g. "Hardware/Test".
	legalStrings = append(legalStrings,
		"Test/Hardware",
		"Hardware/Test/VMs",
	)
	return append(legalRunes, legalStrings[:]...)
}
