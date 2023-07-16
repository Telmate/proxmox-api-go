package test_data_snapshot

// illegal character
func SnapshotName_Character_Illegal() []string {
	return []string{
		"aBc123!4567890_-",
		"Qwer@ty-1234_ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"x1y2#z3_4-5-6-7-8-9",
		"HelloWo$rld_2023",
		"Ab1_%cd2_ef3-gh4-ij5",
		"a-_-^_-_-_-_-_-_-_-_-_-",
		"snaps&hotName_2433242",
		"A1_B2-*C3_D4-E5_F6",
		"Xyz-123(_456_789-0",
		"Test_Cas)e-123_456_789_0",
		"a_1+",
		"B-c_=2-D",
		"E3_f4-G5_:H6-I7",
		"JKL_MNO_PQ;R-STU_VWX_YZ0",
		"aBgnhfjkfgd'ihfghudsfgio",
		`Cdsdjfidshfu"isdghfsgffghdsufsdhfgdsfuah`,
		"Ef-`gh",
		"Ij-k~l-mn",
		"Op-qr-st-u-vw-xy-z0-12-34-56-[78-90",
		"Abcd_1234-EFGH_]5678-IJKL_9012",
		"M-n-Op-qR-sT-uV{-wX-yZ",
		"a_b-c_d_e-f_g_h_}i_j_k_l_m_n-o-p-q-r-s-t",
		"Aa1_Bb2-C,c3_Dd4-Ee5_Ff6-Gg7_Hh8-Ii9",
		"JjKkLl-MmNnOo.PpQqRrSsTtUuVvWwXxYyZz01",
		"A->1",
		"B-2<_C-3",
		"D-4_?E-5-F-6",
		"G-7-H/-8-I-9",
		`J-0_K-\1-L-2-M-3-N-4-O-5-P-6-Q-7-R-8-S-9`,
		"T-0_U-1-|V-2-W-3-X-4-Y-5-Z-6-7-8-9-0",
		"a2😀",
	}
}

// 40 valid characters
func SnapshotName_Max_Legal() string {
	return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMN"
}

// 41 invalid characters
func SnapshotName_Max_Illegal() string {
	return SnapshotName_Max_Legal() + "A"
}

// 3 valid characters
func SnapshotName_Min_Legal() string {
	return SnapshotName_Min_Illegal() + "c"
}

// 2 invalid characters
func SnapshotName_Min_Illegal() string {
	return "ab"
}

// legal starting character
func SnapshotName_Start_Legal() string {
	return "abc"
}

// illegal starting character
func SnapshotName_Start_Illegal() []string {
	return []string{
		"_" + SnapshotName_Start_Legal(),
		"-" + SnapshotName_Start_Legal(),
		"0" + SnapshotName_Start_Legal(),
		"5" + SnapshotName_Start_Legal(),
	}
}

func SnapshotName_Legal() []string {
	return []string{
		"aBc1234567890_-",
		"Qwerty-1234_ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"x1y2z3_4-5-6-7-8-9",
		"HelloWorld_2023",
		"Ab1_cd2_ef3-gh4-ij5",
		"a-_-_-_-_-_-_-_-_-_-_-",
		"snapshotName_2433242",
		"A1_B2-C3_D4-E5_F6",
		"Xyz-123_456_789-0",
		"Test_Case-123_456_789_0",
		"a_1",
		"B-c_2-D",
		"E3_f4-G5_H6-I7",
		"JKL_MNO_PQR-STU_VWX_YZ0",
		"aBgnhfjkfgdihfghudsfgio",
		"Cdsdjfidshfuisdghfsgffghdsufsdhfgdsfuahs",
		"Ef-gh",
		"Ij-kl-mn",
		"Op-qr-st-u-vw-xy-z0-12-34-56-78-90",
		"Abcd_1234-EFGH_5678-IJKL_9012",
		"M-n-Op-qR-sT-uV-wX-yZ",
		"a_b-c_d_e-f_g_h_i_j_k_l_m_n-o-p-q-r-s-t-",
		"Aa1_Bb2-Cc3_Dd4-Ee5_Ff6-Gg7_Hh8-Ii9",
		"JjKkLl-MmNnOoPpQqRrSsTtUuVvWwXxYyZz01",
		"A-1",
		"B-2_C-3",
		"D-4_E-5-F-6",
		"G-7-H-8-I-9",
		"J-0_K-1-L-2-M-3-N-4-O-5-P-6-Q-7-R-8-S-9",
		"T-0_U-1-V-2-W-3-X-4-Y-5-Z-6-7-8-9-0",
		"a2B",
		"c4D",
		"e6F-g8H-i0J",
		"k2L-m4N-o6P-q8R-s0T",
		"u2V-w4X-y6Z-01-23-45-67-89-0",
		"Abc_1234-Def_5678-Ghi_9012-Jkl_3456-Mno_",
		"Pqr_2345-Stu_6789-Vwx_0123-Yz0_4567",
		"a-B",
		"c-D_e-F",
		"g-H_i-J-k-L",
		"m-N-o-P_q-R-s-T-u-V-w-X-y-Z-0",
		"A_1b2-C3d4_E5f6-G7h8_I9j0-K1l2-M3n4",
		"O5p6-Q7r8-S9t0-U1v2-W3x4-Y5z6-01",
		"A2b3-C4d5-E6f7-G8h9-I0j1-K2l3-M4n5-O6",
		"P7q8-R9s0-T1u2-V3w4-X5y6-Z7-89-01-23-45-",
		"Ab_12-cD_34-eF_56-gH_78-iJ_90-kL_12-mN_3",
		"O5p6-Q7r8-S9t0-U1v2-W3x4-Y5z6-01-23-45",
		"A7b8-C9d0-E1f2-G3h4-I5j6-K7l8-M9n0-O1p2-",
		"S5t6-U7v8-W9x0-Y1z2-34-56-78-90-12-34-56",
		"Ab1C_d2E-F3G_h4I-J5k6L-m7N-o8P-q9R-s0T-u",
		SnapshotName_Max_Legal(),
		SnapshotName_Min_Legal(),
		SnapshotName_Start_Legal(),
	}
}
