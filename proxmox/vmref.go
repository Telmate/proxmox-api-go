package proxmox

import (
	"context"
	"crypto/des"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	vmRefQemu                     string = "qemu"
	vmRefLXC                      string = "lxc"
	clone_Error_MutuallyExclusive string = "linked and full clone are mutually exclusive"
	clone_Error_NoneSet           string = "either linked nor full clone must be set"
)

var (
	bufCopy = NewBufCopy()
)

// CloneLxc clones a new LXC container by cloning current container
func (vmr *VmRef) CloneLxc(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	err := settings.Validate()
	if err != nil {
		return nil, err
	}
	return vmr.cloneLxc_Unsafe(ctx, settings, c)
}

// CloneLxcNoCheck creates a new LXC container by cloning the current container, without input validation.
func (vmr *VmRef) CloneLxcNoCheck(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	return vmr.cloneLxc_Unsafe(ctx, settings, c)
}

func (vmr *VmRef) cloneLxc_Unsafe(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, error) {
	id, node, pool, params := settings.mapToAPI()
	var err error
	url := "/nodes/" + vmr.node.String() + "/lxc/" + vmr.vmId.String() + "/clone"
	if id == 0 {
		id, err = guestCreateLoop_Unsafe(ctx, "newid", url, params, nil, c)
	} else {
		_, err = c.PostWithTask(ctx, params, url)
	}
	if err != nil {
		return nil, err
	}
	return &VmRef{
		vmId:   id,
		node:   node,
		pool:   pool,
		vmType: GuestLxc}, nil
}

// CloneQemu creates a new Qemu VM by cloning the current VM.
func (vmr *VmRef) CloneQemu(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	err := settings.Validate()
	if err != nil {
		return nil, err
	}
	return vmr.cloneQemu_Unsafe(ctx, settings, c)
}

// CloneQemuNoCheck creates a new VM by cloning the current VM, without input validation.
func (vmr *VmRef) CloneQemuNoCheck(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	return vmr.cloneQemu_Unsafe(ctx, settings, c)
}

func (vmr *VmRef) cloneQemu_Unsafe(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, error) {
	id, node, pool, params := settings.mapToAPI()
	var err error
	url := "/nodes/" + vmr.node.String() + "/qemu/" + vmr.vmId.String() + "/clone"
	if id == 0 {
		id, err = guestCreateLoop_Unsafe(ctx, "newid", url, params, nil, c)
	} else {
		_, err = c.PostWithTask(ctx, params, url)
	}
	if err != nil {
		return nil, err
	}
	return &VmRef{
		vmId:   id,
		node:   node,
		pool:   pool,
		vmType: GuestQemu}, nil
}

func (vmr VmRef) Delete(ctx context.Context, c *Client) error { return c.new().guestDelete(ctx, &vmr) }

func (c *clientNewTest) guestDelete(ctx context.Context, vmr *VmRef) error {
	guestID := vmr.VmId()
	if guestID == 0 {
		return errors.New(VmRef_Error_IDnotSet)
	}
	ca := c.apiGet()
	rawGuests, err := listGuests_Unsafe(ctx, ca)
	if err != nil {
		return err
	}

	rawGuest, ok := rawGuests.SelectID(guestID)
	if !ok {
		return errorMsg{}.guestDoesNotExist(vmr.vmId)
	}
	if haState := rawGuest.GetHaState(); haState != "" {
		if _, err = guestID.deleteHaResource(ctx, ca); err != nil {
			return err
		}
	}

	guestType := rawGuest.GetType()
	vmr.node = rawGuest.GetNode()
	vmr.vmType = guestType

	var protection bool // Check if guest is protected
	switch guestType {
	case GuestQemu:
		rawConfig, err := guestGetRawQemuConfig_Unsafe(ctx, vmr, ca)
		if err != nil {
			return err
		}
		protection = rawConfig.GetProtection()
	case GuestLxc:
		rawConfig, err := guestGetLxcRawConfig_Unsafe(ctx, vmr, ca)
		if err != nil {
			return err
		}
		protection = rawConfig.GetProtection()
	}
	if protection {
		return errorMsg{}.guestIsProtectedCantDelete(guestID)
	}

	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}

	if rawGuest.GetStatus() != PowerStateStopped { // Check if guest is running
		for {
			var guestStatus RawGuestStatus
			guestStatus, err = vmr.getRawGuestStatus_Unsafe(ctx, c.oldClient)
			if err != nil {
				return err
			}
			if guestStatus.GetState() == PowerStateStopped {
				break
			}
			if version.Encode() >= version_8_0_0 { // Try to force stop the guest if supported
				err = vmr.forceStop_Unsafe(ctx, ca)
			} else {
				err = vmr.stop_Unsafe(ctx, ca)
			}
			if err != nil {
				return err
			}
		}
	}
	return vmr.delete_Unsafe(ctx, c.oldClient)
}

