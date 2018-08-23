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
	LocalIp      string `json:"local_ip"`
	LocalDevice  string `json:"local_device"`
	RemoteIp     string `json:"remote_ip"`
	RemoteDevice string `json:"remote_device"`
	AlertId      int64  `json:",omitempty"`
}

func groupBySession(p []BgpPeer) [][]BgpPeer {
	groups := [][]BgpPeer{[]BgpPeer{p[0]}}
	for i := 1; i < len(p); i++ {
		var found bool
		for j := 0; j < len(groups); j++ {
			if p[i].LocalDevice == groups[j][0].RemoteDevice && p[i].RemoteDevice == groups[j][0].LocalDevice {
				found = true
				groups[j] = append(groups[j], p[i])
				break
			}

		}
		if !found {
			groups = append(groups, []BgpPeer{p[i]})
		}
	}
	return groups
}
