package structures

type Disposal struct {
	Credits      float32      `json:"credits"`
	Weight       float32      `json:"weight"`
	DisposalType DisposalType `json:"disposal_type"`
}
