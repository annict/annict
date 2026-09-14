package db_work

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	animeRepo := repository.NewAnimeRepository(queries)
	animeClassificationRepo := repository.NewAnimeClassificationRepository(queries)

	satelliteRepos := usecase.WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}

	getDBWorksUC := usecase.NewGetDBWorksUsecase(workRepo)
	getDBWorkFormOptionsUC := usecase.NewGetDBWorkFormOptionsUsecase(numberFormatRepo)
	getDBWorkEditUC := usecase.NewGetDBWorkEditUsecase(workRepo, numberFormatRepo)
	createWorkUC := usecase.NewCreateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, validator.NewDBWorkCreateValidator(workRepo, numberFormatRepo))
	updateWorkUC := usecase.NewUpdateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, validator.NewDBWorkCreateValidator(workRepo, numberFormatRepo))
	deleteWorkUC := usecase.NewDeleteWorkUsecase(db, workRepo, animeRepo)

	return NewHandler(cfg, sessionManager, testutil.NewTestFlashManager(), testutil.NewTestImageHelper(), getDBWorksUC, getDBWorkFormOptionsUC, getDBWorkEditUC, createWorkUC, updateWorkUC, deleteWorkUC)
}

// assertNotFoundPageは404が、以前http.Errorが返していた1行のプレーンテキストでは
// なく共通のエラーページとして配信されることを検証する。古いリンクを辿った読み手が、何が
// 起きたかを述べ戻る導線を持つページに着地するようにするため。
func assertNotFoundPage(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q、期待値 = text/html; charset=utf-8", contentType)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>ページが見つかりません | Annict</title>",
		"ページが見つかりません",
		`href="/"`,
		"ホームに戻る",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("404レスポンスに%qが含まれていません", expected)
		}
	}
}

// TestIndexはDB作品一覧ページのテスト
func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// テストデータを作成
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ1").
		WithTitleKana("てすとあにめいち").
		WithTitleEn("Test Anime 1").
		WithMedia(1).
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ2").
		WithSeason(2024, testutil.SeasonSummer).
		Build()

	handler := newTestHandler(t, db, tx)

	// 操作列 (編集リンクを含む) はcommitterでゲートされるため、描画を検証できるよう
	// committerとしてリクエストを実行する。
	req := httptest.NewRequest("GET", "/db/works", nil)
	req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		// DBのページはブラウザのタブで公開画面と区別できるよう " | Annict DB" の
		// タイトルサフィックスを持つ。
		"<title>作品 | Annict DB</title>",
		"テストアニメ1",
		"テストアニメ2",
		"2024",
		"<table",
		"<thead",
		"<tbody",
		// ID列は作品の公開ページを新しいタブで開くリンクになる。
		fmt.Sprintf(`href="/works/%d"`, int64(workID)),
		`target="_blank"`,
		// IDリンクは新しいタブで開くことを知らせるaria-labelを持つ。
		fmt.Sprintf(`aria-label="作品 %d を新しいタブで開く"`, int64(workID)),
		// タイトルセルにふりがな・英語タイトルを、メディア列にメディア名を表示する。
		"てすとあにめいち",
		"Test Anime 1",
		"TV",
		// 各行がDBWorkEditPath経由で編集フォームへリンクする。
		fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID)),
		// 作品一覧は未ログインでも閲覧できるため、og:urlにそのページ自身の絶対URLが
		// 入り、URLを貼られたときにタイトル付きのカードが出る。
		`<meta property="og:url" content="https://test.annict.com/db/works">`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}

	// robots.txtが /db/ をDisallowしているため、管理画面はcanonicalを宣言しない。
	if strings.Contains(body, `rel="canonical"`) {
		t.Error("/dbの画面にcanonicalが含まれてはいけません")
	}

	expectedContentType := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("Content-Type = %v、期待値 = %v", ct, expectedContentType)
	}
}

