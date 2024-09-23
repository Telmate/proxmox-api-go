package test_data_mtu

func MTU_Min_Legal() uint16 {
	return 576
}

func MTU_Min_Illegal() uint16 {
	return MTU_Min_Legal() - 1
}

func MTU_Max_Legal() uint16 {
	return 65520
}

func MTU_Max_Illegal() uint16 {
	return MTU_Max_Legal() + 1
}
