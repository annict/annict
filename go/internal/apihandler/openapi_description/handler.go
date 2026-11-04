// Package openapi_descriptionはREST API V2のOpenAPI記述を配信するハンドラーを提供する
package openapi_description

// HandlerはOpenAPI記述の配信を担当するハンドラー
type Handler struct {
	description []byte
}

// NewHandlerは新しいHandlerを作成する。
// descriptionには配信するOpenAPI記述 (`api.OpenAPIDescription`) を渡す
func NewHandler(description []byte) *Handler {
	return &Handler{description: description}
}
