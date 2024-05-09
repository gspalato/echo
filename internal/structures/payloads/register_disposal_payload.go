package payloads

import "unreal.sh/echo/internal/structures"

type RegisterDisposalPayload struct {
	Success  bool                     `json:"success"`
	Disposal structures.DisposalClaim `json:"disposal"`
	Error    *string                  `json:"error"`
}
