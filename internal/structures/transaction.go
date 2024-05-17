package structures

type TransactionType string

const (
	CLAIM TransactionType = "CLAIM"
	SPEND TransactionType = "SPEND"
)

type Transaction struct {
	TransactionType TransactionType `json:"transaction_type"`
	UserId          string          `json:"user_id"`
	ClaimId         string          `json:"claim_id"`
	Credits         float32         `json:"credits"`
	Timestamp       int64           `json:"timestamp"`
	Description     string          `json:"description"`
}
