package proxmox

// X_SuppressStaticCheck only exists to suppress warnings for unused functions
func X_SuppressStaticCheck_DoNotUse() {
	_ = mockClientAPI{}.new()

	lxc := &RawConfigLXCMock{}
	lxc.get(VmRef{})
	lxc.getBootMount(true)
	lxc.getDigest()
	lxc.getMounts(true)

	qemu := &RawConfigQemuMock{}
	qemu.get(&VmRef{})

	LxcRef{}.generalize()
	QemuRef{}.generalize()
	LxcRef{}.legacy()
	QemuRef{}.legacy()
}
