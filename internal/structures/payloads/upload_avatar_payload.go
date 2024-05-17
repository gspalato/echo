package payloads

type UploadAvatarPayload struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
