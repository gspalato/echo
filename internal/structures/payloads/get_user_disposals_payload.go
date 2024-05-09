package payloads

import "unreal.sh/echo/internal/structures"

type GetUserDisposalsPayload struct {
	UserDisposals []structures.DisposalClaim `json:"user_disposals"`
}
