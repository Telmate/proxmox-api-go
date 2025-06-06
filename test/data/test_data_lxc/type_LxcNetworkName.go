package test_data_lxc

// illegal character
func LxcNetworkName_Character_Illegal() []string {
	return []string{
		`eth0!`,
		`en0@`,
		`lo0#`,
		`eth/0`,
		`eth 0`,
		`eth0\n`,
		`💻net0`,
		`名字123`,
		`int$name`,
		`net%if`,
		`a\tb`,
		`iface*`,
		`name:net`,
		`net\niface`,
		`🔥interface`,
		`if\nace`,
		`net&interface`,
		`net@work`,
		`iface\0`,
		`net iface`,
		`interface!`,
		`enp0s3\nen0`,
		`eth0\r\n`,
		`if#ace`,
		`网卡123`,
		`lo0🙂`}
}

func LxcNetworkName_Special_Illegal() []string {
	return []string{
		".."}
}

// 16 valid characters
func LxcNetworkName_Max_Legal() string {
	return "abcdefghijklmnop"
}

// 17 valid characters
func LxcNetworkName_Max_Illegal() string {
	return LxcNetworkName_Max_Legal() + "A"
}

// 2 valid characters
func LxcNetworkName_Min_Legal() string {
	return LxcNetworkName_Min_Illegal() + "a"
}

// 2 invalid characters
func LxcNetworkName_Min_Illegal() string {
	return "a"
}

func LxcNetworkName_Legal() []string {
	return []string{
		"eth0",
		"lo",
		"wlan0",
		"enp3s0",
		"en0",
		"eth1",
		"br-1234",
		"tun0",
		"tap1",
		"docker0",
		"wlp2s0",
		"enp0s3",
		"eth0.100",
		"eth0-1",
		"bridge0",
		"wlan1",
		"if0",
		"net_01",
		"veth1234",
		"vmnet1",
		"usb0",
		"sit0",
		"gre1",
		"ppp0",
		"eno1",
		"enx001",
		"eth_2",
		"net0-1",
		"wg0",
		"eth0-test",
		LxcNetworkName_Max_Legal(),
		LxcNetworkName_Min_Legal()}
}
