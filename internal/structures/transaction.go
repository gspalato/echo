package structures

type TransactionType string

const (
	CLAIM TransactionType = "CLAIM"
	SPEND TransactionType = "SPEND"
)

type Transaction struct {
	TransactionType TransactionType `json:"transaction_type" bson:"transaction_type"`
	UserId          string          `json:"user_id"          bson:"user_id"`
	ClaimId         string          `json:"claim_id"         bson:"claim_id"`
	Credits         float32         `json:"credits"          bson:"credits"`
	Timestamp       int64           `json:"timestamp"        bson:"timestamp"`
	Description     string          `json:"description"      bson:"description"`
}
