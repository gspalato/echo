package structures

type DisposalClaim struct {
	Id         string     `json:"id" bson:"_id,omitempty"`
	UserId     string     `json:"user_id"`
	OperatorId string     `json:"operator_id"`
	Token      string     `json:"token"`
	Credits    float32    `json:"credits"`
	IsClaimed  bool       `json:"is_claimed"`
	Disposals  []Disposal `json:"disposals"`
	Weight     float32    `json:"weight"`

	/*
	   public double Credits => Disposals.Sum(x => x.Credits);
	   public float Weight => Disposals.Sum(x => x.Weight);
	*/
}
