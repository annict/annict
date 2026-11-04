package openapi_description_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/annict/annict/go/api"
	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apihandler/openapi_description"
)

func TestHandler_GetOpenAPIDescription(t *testing.T) {
	t.Parallel()

	h := openapi_description.NewHandler(api.OpenAPIDescription)

	res, err := h.GetOpenAPIDescription(context.Background(), apigen.GetOpenAPIDescriptionRequestObject{})
	if err != nil {
		t.Fatalf("GetOpenAPIDescription()のエラー = %v", err)
	}
	yamlRes, ok := res.(apigen.GetOpenAPIDescription200ApplicationyamlResponse)
	if !ok {
		t.Fatalf("レスポンスの型 = %T、期待値 = GetOpenAPIDescription200ApplicationyamlResponse", res)
	}

	body, err := io.ReadAll(yamlRes.Body)
	if err != nil {
		t.Fatalf("本文を読めない: %v", err)
	}
	// 生成コードに埋め込んだ記述ではなく、正本のYAML (operationIdが書き換わっていないもの) を返す
	if !bytes.Equal(body, api.OpenAPIDescription) {
		t.Error("本文がapi/openapi.yamlと一致しない")
	}
	if !bytes.Contains(body, []byte("operationId: getOpenAPIDescription")) {
		t.Error("本文のoperationIdが正本のものではない")
	}
	if yamlRes.ContentLength != int64(len(api.OpenAPIDescription)) {
		t.Errorf("ContentLength = %d、期待値 = %d", yamlRes.ContentLength, len(api.OpenAPIDescription))
	}
}
