package inputs

import "unreal.sh/echo/internal/structures"

type RegisterDisposalInput struct {
	Disposals     []structures.Disposal `json:"disposals"`
	OperatorToken *string               `json:"operator_token"`
}
