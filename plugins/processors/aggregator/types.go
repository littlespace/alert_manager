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
