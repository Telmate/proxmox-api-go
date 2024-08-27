package proxmox

import (
	"errors"
	"regexp"
	"strconv"
)

type SerialID uint8

const (
	SerialID0               SerialID = 0
	SerialID1               SerialID = 1
	SerialID2               SerialID = 2
	SerialID3               SerialID = 3
	SerialID_Errors_Invalid string   = "serial id must be one of 0,1,2,3"
)

func (id SerialID) String() string {
	return strconv.Itoa(int(id))
}

func (id SerialID) Validate() error {
	if id > 3 {
		return errors.New(SerialID_Errors_Invalid)
	}
	return nil
}

type SerialInterface struct {
	Delete bool       `json:"delete,omitempty"` // If true, the serial adapter will be removed.
	Path   SerialPath `json:"path,omitempty"`   // Path to the serial device. Mutually exclusive with socket.
	Socket bool       `json:"socket,omitempty"` // If true, the serial device is a socket. Mutually exclusive with path.
}

const (
	SerialInterface_Errors_MutualExclusive string = "path and socket are mutually exclusive"
	SerialInterface_Errors_Empty           string = "path or socket must be set"
)

var regexSerialPortPath = regexp.MustCompile(`^/dev/.+$`)

func (port SerialInterface) mapToAPI(id SerialID, params map[string]interface{}) {
	tmpPath := "socket"
	if !port.Socket {
		tmpPath = string(port.Path)
	}
	params["serial"+id.String()] = tmpPath
}

func (SerialInterface) mapToSDK(v string) SerialInterface {
	if v == "socket" {
		return SerialInterface{
			Socket: true}
	}
	return SerialInterface{
		Path: SerialPath(v),
	}
}

func (port SerialInterface) Validate() error {
	if port.Delete {
		return nil
	}
	if port.Path != "" && port.Socket {
		return errors.New(SerialInterface_Errors_MutualExclusive)
	}
	if !port.Socket {
		if port.Path == "" {
			return errors.New(SerialInterface_Errors_Empty)
		}
		if err := port.Path.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type SerialInterfaces map[SerialID]SerialInterface

func (config SerialInterfaces) mapToAPI(current SerialInterfaces, params map[string]interface{}) (delete string) {
	if len(current) != 0 { // Update
		for id, port := range config {
			if _, ok := current[id]; ok {
				if port.Delete {
					delete += ",serial" + id.String()
					continue
				}
				if current[id].Path != port.Path || current[id].Socket != port.Socket {
					port.mapToAPI(id, params)
				}
			} else if !port.Delete {
				port.mapToAPI(id, params)
			}
		}
		return
	}
	// Create
	for id, port := range config {
		if !port.Delete {
			port.mapToAPI(id, params)
		}
	}
	return
}

func (SerialInterfaces) mapToSDK(params map[string]interface{}) SerialInterfaces {
	Serials := SerialInterfaces{}
	if v, isSet := params["serial0"]; isSet {
		Serials[SerialID0] = SerialInterface{}.mapToSDK(v.(string))
	}
	if v, isSet := params["serial1"]; isSet {
		Serials[SerialID1] = SerialInterface{}.mapToSDK(v.(string))
	}
	if v, isSet := params["serial2"]; isSet {
		Serials[SerialID2] = SerialInterface{}.mapToSDK(v.(string))
	}
	if v, isSet := params["serial3"]; isSet {
		Serials[SerialID3] = SerialInterface{}.mapToSDK(v.(string))
	}
	if len(Serials) > 0 {
		return Serials
	}
	return nil
}

func (s SerialInterfaces) Validate() error {
	for id, e := range s {
		if err := id.Validate(); err != nil {
			return err
		}
		if err := e.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type SerialPath string

const SerialPath_Errors_Invalid string = "path must start with /dev/"

func (path SerialPath) Validate() error {
	matches, _ := regexp.MatchString(regexSerialPortPath.String(), string(path))
	if !matches {
		return errors.New(SerialPath_Errors_Invalid)
	}
	return nil
}
