package types

type Device struct {
	Name   string
	Ip     string
	Site   string
	Region string `json:",omitempty"`
	Status string
}

func NewDevice() *Device {
	return &Device{}
}

type Interface struct {
	Device                        string
	Interface                     string
	Description                   string
	Role                          string
	Type                          string
	Agg                           string `json:",omitempty"`
	PeerDevice, PeerIntf, PeerAgg string
}

func NewInterface(device, intface string) *Interface {
	return &Interface{Device: device, Interface: intface}
}

type Circuit struct {
	ASide, ZSide struct {
		Device         *Device
		Interface, Agg string
	}
	Role     string
	CktId    string `json:"circuit_id,omitempty"`
	Provider string `json:",omitempty"`
	AlertId  int64  `json:",omitempty"`
}

func NewCircuit() *Circuit {
	return &Circuit{}
}

type BgpPeer struct {
	Type            string  `json:"bgp_type"`
	LocalIp         string  `json:"local_ip"`
	LocalDevice     *Device `json:"local_device"`
	LocalInterface  string  `json:"local_interface"`
	RemoteIp        string  `json:"remote_ip"`
	RemoteDevice    *Device `json:"remote_device"`
	RemoteInterface string  `json:"remote_interface"`
	AlertId         int64   `json:",omitempty"`
}

func NewBgpPeer() *BgpPeer {
	return &BgpPeer{}
}
