package structures

type User struct {
	Id           string        `json:"id"           bson:"_id,omitempty"`
	Name         string        `json:"name"         bson:"name"`
	Username     string        `json:"username"     bson:"username"`
	Credits      float64       `json:"credits"      bson:"credits"`
	IsOperator   bool          `json:"is_operator"  bson:"is_operator"`
	Transactions []Transaction `json:"transactions" bson:"transactions"`
	PasswordHash string        `json:"-"            bson:"password_hash"`
}

type Profile struct {
	Name         string        `json:"name"`
	Username     string        `json:"username"`
	Credits      float64       `json:"credits"`
	IsOperator   bool          `json:"is_operator"`
	Transactions []Transaction `json:"transactions"`
}

func (u *User) ToProfile() *Profile {
	return &Profile{
		Name:         u.Name,
		Username:     u.Username,
		Credits:      u.Credits,
		IsOperator:   u.IsOperator,
		Transactions: u.Transactions,
	}
}