func (vmr VmRef) DeleteNoCheck(ctx context.Context, c *Client) error {
	if err := c.checkInitialized(); err != nil {
		return err
	}
	return vmr.delete_Unsafe(ctx, c)
}

func (vmr *VmRef) delete_Unsafe(ctx context.Context, c *Client) error {
	_, err := c.DeleteVmParams(ctx, vmr, map[string]interface{}{"destroy-unreferenced-disks": true, "purge": true}) // TODO use a more optimized version
	return err
}

// ForceStop stops the guest immediately without a graceful shutdown and cancels any stop/shutdown operations in progress.
// This function requires Proxmox VE 8.0 or later.
func (vmr *VmRef) ForceStop(ctx context.Context, c *Client) error {
	return c.new().guestStopForce(ctx, vmr)
}

func (c *clientNewTest) guestStopForce(ctx context.Context, vmr *VmRef) error {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}
	if version.Encode() < version_8_0_0 {
		return functionalityNotSupportedInVersion("force stop", version)
	}
	if err := c.oldClient.CheckVmRef(ctx, vmr); err != nil {
		return err
	}
	return vmr.forceStop_Unsafe(ctx, c.apiGet())
}

func (vmr *VmRef) forceStop_Unsafe(ctx context.Context, c clientApiInterface) error {
	return c.updateGuestStatus(ctx, vmr, "stop", map[string]any{"overrule-shutdown": int(1)})
}

func (vmr *VmRef) GetRawGuestStatus(ctx context.Context, c *Client) (RawGuestStatus, error) {
	if err := c.checkInitialized(); err != nil {
		return nil, err
	}
	err := c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return vmr.getRawGuestStatus_Unsafe(ctx, c)
}

