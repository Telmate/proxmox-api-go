package proxmox

import "errors"

type QemuNetworkInterfaceID uint8

const (
	QemuNetworkInterfaceID_Error_Invalid string = "network interface ID must be in the range 0-31"

	QemuNetworkInterfaceID0  QemuNetworkInterfaceID = 0
	QemuNetworkInterfaceID1  QemuNetworkInterfaceID = 1
	QemuNetworkInterfaceID2  QemuNetworkInterfaceID = 2
	QemuNetworkInterfaceID3  QemuNetworkInterfaceID = 3
	QemuNetworkInterfaceID4  QemuNetworkInterfaceID = 4
	QemuNetworkInterfaceID5  QemuNetworkInterfaceID = 5
	QemuNetworkInterfaceID6  QemuNetworkInterfaceID = 6
	QemuNetworkInterfaceID7  QemuNetworkInterfaceID = 7
	QemuNetworkInterfaceID8  QemuNetworkInterfaceID = 8
	QemuNetworkInterfaceID9  QemuNetworkInterfaceID = 9
	QemuNetworkInterfaceID10 QemuNetworkInterfaceID = 10
	QemuNetworkInterfaceID11 QemuNetworkInterfaceID = 11
	QemuNetworkInterfaceID12 QemuNetworkInterfaceID = 12
	QemuNetworkInterfaceID13 QemuNetworkInterfaceID = 13
	QemuNetworkInterfaceID14 QemuNetworkInterfaceID = 14
	QemuNetworkInterfaceID15 QemuNetworkInterfaceID = 15
	QemuNetworkInterfaceID16 QemuNetworkInterfaceID = 16
	QemuNetworkInterfaceID17 QemuNetworkInterfaceID = 17
	QemuNetworkInterfaceID18 QemuNetworkInterfaceID = 18
	QemuNetworkInterfaceID19 QemuNetworkInterfaceID = 19
	QemuNetworkInterfaceID20 QemuNetworkInterfaceID = 20
	QemuNetworkInterfaceID21 QemuNetworkInterfaceID = 21
	QemuNetworkInterfaceID22 QemuNetworkInterfaceID = 22
	QemuNetworkInterfaceID23 QemuNetworkInterfaceID = 23
	QemuNetworkInterfaceID24 QemuNetworkInterfaceID = 24
	QemuNetworkInterfaceID25 QemuNetworkInterfaceID = 25
	QemuNetworkInterfaceID26 QemuNetworkInterfaceID = 26
	QemuNetworkInterfaceID27 QemuNetworkInterfaceID = 27
	QemuNetworkInterfaceID28 QemuNetworkInterfaceID = 28
	QemuNetworkInterfaceID29 QemuNetworkInterfaceID = 29
	QemuNetworkInterfaceID30 QemuNetworkInterfaceID = 30
	QemuNetworkInterfaceID31 QemuNetworkInterfaceID = 31

	QemuNetworkInterfaceIDMaximum QemuNetworkInterfaceID = QemuNetworkInterfaceID31
)

func (id QemuNetworkInterfaceID) Validate() error {
	if id > QemuNetworkInterfaceIDMaximum {
		return errors.New(QemuNetworkInterfaceID_Error_Invalid)
	}
	return nil
}
