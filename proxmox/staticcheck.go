package proxmox

import "context"

// X_SuppressStaticCheck only exists to suppress warnings for unused functions
func X_SuppressStaticCheck_DoNotUse() {
	_ = mockClientAPI{}.new()

	ca := &clientAPI{}
	_ = ca.post(context.Background(), "", nil)
	_, _ = ca.postTask(context.Background(), "", nil)

	lxc := &RawConfigLXCMock{}
	lxc.get(VmRef{})
	lxc.getBootMount(true)
	lxc.getDigest()
	lxc.getMounts(true)

	qemu := &RawConfigQemuMock{}
	qemu.get(&VmRef{})

}
