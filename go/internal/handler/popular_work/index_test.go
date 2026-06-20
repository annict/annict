package popular_work

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

// TestIndex は人気作品ページのテスト（templ対応）
func TestIndex(t *testing.T) {
	t.Parallel()

	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTx(t)

	// テストデータを作成
	workID1 := testutil.NewWorkBuilder(t, tx).
		WithTitle("人気アニメ1").
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	// 画像付きのWork作成
	testutil.NewWorkImageBuilder(t, tx, workID1).Build()

	workID2 := testutil.NewWorkBuilder(t, tx).
		WithTitle("人気アニメ2").
		WithSeason(2024, testutil.SeasonSummer).
		Build()

	workID3 := testutil.NewWorkBuilder(t, tx).
		WithTitle("人気アニメ3").
		WithSeason(2023, testutil.SeasonAutumn).
		Build()

	// sqlcリポジトリを作成（トランザクションを使用）
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		Env: "test",
	}

	// WorkRepository, CastRepository, StaffRepository とUseCaseを作成
	workRepo := repository.NewWorkRepository(queries)
	castRepo := repository.NewCastRepository(queries)
	staffRepo := repository.NewStaffRepository(queries)
	getPopularWorksUC := usecase.NewGetPopularWorksUsecase(workRepo, castRepo, staffRepo)

	// ハンドラーを作成（templ対応版）
	handler := NewHandler(cfg, getPopularWorksUC, testutil.NewTestImageHelper())

	// HTTPリクエストとレスポンスレコーダーを作成
	req, err := http.NewRequest("GET", "/works/popular", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Index(rr, req)

	// ステータスコードを確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// レスポンスボディを確認
	body := rr.Body.String()

	// 期待される内容が含まれているか確認
	expectedContents := []string{
		"人気アニメ",  // ページタイトル
		"人気アニメ1", // 作品タイトル1
		"人気アニメ2", // 作品タイトル2
		"人気アニメ3", // 作品タイトル3
		"100人",   // デフォルトのwatchers数（デフォルトロケールは日本語）
		"2024",   // シーズン年
		"👥",      // アイコン
		`href="/works/` + workID1.String() + `"`,
		`href="/works/` + workID2.String() + `"`,
		`href="/works/` + workID3.String() + `"`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}

	// Content-Typeヘッダーを確認
	expectedContentType := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("handler returned wrong content-type: got %v want %v",
			ct, expectedContentType)
	}
}

// TestIndexEmptyResult は結果が空の場合のテスト（templ対応）
func TestIndexEmptyResult(t *testing.T) {
	t.Parallel()

	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTx(t)

	// データは作成しない（空の結果をテスト）

	// sqlcリポジトリを作成（トランザクションを使用）
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		Env: "test",
	}

	// WorkRepository, CastRepository, StaffRepository とUseCaseを作成
	workRepo := repository.NewWorkRepository(queries)
	castRepo := repository.NewCastRepository(queries)
	staffRepo := repository.NewStaffRepository(queries)
	getPopularWorksUC := usecase.NewGetPopularWorksUsecase(workRepo, castRepo, staffRepo)

	// ハンドラーを作成（templ対応版）
	handler := NewHandler(cfg, getPopularWorksUC, testutil.NewTestImageHelper())

	// HTTPリクエストとレスポンスレコーダーを作成
	req, err := http.NewRequest("GET", "/works/popular", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Index(rr, req)

	// ステータスコードを確認（空でも200 OKを返す）
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// レスポンスボディを確認
	body := rr.Body.String()

	// ページタイトルは表示される
	if !strings.Contains(body, "人気アニメ") {
		t.Errorf("response doesn't contain page title")
	}

	// 作品が1つも表示されないことを確認（カードが表示されない）
	// templではループがない場合、card要素自体が出力されない
	if strings.Contains(body, `class="card`) {
		t.Errorf("response contains work card when it shouldn't")
	}

	// Content-Typeヘッダーを確認
	expectedContentType := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("handler returned wrong content-type: got %v want %v",
			ct, expectedContentType)
	}
}

// BenchmarkIndex は人気作品ページのベンチマーク
func BenchmarkIndex(b *testing.B) {
	// ベンチマーク用のDBセットアップ
	db, tx := testutil.SetupTx(&testing.T{})

	// 100件のテストデータを作成
	for i := 0; i < 100; i++ {
		workID := testutil.NewWorkBuilder(&testing.T{}, tx).
			WithTitle(fmt.Sprintf("ベンチマーク作品%d", i)).
			Build()

		if i%10 == 0 { // 10件に1件は画像付き
			testutil.NewWorkImageBuilder(&testing.T{}, tx, workID).Build()
		}
	}

	// sqlcリポジトリを作成
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		Env: "test",
	}

	// WorkRepository, CastRepository, StaffRepository とUseCaseを作成
	workRepo := repository.NewWorkRepository(queries)
	castRepo := repository.NewCastRepository(queries)
	staffRepo := repository.NewStaffRepository(queries)
	getPopularWorksUC := usecase.NewGetPopularWorksUsecase(workRepo, castRepo, staffRepo)

	handler := NewHandler(cfg, getPopularWorksUC, testutil.NewTestImageHelper())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/works/popular", nil)

		rr := httptest.NewRecorder()
		handler.Index(rr, req)
	}
}
