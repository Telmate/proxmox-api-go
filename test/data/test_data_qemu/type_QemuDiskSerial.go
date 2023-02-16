package test_data_qemu

import "strings"

// 60 valid charaters
func QemuDiskSerial_Max_Legal() string {
	return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12345678"
}

// 61 valid charaters
func QemuDiskSerial_Max_Illegal() string {
	return QemuDiskSerial_Max_Legal() + "A"
}

// Has all the legal runes for the QemuDiskSerial type.
func QemuDiskSerial_Legal() []string {
	legalRunes := strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_", "")
	legalStrings := []string{
		"85__2-_2-p-d-___3GEEJ_6__--ccli_-_8--e-RJ3-_A_f_S_Z8-7Gga__5",
		"___7_7-G-_rsMa_6---___a-TmE-H-AD_oV_K-0_9W_-y_4_k-_FU__fev-q",
		"mU_a-8-_-3U--_D--o01_-T4E__4fsV-nX4kk_Nb-S_wYR7TktI-_vn1mcfR",
		"5_GX4DevvIZ-_2-_u-_4_dKK19_P-K-kLpj-Hzw-b_12L20UU3--__Y_5_O-",
		"h-m_B_J-mH_o3-r-JmE4-WqZ_tly--3aT-w_wznK_Q-0Hkk7W6b-Z____1IQ",
		"-D-aS-dJSG_-s9_6_BWnt9z_6m_67oW__d4V8m-wb_6_8-_A--__-e--1X1-",
		"0gLe0_P-r53-BfDk--1__23_-Zo0_V---f-__7_b52u6f72",
		"dg-5d_Y1_v-7g-__I7__79_-j8-_-_",
		"-7_G_-I-__a-w_-5GK-_9BMr-I_-_-f_7_-F-----6-FE-q7__0NJY-vL-e-",
		"Xlgd__41_0E-S---1--_--_X1p8-_5YHt_1hO__2Op_73-5r-4",
		"_0-Nn6__-u_07_A-1dqt-EG-_4-w90--A-ur_-3_",
		"8_-xO-lN1_O_Q-4__-_7d-k-__s-p-_uQ2S_Ft_OR_-Ct_--Pb__U_U_g2t-",
		"-q_-6h4--bbIl_A-xc_zl-v6Y-b-6__EG5_G__6_-q__pa5lAvq_F-Fhl-d3",
		"53X-j50-ix--Iv_i",
		"0_-7O6--51_4____-_-Zm4E__s-4_c_xN_Ik3_g-_t__-__C_---e-_--K_m",
		"--_8__BvpmE6t-r_Ho_x-VZ_0___g__ui1v28ne--_-_k74_E3x_T_s--__B",
		"__--9--c__7__9r-s-__yDTi-JSk-M_fH_-hGO9",
		"_J-q_f_o_--l-MSe--9I_L_-lAs_-G-0--l-9_-6",
		"--1tcB8J210JwYy22--c-_oXhHQ-Zyy--A1-dZ9394ieAaZrvC_U--KS2-r7",
		"442_h_M-2-4G3K-_bN",
		"b_NDX_3e6-k9_-HWZL_A5T_L-_-je66",
		"__-_DfxD-9_l2_-n__Tn_-n_6aE_Bj_8chVS5p7Z_2812---h_4-_hsh-wAb",
		"T08_-bRb-_zRKF1MPN6j7vyp0Pt_Q_x44__Y6_7_-XX-_-VqsRi_-0-s2-",
		"Z___e_UT--G_9-9_E--__--ZT---1Leg__7-92_KErs-S-_t-K21_-OK_6Nx",
		"-Plro-431-_",
		"b__-_2_-u-K8--3-1-1",
		"3_T40b6-Q__-Lz_3_4qN_",
		"-p-_-T__MK-4D46-r7F_",
		"Cn_f1-3J5T4x-g_-yoX3__0K_lV_W-glt-_eC_c_A_3l-m-_d_5_-5-iC___",
		"-v7Y--G8T6-5k64-2-4kz-zqn70C__",
		"7-7S70i_e-_a_8E---1X-_cW_--4z_8_4rm--6KCp_5_os8-_-4pvE_UC-_X",
		"-mjp2--E8z-K-__2t",
		"2b6-_-4jx6nM-Mm_-Sf_Wz-cv_---keD_6-_x_O0_D__d_Q-v04-h__x__6g",
		"_-t0_Et_1_P0C5xSp_-3G2u68q4--YU7zQ_-_4-Y3OwT-8k-2F__Zf7_K78k",
		"_R_-9-_5-K2R5zF_UJWfH_--_YQe8P4_nQ--JQ-Xu--lH-e_V0-c1_--e",
		"J_y_-1X-_N84-Y-Ik6h-Y5Hz-_S__0-_-Y_V_-305-9i38TP_T--4y7Gb0g-",
		"0Yl7_-Sn-_H__D_--_-_y-h-j1_-7mDVS9__R8Ty_-mwOJ-h__VmAy__q-_N",
		"Yk-3Pt9VJqW-uu45w_f__a-TH_M_vkWq-O6__tk-P4-2-_CP-_7_owV_GyLZ",
		"-3-M2362g3_9-___X404__0-E01TFHW3HV-1n_E_Ev_-_-E_-S----_F_k--",
		"-0-5l1u_D_-Hyk_-_-s-3663_JE2-k",
		"_-ML",
		"75BHJ-tj-2Q-m-s4-f-y5-9L_-_dPC6A-_3y-3-Rp_-0V-3_I-6--_4THR2-",
		"V__HM_guP-__ltfk___Et9V_v7U-83-TK15Xun",
		"ax2_-7u_-GqG7d1BAe_4-dW3S0-1Vbj--_r_-63Hvn36P0---6N--e-_89_b",
		"__-h_L-_3_Q__3czJ1_J_",
		"-_-_-_9xGr-8207h9-4WerP5_-v0G_G9_-K9T8-gvQg-d__g____xf-F_---",
		"M_6K2LVn--73f__-__A5V_t-l-_--5s92_u_f_n-R2Iy580M7J2vPe76GIt-",
		"-f5oc-MG_GhR-8oxJ_---G____5--2-5___b-cTmfyO18-6Vip__-c_i5uV_",
		"7-_17B__b113_9-cZ8S-",
		"VvF1-MX6--Pj-M_-F_tU_6T-G___3U3-F4-5L__x-9uzgz_-_y4_-_-J--kE",
		"",
		QemuDiskSerial_Max_Legal(),
	}
	return append(legalRunes, legalStrings[:]...)
}

// Has all the legal runes for the QemuDiskSerial type.
func QemuDiskSerial_Illegal() []string {
	illegalRunes := strings.Split("`~!@#$%^&*()=+{}[]|\\;:'\"<,>.?/", "")
	illegalSrings := []string{QemuDiskSerial_Max_Illegal()}
	return append(illegalRunes, illegalSrings[:]...)
}