func (vmr *VmRef) getRawGuestStatus_Unsafe(ctx context.Context, c *Client) (RawGuestStatus, error) {
	return c.GetItemConfigMapStringInterface(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/status/current", "vm", "STATE")
}

func (vmr *VmRef) Migrate(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) error {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	if err := c.checkInitialized(); err != nil {
		return err
	}
	if err := newNode.Validate(); err != nil {
		return err
	}
	return vmr.migrate_Unsafe(ctx, c, newNode, LiveMigrate)
}

func (vmr *VmRef) MigrateNoCheck(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) error {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	if err := c.checkInitialized(); err != nil {
		return err
	}
	return vmr.migrate_Unsafe(ctx, c, newNode, LiveMigrate)
}

func (vmr *VmRef) migrate_Unsafe(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) error {
	params := map[string]interface{}{
		"target":           newNode.String(),
		"with-local-disks": 1,
	}
	if LiveMigrate {
		params["online"] = 1
	}
	_, err := c.PostWithTask(ctx, params, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/migrate")
	return err
}

func (vmr *VmRef) PendingChanges(ctx context.Context, c *Client) (bool, error) {
	return c.new().guestCheckPendingChanges(ctx, vmr)
}

func (vmr *VmRef) pendingChanges(ctx context.Context, c clientApiInterface) (bool, error) {
	changes, err := c.getGuestPendingChanges(ctx, vmr)
	if err != nil {
		return false, err
	}
	for _, item := range changes {
		m := item.(map[string]any)
		// we always have the key `key`
		if _, ok := m[pendingApiValueKey]; ok {
			if len(m) > 2 {
				return true, nil
			}
		} else if len(m) > 1 {
			return true, nil
		}
	}
	return false, nil
}

func (c *clientNewTest) guestCheckPendingChanges(ctx context.Context, vmr *VmRef) (bool, error) {
	return vmr.pendingChanges(ctx, c.apiGet())
}

func (vmr *VmRef) pendingConfig(ctx context.Context, c clientApiInterface) (map[string]any, bool, error) {
	changes, err := c.getGuestPendingChanges(ctx, vmr)
	if err != nil {
		return nil, false, err
	}
	var pending bool
	config := make(map[string]any, len(changes))
	for _, item := range changes {
		m := item.(map[string]any)
		// we always have the key `key`
		if v, ok := m[pendingApiValueKey]; ok {
			config[m[pendingApiKeyKey].(string)] = v
			if len(m) > 2 {
				pending = true
			}
		} else if len(m) > 1 {
			pending = true
		}
	}
	return config, pending, nil
}

const (
	pendingApiKeyKey   string = "key"
	pendingApiValueKey string = "value"
)

func (vmr *VmRef) Stop(ctx context.Context, c *Client) error { return c.new().guestStop(ctx, vmr) }

func (c *clientNewTest) guestStop(ctx context.Context, vmr *VmRef) error {
	if err := c.oldClient.CheckVmRef(ctx, vmr); err != nil {
		return err
	}
	return vmr.stop_Unsafe(ctx, c.apiGet())
}

func (vmr *VmRef) stop_Unsafe(ctx context.Context, c clientApiInterface) error {
	return c.updateGuestStatus(ctx, vmr, "stop", nil)
}

func (vmr *VmRef) TermProxyWebsocketServeHTTP(c *Client, w http.ResponseWriter, r *http.Request, responseHeader http.Header) (err error) {

	ticket, err := c.CreateTermProxy(context.Background(), vmr, map[string]interface{}{})
	if nil != err {
		return err
	}

	path := fmt.Sprintf("/nodes/%s/qemu/%s/vncwebsocket?port=%d&vncticket=%s",
		vmr.Node().String(), vmr.vmId.String(), ticket.Port, url.QueryEscape(ticket.Ticket))

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	websocketServe, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		err = fmt.Errorf("upgrade http to websocket err: %+v", err)
		fmt.Println(err)
		return
	}

	if strings.HasPrefix(path, "/") {
		path = strings.Replace(c.ApiUrl, "https://", "wss://", 1) + path
	}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ReadBufferSize:  1024 * 10,
		WriteBufferSize: 1024 * 1024 * 4,
	}

	headers := c.session.BuildHeaders(nil)
	pveVncConn, _, err := dialer.Dial(path, headers)

	if err != nil {
		err = fmt.Errorf("connect to pve err: %+v", err)
		_ = websocketServe.Close()
		return
	}

	defer func() {
		_ = pveVncConn.Close()
		_ = websocketServe.Close()
	}()

	//authMsg := ticket.User + ":" + ticket.Password+ "\n"
	/*authMsg :=  "root:ogsFPGsTCj8Br0TI\n"

		err = pveVncConn.WriteMessage(websocket.TextMessage, []byte(authMsg))
	fmt.Println("pveVncConn write err:",err)
		if err != nil {
			return
		}*/

	go func() {
		for {
			_, err = bufCopy.Copy(websocketServe.NetConn(), pveVncConn.NetConn())
			if err != nil {
				err = fmt.Errorf("buf copy pve to websocket err: %+v", err)
				_ = pveVncConn.Close()
				_ = websocketServe.Close()
				return
			}
		}
	}()

	for {
		_, err = bufCopy.Copy(pveVncConn.NetConn(), websocketServe.NetConn())
		if err != nil {
			err = fmt.Errorf("buf copy websocket to pve err: %+v", err)
			_ = pveVncConn.Close()
			_ = websocketServe.Close()
			return
		}
	}

	return
}

