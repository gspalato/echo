package payloads

type GetAvatarPayload struct {
	Success   bool   `json:"success"`
	AvatarUrl string `json:"avatar_url"`
	Error     string `json:"error"`
}
