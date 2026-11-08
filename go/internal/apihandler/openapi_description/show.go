package openapi_description

import (
	"bytes"
	"context"

	"github.com/annict/annict/go/internal/apigen"
)

// GetOpenAPIDescriptionはOpenAPI記述をYAMLで返す (GET /api/v2/openapi.yaml)
func (h *Handler) GetOpenAPIDescription(_ context.Context, _ apigen.GetOpenAPIDescriptionRequestObject) (apigen.GetOpenAPIDescriptionResponseObject, error) {
	return apigen.GetOpenAPIDescription200ApplicationyamlResponse{
		Body:          bytes.NewReader(h.description),
		ContentLength: int64(len(h.description)),
	}, nil
}