func (vmr *VmRef) VNCProxyWebsocketServeHTTP(c *Client, w http.ResponseWriter, r *http.Request, responseHeader http.Header) (err error) {
	ticket, err := c.CreateVNCProxy(context.Background(), vmr, map[string]interface{}{
		"websocket":         true,
		"generate-password": false,
	})
	if nil != err {
		return err
	}

	path := fmt.Sprintf("/nodes/%s/qemu/%s/vncwebsocket?port=%d&vncticket=%s",
		vmr.Node().String(), vmr.vmId.String(), ticket.Port, url.QueryEscape(ticket.Ticket))

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	websocketServe, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		err = fmt.Errorf("upgrade http to websocket err: %+v", err)
		return
	}

	if strings.HasPrefix(path, "/") {
		path = strings.Replace(c.ApiUrl, "https://", "wss://", 1) + path
	}

	var tlsConfig *tls.Config
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		//DisableKeepAlives: true,
	}
	if transport != nil {
		tlsConfig = transport.TLSClientConfig
	}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		TLSClientConfig:  tlsConfig,
	}

	headers := c.session.BuildHeaders(nil)
	pveVncConn, _, err := dialer.Dial(path, headers)

	if err != nil {
		err = fmt.Errorf("connect to pve err: %+v", err)
		_ = websocketServe.Close()
		return
	}

	defer func() {
		_ = pveVncConn.Close()
		_ = websocketServe.Close()
	}()

	var msgType int
	var msg []byte

	/*err = pveVncConn.WriteMessage(websocket.TextMessage, []byte(ticket.Ticket+"\n"))
	if err != nil {
		err = fmt.Errorf("write vncticket message err: %+v", err)
		return
	}*/
	msgType, msg, err = pveVncConn.ReadMessage()
	//fmt.Println("pveVncConn1", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read pve rfb message err: %+v", err)
		return
	}
	//fmt.Println("pveVncConn1", msgType, msg, err, string(msg))
	//"RFB 003.008\n"
	if err = websocketServe.WriteMessage(msgType, msg); err != nil {
		err = fmt.Errorf("write websocket rfb message err: %+v", err)
		return
	}

	msgType, msg, err = websocketServe.ReadMessage()
	//fmt.Println("websocketServe2", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read websocket rfb message err: %+v", err)
		return
	}
	//"RFB 003.008\n"
	if err = pveVncConn.WriteMessage(msgType, msg); err != nil {
		err = fmt.Errorf("write pve rfb message err: %+v", err)
		return
	}
	msgType, msg, err = pveVncConn.ReadMessage()
	//fmt.Println("pveVncConn3", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read websocket auth type message err: %+v", err)
		return
	}
	//[]uint8{1,2}  type 2 is need password
	if err = pveVncConn.WriteMessage(websocket.BinaryMessage, []uint8{2}); err != nil {
		err = fmt.Errorf("write pve auth type message err: %+v", err)
		return
	}

	msgType, msg, err = pveVncConn.ReadMessage()
	//fmt.Println("pveVncConn4", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read pve auth random key message err: %+v", err)
		return
	}
	//[]unit8{...}  len 16
	enPassword, err := VNCAuthPasswordEncrypt(ticket.Ticket, msg)
	//fmt.Println("enPassword", enPassword, err, string(enPassword))
	if err = pveVncConn.WriteMessage(websocket.BinaryMessage, enPassword); err != nil {
		err = fmt.Errorf("write pve auth password message err: %+v", err)
		return
	}
	//msgType, msg, err = pveVncConn.ReadMessage()
	//fmt.Println("pveVncConn5", msgType, msg, err, string(msg))

	//send websocket do not need password
	if err = websocketServe.WriteMessage(websocket.BinaryMessage, []uint8{1, 1}); err != nil {
		err = fmt.Errorf("write websocket auth type message err: %+v", err)
		return
	}
	msgType, msg, err = websocketServe.ReadMessage()
	//fmt.Println("websocketServe6", msgType, msg, err, string(msg))
	if nil != err {
		err = fmt.Errorf("read websocket auth type return message err: %+v", err)
		return
	}

	go func() {

		for {
			_, err = bufCopy.Copy(websocketServe.NetConn(), pveVncConn.NetConn())
			if err != nil {
				err = fmt.Errorf("buf copy pve to websocket err: %+v", err)
				_ = pveVncConn.Close()
				_ = websocketServe.Close()
				return
			}
		}
	}()

	for {
		_, err = bufCopy.Copy(pveVncConn.NetConn(), websocketServe.NetConn())
		if err != nil {
			err = fmt.Errorf("buf copy websocket to pve err: %+v", err)
			_ = pveVncConn.Close()
			_ = websocketServe.Close()
			return
		}
	}

	return
}

