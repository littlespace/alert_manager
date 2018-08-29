package aggregator

type Device struct {
	Device string
	Site   string
	Region string
}

type Interface struct {
	Device      string
	Interface   string
	Description string
	Agg         string
	TermSide    string
	CktId       string
	Cid         float64
}

type Circuit struct {
	ASide, ZSide *Interface
	CktId        string `json:"circuit_id"`
	Provider     string
	AlertId      int64 `json:",omitempty"`
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
