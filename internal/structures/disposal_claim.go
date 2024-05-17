package structures

type DisposalClaim struct {
	Id         string     `json:"id"          bson:"_id,omitempty"`
	UserId     string     `json:"user_id"     bson:"user_id"`
	OperatorId string     `json:"operator_id" bson:"operator_id"`
	Token      string     `json:"token"       bson:"token"`
	Credits    float32    `json:"credits"     bson:"credits"`
	IsClaimed  bool       `json:"is_claimed"  bson:"is_claimed"`
	Disposals  []Disposal `json:"disposals"   bson:"disposals"`
	Weight     float32    `json:"weight"      bson:"weight"`
}