func VNCAuthPasswordEncrypt(key string, bytes []byte) ([]byte, error) {
	keyBytes := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	if len(key) > 8 {
		key = key[:8]
	}

	for i := 0; i < len(key); i++ {
		keyBytes[i] = ReverseBits(key[i])
	}

	block, err := des.NewCipher(keyBytes)

	if err != nil {
		return nil, err
	}

	result1 := make([]byte, 8)
	block.Encrypt(result1, bytes)
	result2 := make([]byte, 8)
	block.Encrypt(result2, bytes[8:])

	crypted := append(result1, result2...)

	return crypted, nil
}
func ReverseBits(b byte) byte {
	var reverse = [256]int{
		0, 128, 64, 192, 32, 160, 96, 224,
		16, 144, 80, 208, 48, 176, 112, 240,
		8, 136, 72, 200, 40, 168, 104, 232,
		24, 152, 88, 216, 56, 184, 120, 248,
		4, 132, 68, 196, 36, 164, 100, 228,
		20, 148, 84, 212, 52, 180, 116, 244,
		12, 140, 76, 204, 44, 172, 108, 236,
		28, 156, 92, 220, 60, 188, 124, 252,
		2, 130, 66, 194, 34, 162, 98, 226,
		18, 146, 82, 210, 50, 178, 114, 242,
		10, 138, 74, 202, 42, 170, 106, 234,
		26, 154, 90, 218, 58, 186, 122, 250,
		6, 134, 70, 198, 38, 166, 102, 230,
		22, 150, 86, 214, 54, 182, 118, 246,
		14, 142, 78, 206, 46, 174, 110, 238,
		30, 158, 94, 222, 62, 190, 126, 254,
		1, 129, 65, 193, 33, 161, 97, 225,
		17, 145, 81, 209, 49, 177, 113, 241,
		9, 137, 73, 201, 41, 169, 105, 233,
		25, 153, 89, 217, 57, 185, 121, 249,
		5, 133, 69, 197, 37, 165, 101, 229,
		21, 149, 85, 213, 53, 181, 117, 245,
		13, 141, 77, 205, 45, 173, 109, 237,
		29, 157, 93, 221, 61, 189, 125, 253,
		3, 131, 67, 195, 35, 163, 99, 227,
		19, 147, 83, 211, 51, 179, 115, 243,
		11, 139, 75, 203, 43, 171, 107, 235,
		27, 155, 91, 219, 59, 187, 123, 251,
		7, 135, 71, 199, 39, 167, 103, 231,
		23, 151, 87, 215, 55, 183, 119, 247,
		15, 143, 79, 207, 47, 175, 111, 239,
		31, 159, 95, 223, 63, 191, 127, 255,
	}

	return byte(reverse[int(b)])
}