// TestIndex_ActionColumnByRoleはハンドラーがリクエストコンテキストから閲覧者のロールを
// 解決し、操作列をend-to-endで出し分けることを検証する。adminは削除ボタンを、committer
// (編集者) は編集・非公開リンクを見るが削除ボタンは見えず、一般ユーザーや未ログインは操作
// コントロールを一切見ない。IndexPageDataを直接与えるtempl層テストを補い、
// GetUserFromContext -> IsCommitter / IsAdminのハンドラー配線を実際に通す。
func TestIndex_ActionColumnByRole(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// 公開中の作品: committerは編集・非公開リンクを、adminはさらに削除ボタンを得る。
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("操作列ロールテスト").WithMedia(1).Build()
	handler := newTestHandler(t, db, tx)

	editLink := fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID))
	unpublishLink := fmt.Sprintf(`href="/db/works/%d/archive/new"`, int64(workID))
	deleteButton := fmt.Sprintf(`hx-delete="/db/works/%d"`, int64(workID))

	tests := []struct {
		name        string
		user        *model.User
		wantPresent []string
		wantAbsent  []string
	}{
		{
			name:        "管理者: 編集・非公開・削除",
			user:        &model.User{ID: 1, Role: model.RoleAdmin},
			wantPresent: []string{editLink, unpublishLink, deleteButton},
		},
		{
			name:        "編集者: 編集・非公開、削除は不可",
			user:        &model.User{ID: 1, Role: model.RoleEditor},
			wantPresent: []string{editLink, unpublishLink},
			wantAbsent:  []string{deleteButton},
		},
		{
			name:       "一般ユーザー: 操作リンクなし",
			user:       &model.User{ID: 1, Role: model.RoleUser},
			wantAbsent: []string{editLink, unpublishLink, deleteButton},
		},
		{
			name:       "未ログイン: 操作リンクなし",
			user:       nil,
			wantAbsent: []string{editLink, unpublishLink, deleteButton},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/db/works", nil)
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()

			handler.Index(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusOK)
			}
			body := rr.Body.String()
			for _, want := range tt.wantPresent {
				if !strings.Contains(body, want) {
					t.Errorf("期待する文字列が含まれていません: %q", want)
				}
			}
			for _, notWant := range tt.wantAbsent {
				if strings.Contains(body, notWant) {
					t.Errorf("含まれてはいけない文字列が含まれています: %q", notWant)
				}
			}
		})
	}
}

// TestIndex_WithFiltersはフィルタパラメータ付きリクエストのテスト
func TestIndex_WithFilters(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	handler := newTestHandler(t, db, tx)

	req := httptest.NewRequest("GET", "/db/works?filter_no_episodes=1&filter_no_image=1&page=1", nil)
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	// フィルタのチェックボックスがチェックされていることを確認
	if !strings.Contains(body, `checked`) {
		t.Error("有効な絞り込みに対応するチェック済みのチェックボックスがレスポンスに含まれていない")
	}
}

