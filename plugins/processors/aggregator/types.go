package aggregator

type Circuit struct {
	ASide, ZSide struct{ Device, Interface, Description string }
	Role         string
	CktId        string `json:"circuit_id,omitempty"`
	Provider     string `json:",omitempty"`
	AlertId      int64  `json:",omitempty"`
}

type BgpPeer struct {
	Type            string `json:"bgp_type"`
	LocalIp         string `json:"local_ip"`
	LocalDevice     string `json:"local_device"`
	LocalInterface  string `json:"local_interface"`
	RemoteIp        string `json:"remote_ip"`
	RemoteDevice    string `json:"remote_device"`
	RemoteInterface string `json:"remote_interface"`
	AlertId         int64  `json:",omitempty"`
}