const (
	cloneLxcFlagName  string = "hostname"
	cloneQemuFlagName string = "name"
)

type CloneLxcTarget struct {
	Full   *CloneLxcFull
	Linked *CloneLinked
}

const (
	CloneLxcTarget_Error_MutualExclusivity = clone_Error_MutuallyExclusive
	CloneLxcTarget_Error_NoneSet           = clone_Error_NoneSet
)

func (target CloneLxcTarget) Validate() error {
	if target.Full == nil && target.Linked == nil {
		return errors.New(CloneQemuTarget_Error_NoneSet)
	}
	if target.Full != nil && target.Linked != nil {
		return errors.New(CloneQemuTarget_Error_MutualExclusivity)
	}
	if target.Full != nil {
		return target.Full.Validate()
	}
	return target.Linked.Validate()
}

func (target CloneLxcTarget) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	if target.Linked != nil {
		return target.Linked.mapToAPI(cloneLxcFlagName)
	}
	if target.Full != nil {
		return target.Full.mapToAPI()
	}
	return 0, "", "", nil
}

type CloneQemuTarget struct {
	Full   *CloneQemuFull `json:"full,omitempty"`
	Linked *CloneLinked   `json:"linked,omitempty"`
}

const (
	CloneQemuTarget_Error_MutualExclusivity = clone_Error_MutuallyExclusive
	CloneQemuTarget_Error_NoneSet           = clone_Error_NoneSet
)

func (target CloneQemuTarget) Validate() error {
	if target.Full == nil && target.Linked == nil {
		return errors.New(CloneQemuTarget_Error_NoneSet)
	}
	if target.Full != nil && target.Linked != nil {
		return errors.New(CloneQemuTarget_Error_MutualExclusivity)
	}
	if target.Full != nil {
		return target.Full.Validate()
	}
	return target.Linked.Validate()
}

func (target CloneQemuTarget) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	if target.Linked != nil {
		return target.Linked.mapToAPI(cloneQemuFlagName)
	}
	if target.Full != nil {
		return target.Full.mapToAPI()
	}
	return 0, "", "", nil
}

// Linked Clone in the same for both LXC and QEMU
type CloneLinked struct {
	Node NodeName   `json:"node"`
	ID   *GuestID   `json:"id,omitempty"`   // Optional
	Name *GuestName `json:"name,omitempty"` // Optional
	Pool *PoolName  `json:"pool,omitempty"` // Optional
}

func (linked CloneLinked) Validate() (err error) {
	if linked.ID != nil {
		if err = linked.ID.Validate(); err != nil {
			return
		}
	}
	if linked.Name != nil {
		if err = linked.Name.Validate(); err != nil {
			return
		}
	}
	if linked.Pool != nil {
		if err = linked.Pool.Validate(); err != nil {
			return
		}
	}
	return linked.Node.Validate()
}

func (linked CloneLinked) mapToAPI(nameFlag string) (GuestID, NodeName, PoolName, map[string]interface{}) {
	return cloneSettings{
		FullClone: false,
		ID:        linked.ID,
		nameFlag:  nameFlag,
		Name:      linked.Name,
		Node:      linked.Node,
		Pool:      linked.Pool}.mapToAPI()
}

type CloneLxcFull struct {
	Node    NodeName   `json:"node"`
	ID      *GuestID   `json:"id,omitempty"`      // Optional
	Name    *GuestName `json:"name,omitempty"`    // Optional
	Pool    *PoolName  `json:"pool,omitempty"`    // Optional
	Storage *string    `json:"storage,omitempty"` // Optional // TODO replace one we have a type for it
}

func (full CloneLxcFull) Validate() (err error) {
	if full.ID != nil {
		if err = full.ID.Validate(); err != nil {
			return
		}
	}
	if full.Name != nil {
		if err = full.Name.Validate(); err != nil {
			return
		}
	}
	if full.Pool != nil {
		if err = full.Pool.Validate(); err != nil {
			return
		}
	}
	return full.Node.Validate()
}