// TestIndex_WithSeasonAndSlotFiltersはリリース時期の複数選択・放送予定未登録フィルタのテスト
func TestIndex_WithSeasonAndSlotFilters(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	spring2024 := testutil.NewWorkBuilder(t, tx).
		WithTitle("2024春アニメ").
		WithSeason(2024, testutil.SeasonSpring).
		Build()
	testutil.NewWorkBuilder(t, tx).
		WithTitle("2024夏アニメ").
		WithSeason(2024, testutil.SeasonSummer).
		Build()

	handler := newTestHandler(t, db, tx)

	// 一致した作品をcommitterでゲートされる編集リンクで確認するため、committerとして
	// リクエストを実行する。
	req := httptest.NewRequest("GET", "/db/works?season_slugs=2024-spring&filter_no_slots=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	// 選択したシーズンの作品のみが表示され、別シーズンは除外される。
	if !strings.Contains(body, "2024春アニメ") {
		t.Error("選択したシーズンに一致する作品が表示されるべき")
	}
	if strings.Contains(body, "2024夏アニメ") {
		t.Error("選択外のシーズンの作品は除外されるべき")
	}
	// 一致した作品の編集リンクで一致を厳密に確認する。
	if !strings.Contains(body, fmt.Sprintf(`href="/db/works/%d/edit"`, int64(spring2024))) {
		t.Error("一致した作品の編集リンクが表示されるべき")
	}
	// 選択したシーズンオプションがselected状態で再描画される。
	if !strings.Contains(body, `<option value="2024-spring" selected>`) {
		t.Error("選択したシーズンオプションがselectedで再描画されるべき")
	}
	if !strings.Contains(body, `<div role="option" data-value="2024-spring" aria-selected="true">`) {
		t.Error("選択したシーズンがcomboboxでもaria-selectedで再描画されるべき")
	}
	// 放送予定未登録チェックボックス自体がchecked状態で描画される。
	if !strings.Contains(body, `<input type="checkbox" name="filter_no_slots" value="1" checked>`) {
		t.Error("放送予定未登録フィルタのチェックボックスがcheckedで描画されるべき")
	}
}

// TestIndex_OGURLはog:urlが共有される表示をそのまま再現することを検証する。ページに
// 並ぶ作品を変えるパラメータ (フィルタとページ番号) はすべて含み、共有されたリンクがたまたま
// 持つ未知のパラメータは落とす。1ページ目はパラメータなしで宣言し、/db/worksと
// /db/works?page=1が1つの代表URLを共有する。
func TestIndex_OGURL(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	handler := newTestHandler(t, db, tx)

	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "ページ指定なしは代表URL",
			target: "/db/works",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works">`,
		},
		{
			name:   "1ページ目はパラメータなしへまとめる",
			target: "/db/works?page=1",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works">`,
		},
		{
			name:   "2ページ目以降はページ番号を含む",
			target: "/db/works?page=3",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works?page=3">`,
		},
		{
			name:   "フィルタとページ番号を併せて含む",
			target: "/db/works?filter_no_image=1&page=3",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works?filter_no_image=1&amp;page=3">`,
		},
		{
			name:   "リリース時期の複数選択をサーバー定義順へ揃える",
			target: "/db/works?season_slugs=2024-spring&season_slugs=2024-summer",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works?season_slugs=2024-summer&amp;season_slugs=2024-spring">`,
		},
		{
			name:   "不正・範囲外のリリース時期は含めない",
			target: "/db/works?season_slugs=not-a-season&season_slugs=1000-winter",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works">`,
		},
		{
			name:   "リリース時期の重複と入力順を正規化する",
			target: "/db/works?season_slugs=2024-spring&season_slugs=2024-summer&season_slugs=2024-spring",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works?season_slugs=2024-summer&amp;season_slugs=2024-spring">`,
		},
		{
			name:   "無効な値のフィルタは適用されないため含めない",
			target: "/db/works?filter_no_image=true",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works">`,
		},
		{
			name:   "未知のパラメータは落とす",
			target: "/db/works?utm_source=newsletter&page=2",
			want:   `<meta property="og:url" content="https://test.annict.com/db/works?page=2">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.target, nil)
			rr := httptest.NewRecorder()

			handler.Index(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
			}
			if body := rr.Body.String(); !strings.Contains(body, tt.want) {
				t.Errorf("レスポンスに含まれていない文字列 = %q", tt.want)
			}
		})
	}
}

// TestIndex_PaginationLinksはページネーションのリンクがog:urlと同じ代表パスを伸ばす
// ことを検証する。並ぶ作品を選ぶフィルタは引き継ぎ、共有されたリンクがたまたま持つ未知の
// パラメータはページを送っても付いて回らずに落ちる。
func TestIndex_PaginationLinks(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// 一覧は1ページにperPage件を並べるため、1件だけ超えるのがページネーションの
	// リンクを描画させる最小の前提データになる。
	for i := range int(perPage) + 1 {
		testutil.NewWorkBuilder(t, tx).
			WithTitle(fmt.Sprintf("ページ送り確認作品%d", i)).
			Build()
	}

	handler := newTestHandler(t, db, tx)

	req := httptest.NewRequest("GET", "/db/works?filter_no_episodes=1&utm_source=newsletter", nil)
	rr := httptest.NewRecorder()

	handler.Index(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	if want := `href="/db/works?filter_no_episodes=1&amp;page=2"`; !strings.Contains(body, want) {
		t.Errorf("レスポンスに含まれていない文字列 = %q", want)
	}
	if strings.Contains(body, "utm_source") {
		t.Error("ページネーションのリンクに未知のパラメータが残ってはいけません")
	}
}

// TestNewはDB作品新規作成フォームのテスト
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	handler := newTestHandler(t, db, tx)

	req := httptest.NewRequest("GET", "/db/works/new", nil)
	rr := httptest.NewRecorder()

	handler.New(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		// DBのページはブラウザのタブで公開画面と区別できるよう " | Annict DB" の
		// タイトルサフィックスを持つ。
		"<title>作品登録 | Annict DB</title>",
		"<form",
		`action="/db/works"`,
		`method="POST"`,
		"csrf_token",
		`name="title"`,
		`name="media"`,
		`name="season_year"`,
		`name="season_name"`,
		`name="number_format_id"`,
		`name="no_episodes"`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}

	expectedContentType := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("Content-Type = %v、期待値 = %v", ct, expectedContentType)
	}
}

// TestNew_RequiresCommitterは新規作成フォームのルートがcommitterロールで保護されて
// いることを検証する (committerは処理続行、一般ユーザーは403、未認証はサインインへ
// リダイレクト)。
func TestNew_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.RequireCommitter).Get("/db/works/new", handler.New)

	tests := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{
			name:       "未認証はサインインへリダイレクト",
			user:       nil,
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "一般ユーザーは403",
			user:       &model.User{ID: 1, Role: model.RoleUser},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "管理者はアクセス許可",
			user:       &model.User{ID: 1, Role: model.RoleAdmin},
			wantStatus: http.StatusOK,
		},
		{
			name:       "編集者はアクセス許可",
			user:       &model.User{ID: 1, Role: model.RoleEditor},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/db/works/new", nil)
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
