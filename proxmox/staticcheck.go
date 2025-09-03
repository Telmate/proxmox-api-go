package proxmox

import "context"

// _SuppressStaticCheck only exists to suppress warnings for unused functions
func X_SuppressStaticCheck_DoNotUse() {
	_ = mockClientAPI{}.new()

	ca := &clientAPI{}
	_ = ca.post(context.Background(), "", nil)
	_, _ = ca.postTask(context.Background(), "", nil)

}