func (full CloneLxcFull) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	return cloneSettings{
		FullClone: true,
		ID:        full.ID,
		nameFlag:  cloneLxcFlagName,
		Name:      full.Name,
		Node:      full.Node,
		Pool:      full.Pool,
		Storage:   full.Storage}.mapToAPI()
}

type CloneQemuFull struct {
	Node          NodeName        `json:"node"`
	ID            *GuestID        `json:"id,omitempty"`      // Optional
	Name          *GuestName      `json:"name,omitempty"`    // Optional
	Pool          *PoolName       `json:"pool,omitempty"`    // Optional
	Storage       *string         `json:"storage,omitempty"` // Optional // TODO replace one we have a type for it
	StorageFormat *QemuDiskFormat `json:"format,omitempty"`  // Optional
}

func (full CloneQemuFull) Validate() (err error) {
	if full.ID != nil {
		if err = full.ID.Validate(); err != nil {
			return
		}
	}
	if full.Name != nil {
		if err = full.Name.Validate(); err != nil {
			return
		}
	}
	if full.Pool != nil {
		if err = full.Pool.Validate(); err != nil {
			return
		}
	}
	if full.StorageFormat != nil {
		if err = full.StorageFormat.Validate(); err != nil {
			return
		}
	}
	return full.Node.Validate()
}

func (full CloneQemuFull) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	return cloneSettings{
		FullClone:     true,
		ID:            full.ID,
		nameFlag:      cloneQemuFlagName,
		Name:          full.Name,
		Node:          full.Node,
		Pool:          full.Pool,
		Storage:       full.Storage,
		StorageFormat: full.StorageFormat}.mapToAPI()
}

type cloneSettings struct {
	FullClone     bool
	ID            *GuestID
	nameFlag      string
	Name          *GuestName
	Node          NodeName
	Pool          *PoolName
	Storage       *string // TODO replace one we have a type for it
	StorageFormat *QemuDiskFormat
}

func (clone cloneSettings) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	params := map[string]interface{}{
		"target": clone.Node.String(),
		"full":   clone.FullClone,
	}
	var id GuestID
	if clone.ID != nil {
		id = *clone.ID
		params["newid"] = int(id)
	}
	if clone.Name != nil {
		params[clone.nameFlag] = (*clone.Name).String()
	}
	var pool PoolName
	if clone.Pool != nil {
		pool = *clone.Pool
		params["pool"] = pool.String()
	}
	if clone.Storage != nil {
		params["storage"] = *clone.Storage
	}
	if clone.StorageFormat != nil {
		params["format"] = clone.StorageFormat.String()
	}
	return id, clone.Node, pool, params
}

// TODO add more properties to GuestStatus
type GuestStatus struct {
	Name   GuestName     `json:"name"`
	State  PowerState    `json:"state"`
	Uptime time.Duration `json:"uptime"`
}

type RawGuestStatus map[string]any

func (raw RawGuestStatus) GetName() GuestName {
	if v, isSet := raw["name"]; isSet {
		if name, ok := v.(string); ok {
			return GuestName(name)
		}
	}
	return ""
}

func (raw RawGuestStatus) Get() GuestStatus {
	return GuestStatus{
		Name:   raw.GetName(),
		State:  raw.GetState(),
		Uptime: raw.GetUptime()}
}

func (raw RawGuestStatus) GetState() PowerState {
	if v, isSet := raw["status"]; isSet {
		if state, ok := v.(string); ok {
			return PowerState(0).parse(state)
		}
	}
	return PowerStateUnknown
}

func (raw RawGuestStatus) GetUptime() time.Duration {
	if v, isSet := raw["uptime"]; isSet {
		if uptime, ok := v.(float64); ok {
			return time.Duration(uptime) * time.Second
		}
	}
	return 0
}
