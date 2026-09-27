package home

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/go/internal/config"
)

func TestShow(t *testing.T) {
	t.Parallel()

	// 設定を作成
	cfg := &config.Config{
		RailsAppURL: "https://annict.com",
	}

	// ハンドラーを作成
	handler := NewHandler(cfg)

	// HTTPリクエストを作成
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Show(rr, req)

	// ステータスコードを確認 (302 Found)
	if rr.Code != http.StatusFound {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusFound)
	}

	// Locationヘッダーを確認
	location := rr.Header().Get("Location")
	if location != "https://annict.com" {
		t.Errorf("リダイレクト先 = %v、期待値 = %v", location, "https://annict.com")
	}
}
