package payloads

import "unreal.sh/echo/internal/structures"

type ClaimDisposalPayload struct {
	Success  bool                      `json:"success"`
	Error    *string                   `json:"error"`
	Disposal *structures.DisposalClaim `json:"disposal"`
}
