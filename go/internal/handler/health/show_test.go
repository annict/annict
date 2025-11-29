package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/testutil"
)

func TestShow(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)

	// 設定を作成
	cfg := &config.Config{
		Env: "test",
	}

	// ハンドラーを作成
	queries := testutil.NewQueriesWithTx(db, tx)
	workRepo := repository.NewWorkRepository(queries)
	handler := NewHandler(cfg, workRepo)

	// HTTPリクエストを作成
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Show(rr, req)

	// ステータスコードを確認
	if rr.Code != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Content-Typeを確認
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("wrong content type: got %v want %v", ct, "application/json")
	}

	// レスポンスボディをパース
	var response map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// レスポンスの内容を確認
	if response["status"] != "ok" {
		t.Errorf("wrong status: got %v want %v", response["status"], "ok")
	}
	if response["database"] != "healthy" {
		t.Errorf("wrong database status: got %v want %v", response["database"], "healthy")
	}
	if response["env"] != "test" {
		t.Errorf("wrong env: got %v want %v", response["env"], "test")
	}
}
