package structures

type TransactionType int

const (
	CLAIM TransactionType = iota
	SPEND
)

type Transaction struct {
	TransactionType TransactionType `json:"transaction_type"`
	UserId          string          `json:"user_id"`
	ClaimId         string          `json:"claim_id"`
	Credits         float64         `json:"credits"`
	Timestamp       int64           `json:"timestamp"`
	Description     string          `json:"description"`
}
