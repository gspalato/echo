package inputs

type ClaimDisposalInput struct {
	UserToken     *string `json:"user_token"`
	DisposalToken *string `json:"disposal_token"`
}
