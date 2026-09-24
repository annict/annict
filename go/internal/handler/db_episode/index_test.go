package db_episode

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
	"github.com/annict/annict/go/internal/viewmodel"
)

func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionManager := session.NewManager(repository.NewSessionRepository(queries), cfg)
	workRepo := repository.NewWorkRepository(queries)
	episodeRepo := repository.NewEpisodeRepository(queries)

	// 作成・更新・削除UseCaseは自前のトランザクションを開くため、テスト用トランザクション
	// ではなくプールに対して組み立てる。コミットされた行はそのテスト側で後始末する。
	createEpisodesUC := usecase.NewCreateEpisodesUsecase(
		db,
		repository.NewWorkRepository(query.New(db)),
		repository.NewEpisodeRepository(query.New(db)),
		repository.NewAnimeRepository(query.New(db)),
		repository.NewAnimeClassificationRepository(query.New(db)),
		validator.NewDBEpisodeCreateValidator(),
	)
	updateEpisodeUC := usecase.NewUpdateEpisodeUsecase(
		db,
		repository.NewEpisodeRepository(query.New(db)),
		repository.NewAnimeRepository(query.New(db)),
		repository.NewAnimeClassificationRepository(query.New(db)),
		validator.NewDBEpisodeUpdateValidator(),
	)
	deleteEpisodeUC := usecase.NewDeleteEpisodeUsecase(
		db,
		repository.NewEpisodeRepository(query.New(db)),
		repository.NewAnimeRepository(query.New(db)),
	)

	return NewHandler(
		cfg,
		sessionManager,
		testutil.NewTestFlashManager(),
		usecase.NewGetDBEpisodesUsecase(workRepo, episodeRepo),
		usecase.NewGetDBEpisodeNewUsecase(workRepo),
		createEpisodesUC,
		usecase.NewGetDBEpisodeEditUsecase(episodeRepo),
		updateEpisodeUC,
		deleteEpisodeUC,
	)
}

// newIndexRequestはある作品のエピソード一覧へのGETリクエストを、chiがルートパターン
// から取り出すwork_idのURLパラメータ付きで組み立てる。
func newIndexRequest(workID model.WorkID, query string) *http.Request {
	path := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	req := httptest.NewRequest("GET", path+query, nil)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("work_id", fmt.Sprintf("%d", int64(workID)))

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("第1話").
		WithTitle("はじまり").
		Build()

	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Index(rr, newIndexRequest(workID, ""))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		// DBのページはブラウザのタブで公開画面と区別できるよう " | Annict DB" の
		// タイトルサフィックスを持つ。
		"<title>エピソード | テストアニメ | Annict DB</title>",
		// 見出しは親作品を名指しし、サブナビはそのフォームへ戻るリンクを持つ。
		"テストアニメ",
		fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID)),
		"<table",
		"<thead",
		"<tbody",
		"第1話",
		"はじまり",
		// ID列はエピソードの公開ページを新しいタブで開くリンクになる。
		fmt.Sprintf(`href="/works/%d/episodes/%d"`, int64(workID), int64(episodeID)),
		`target="_blank"`,
		fmt.Sprintf(`aria-label="エピソード %d を新しいタブで開く"`, int64(episodeID)),
		// 公開中のエピソードには公開のバッジが出る。
		`<span class="badge" data-variant="success">公開</span>`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}
}

// TestIndex_NewEpisodesLinkIsCommitterOnlyは一括作成フォームへの導線を検証する。一覧は
// 公開のため、リンクはcommitterにだけ出し、それ以外には出さない。403が返るだけのリンクを
// 出さないようにするため。
func TestIndex_NewEpisodesLinkIsCommitterOnly(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()
	handler := newTestHandler(t, db, tx)

	newLink := fmt.Sprintf(`href="/db/works/%d/episodes/new"`, int64(workID))
	// actionsContainerは見出しが操作の周りに描画するラッパー。操作の無い閲覧者にはこれも
	// 出さない。空のラッパーはモバイル幅では単独で全幅のflex行になり、公開されている一覧の
	// 見出しの下に余白を足してしまうため。
	actionsContainer := `<div class="flex w-full flex-none justify-end gap-2 md:w-auto">`

	tests := []struct {
		name     string
		user     *model.User
		wantLink bool
	}{
		{name: "未ログインには出さない", user: nil, wantLink: false},
		{name: "一般ユーザーには出さない", user: &model.User{ID: 1, Role: model.RoleUser}, wantLink: false},
		{name: "編集者には出す", user: &model.User{ID: 1, Role: model.RoleEditor}, wantLink: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newIndexRequest(workID, "")
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
			}
			rr := httptest.NewRecorder()
			handler.Index(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
			}
			body := rr.Body.String()
			if got := strings.Contains(body, newLink); got != tt.wantLink {
				t.Errorf("一括作成フォームへのリンクの有無 = %v、期待値 = %v", got, tt.wantLink)
			}
			if got := strings.Contains(body, actionsContainer); got != tt.wantLink {
				t.Errorf("見出しの操作コンテナの有無 = %v、期待値 = %v", got, tt.wantLink)
			}
		})
	}
}

// TestIndex_ActionColumnFollowsViewerRoleは行ごとの操作列を通しで検証する。一覧は公開の
// ため、閲覧者のロールがどのコントロールを描画するかを決める。編集・非公開・公開はcommitter、
// 削除はadminの操作で、各エンドポイントを守るmiddlewareと揃う。これにより403が返るだけの
// コントロールを出さない。
func TestIndex_ActionColumnFollowsViewerRole(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()
	publishedID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第1話").Build()
	archivedID := testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("第2話").
		WithUnpublishedAt(time.Now()).
		Build()

	handler := newTestHandler(t, db, tx)
	sessionManager := session.NewManager(
		repository.NewSessionRepository(query.New(db).WithTx(tx)),
		&config.Config{Env: "test", Domain: "test.annict.com"},
	)

	editLink := fmt.Sprintf(`href="/db/episodes/%d/edit"`, int64(publishedID))
	archiveLink := fmt.Sprintf(`href="/db/episodes/%d/archive/new"`, int64(publishedID))
	publishButton := fmt.Sprintf(`hx-delete="/db/episodes/%d/archive"`, int64(archivedID))
	deleteButton := fmt.Sprintf(`hx-delete="/db/episodes/%d"`, int64(publishedID))

	tests := []struct {
		name              string
		user              *model.User
		wantCommitterOnly bool
		wantAdminOnly     bool
	}{
		{name: "未ログイン", user: nil},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}, wantCommitterOnly: true},
		{name: "管理者", user: &model.User{ID: 1, Role: model.RoleAdmin}, wantCommitterOnly: true, wantAdminOnly: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newIndexRequest(workID, "")
			var csrfToken string
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))

				// コンテキスト内のロールが表示を決め、永続化したログインセッションが描画後の
				// htmxリクエスト用CSRFトークンを供給する。実リクエストの両方の配線をこの
				// ハンドラーレベルのテストで通す。
				sessionRecorder := httptest.NewRecorder()
				if err := sessionManager.CreateSession(req.Context(), sessionRecorder, req, tt.user.ID); err != nil {
					t.Fatalf("セッションの作成に失敗しました: %v", err)
				}

				var sessionCookie *http.Cookie
				for _, cookie := range sessionRecorder.Result().Cookies() {
					if cookie.Name == session.SessionKey {
						sessionCookie = cookie
						break
					}
				}
				if sessionCookie == nil {
					t.Fatal("セッションCookieが設定されていません")
				}
				req.AddCookie(sessionCookie)

				csrfToken = authMiddleware.GetCSRFToken(req, sessionManager)
				if csrfToken == "" {
					t.Fatal("セッションから空でないCSRFトークンを取得できませんでした")
				}
			}
			rr := httptest.NewRecorder()
			handler.Index(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
			}
			body := rr.Body.String()

			for _, control := range []string{editLink, archiveLink, publishButton} {
				if got := strings.Contains(body, control); got != tt.wantCommitterOnly {
					t.Errorf("%qの有無 = %v、期待値 = %v", control, got, tt.wantCommitterOnly)
				}
			}
			if got := strings.Contains(body, deleteButton); got != tt.wantAdminOnly {
				t.Errorf("%qの有無 = %v、期待値 = %v", deleteButton, got, tt.wantAdminOnly)
			}
			// 列の見出しはコントロールと同じ条件で出る。操作の無い閲覧者には、空の
			// セルが並ぶ列ではなく列そのものを出さない。
			wantHeader := tt.wantCommitterOnly || tt.wantAdminOnly
			if got := strings.Contains(body, `<th scope="col" class="text-center">操作</th>`); got != wantHeader {
				t.Errorf("操作列の見出しの有無 = %v、期待値 = %v", got, wantHeader)
			}
			if wantHeader {
				wantCSRFHeader := fmt.Sprintf(`hx-headers="{&#34;X-CSRF-Token&#34;:&#34;%s&#34;}"`, csrfToken)
				if !strings.Contains(body, wantCSRFHeader) {
					t.Errorf("操作ボタンにセッションのCSRFトークンが含まれていません: %q", wantCSRFHeader)
				}
			}
		})
	}
}

// TestIndex_ShowsGenerationNoticeAndDerivedColumnsは、エピソードそのもの以外にページが
// 運ぶ情報を検証する。編集者が計画に使う自動生成の案内と、一覧が行ごとに導出する2つの列
// (直前のエピソードと記録数)。
func TestIndex_ShowsGenerationNoticeAndDerivedColumns(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("テストアニメ").
		WithManualEpisodesCount(12).
		Build()
	insertEpisodeForIndex(t, tx, workID, "第1話", 100, 0)
	insertEpisodeForIndex(t, tx, workID, "第2話", 200, 42)

	channelID := testutil.NewChannelBuilder(t, tx).Build()
	testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).WithNumber(9).Build()

	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Index(rr, newIndexRequest(workID, ""))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		// 案内は作品の予定エピソード数・公開中のエピソード数・自動生成されるエピソード数を報告する。
		"<dt>予定エピソード数</dt>",
		`<dd class="text-card-foreground">12</dd>`,
		"<dt>公開中のエピソード数</dt>",
		`<dd class="text-card-foreground">2</dd>`,
		"<dt>自動生成されるエピソード数</dt>",
		`<dd class="text-card-foreground">9</dd>`,
		// 第2話は直前のエピソードとして第1話を名指しし、記録数を持つ。
		"前のエピソード",
		`<td class="whitespace-normal [overflow-wrap:anywhere]">第1話</td>`,
		"<td>42</td>",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}
}

// insertEpisodeForIndexはsort_numberと記録数を明示してエピソードを作成する。共有の
// EpisodeBuilderはどちらも固定しており、一覧は直前のエピソードをsort_number順から導出する
// ため、順序を問うこのテストの行は直接挿入する。
func insertEpisodeForIndex(t *testing.T, tx *sql.Tx, workID model.WorkID, number string, sortNumber, episodeRecordsCount int32) {
	t.Helper()

	if _, err := tx.Exec(`
		INSERT INTO episodes (work_id, number, sort_number, episode_records_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, int64(workID), number, sortNumber, episodeRecordsCount); err != nil {
		t.Fatalf("エピソードの作成に失敗: %v", err)
	}
}

func TestIndex_ExcludesDeletedEpisodes(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()
	testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("第1話").
		WithTitle("非公開の話").
		WithUnpublishedAt(time.Now()).
		Build()
	testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("第2話").
		WithTitle("削除済みの話").
		WithDeletedAt(time.Now()).
		Build()

	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Index(rr, newIndexRequest(workID, ""))

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	if !strings.Contains(body, "非公開の話") {
		t.Error("非公開のエピソードは一覧に表示されるべきです")
	}
	if !strings.Contains(body, `<span class="badge" data-variant="warning">非公開</span>`) {
		t.Error("非公開のエピソードには非公開のバッジが表示されるべきです")
	}
	if strings.Contains(body, "削除済みの話") {
		t.Error("削除済みのエピソードは一覧に表示されてはいけません")
	}
}

func TestIndex_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	deletedWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("削除済みアニメ").Build()
	if _, err := tx.Exec("UPDATE works SET deleted_at = NOW() WHERE id = $1", int64(deletedWorkID)); err != nil {
		t.Fatalf("作品の削除に失敗: %v", err)
	}

	handler := newTestHandler(t, db, tx)

	t.Run("存在しない作品", func(t *testing.T) {
		rr := httptest.NewRecorder()
		handler.Index(rr, newIndexRequest(model.WorkID(999999999), ""))

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})

	t.Run("削除済みの作品", func(t *testing.T) {
		rr := httptest.NewRecorder()
		handler.Index(rr, newIndexRequest(deletedWorkID, ""))

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})

	t.Run("数値でないwork_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/db/works/abc/episodes", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("work_id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

		rr := httptest.NewRecorder()
		handler.Index(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("ステータスコード = %v、期待値 = %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})
}

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

func TestSetIndexTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		locale    string
		workTitle string
		page      int32
		want      string
	}{
		{
			name:      "日本語の1ページ目",
			locale:    "ja",
			workTitle: "テストアニメ",
			page:      1,
			want:      "エピソード | テストアニメ | Annict DB",
		},
		{
			name:      "日本語の2ページ目",
			locale:    "ja",
			workTitle: "テストアニメ",
			page:      2,
			want:      "エピソード (2ページ目) | テストアニメ | Annict DB",
		},
		{
			name:      "英語の2ページ目",
			locale:    "en",
			workTitle: "Test Anime",
			page:      2,
			want:      "Episodes (Page 2) | Test Anime | Annict DB",
		},
		{
			name:      "空の作品タイトル",
			locale:    "en",
			workTitle: "   ",
			page:      2,
			want:      "Episodes (Page 2) | Annict DB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var meta viewmodel.PageMeta
			setIndexTitle(ctx, &meta, viewmodel.DBEpisodeListWorkName(tt.workTitle), tt.page)

			if meta.Title != tt.want {
				t.Errorf("meta.Title = %q、期待値 = %q", meta.Title, tt.want)
			}
		})
	}
}

// canonicalTagは、ある作品のエピソード一覧が指定のクエリで持つog:urlタグを組み立てる。
// DBのページは代表URLをog:urlとして持つ (robots.txtでクロールを禁止しているためDBHead
// はcanonicalのリンクを出さない)。そのためこのタグを検証する。
func canonicalTag(workID model.WorkID, query string) string {
	return fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/works/%d/episodes%s">`, int64(workID), query)
}

func TestIndex_Pagination(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()
	testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第1話").Build()

	handler := newTestHandler(t, db, tx)

	tests := []struct {
		name          string
		query         string
		wantCanonical string
		wantTitle     string
	}{
		// 1ページ目と ?page=1の形は1つの代表URLを共有する。
		{name: "ページ指定なし", query: "", wantCanonical: canonicalTag(workID, ""), wantTitle: "エピソード | テストアニメ | Annict DB"},
		{name: "1ページ目の明示指定", query: "?page=1", wantCanonical: canonicalTag(workID, ""), wantTitle: "エピソード | テストアニメ | Annict DB"},
		// 正の整数でないページ番号は1ページ目にフォールバックする。
		{name: "不正なページ番号", query: "?page=abc", wantCanonical: canonicalTag(workID, ""), wantTitle: "エピソード | テストアニメ | Annict DB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.Index(rr, newIndexRequest(workID, tt.query))

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
			}

			if !strings.Contains(rr.Body.String(), tt.wantCanonical) {
				t.Errorf("canonical URLに%qが含まれていません", tt.wantCanonical)
			}

			if !strings.Contains(rr.Body.String(), "<title>"+tt.wantTitle+"</title>") {
				t.Errorf("titleに%qが含まれていません", tt.wantTitle)
			}
		})
	}
}

// TestIndex_SecondPageは作品が実際に持つページ番号を検証する。代表URLと文書タイトルが
// そのページを名乗り、そのページに属する行だけが並ぶ。
func TestIndex_SecondPage(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()

	// 1ページ100件の境界をちょうど1件超える件数を入れ、2ページ目に並び順が最も
	// 小さい1件だけが載るようにする。行は1文で投入する。ビルダーは1回の呼び出しにつき
	// 1行を挿入するうえ、ここでの一覧の並びはビルダーが固定するsort_numberに依存するため。
	if _, err := tx.Exec(`
		INSERT INTO episodes (work_id, number, sort_number, created_at, updated_at)
		SELECT $1, '第' || i || '話', i, NOW(), NOW()
		FROM generate_series(1, 101) AS i
	`, int64(workID)); err != nil {
		t.Fatalf("エピソードの作成に失敗: %v", err)
	}

	handler := newTestHandler(t, db, tx)

	rr := httptest.NewRecorder()
	handler.Index(rr, newIndexRequest(workID, "?page=2"))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	for _, expected := range []string{
		canonicalTag(workID, "?page=2"),
		"<title>エピソード (2ページ目) | テストアニメ | Annict DB</title>",
		// sort_numberは話数とともに増え、一覧は降順に並ぶため、2ページ目には第1話が
		// 1件だけ載る。
		"第1話",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	if strings.Contains(body, "第101話") {
		t.Error("1ページ目のエピソードが2ページ目に含まれています")
	}
}

// TestIndex_PageBeyondLastRedirectsは一覧の末尾を超えるページ番号を検証する。そのまま
// 描画すると、作品が持たないページ番号を名乗る空の一覧に着地し、テーブルと一緒に
// ページネーションも落ちてしまうため、代わりに最後のページへ送る。
func TestIndex_PageBeyondLastRedirects(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	withEpisodeID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソードあり").Build()
	testutil.NewEpisodeBuilder(t, tx, withEpisodeID).WithNumber("第1話").Build()

	noEpisodeID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソードなし").Build()

	handler := newTestHandler(t, db, tx)

	tests := []struct {
		name         string
		workID       model.WorkID
		query        string
		wantLocation string
	}{
		{
			name:         "エピソード1件の作品の2ページ目",
			workID:       withEpisodeID,
			query:        "?page=2",
			wantLocation: fmt.Sprintf("/db/works/%d/episodes", int64(withEpisodeID)),
		},
		{
			name:         "最大ページ番号",
			workID:       withEpisodeID,
			query:        fmt.Sprintf("?page=%d", math.MaxInt32),
			wantLocation: fmt.Sprintf("/db/works/%d/episodes", int64(withEpisodeID)),
		},
		// エピソードが無い作品も1ページ目は持つため、空の一覧はページ番号なしのパスで
		// 見られる。
		{
			name:         "エピソードが無い作品",
			workID:       noEpisodeID,
			query:        "?page=5",
			wantLocation: fmt.Sprintf("/db/works/%d/episodes", int64(noEpisodeID)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.Index(rr, newIndexRequest(tt.workID, tt.query))

			if status := rr.Code; status != http.StatusFound {
				t.Fatalf("ステータスコード = %v、期待値 = %v", status, http.StatusFound)
			}
			if location := rr.Header().Get("Location"); location != tt.wantLocation {
				t.Errorf("リダイレクト先 = %q、期待値 = %q", location, tt.wantLocation)
			}
		})
	}
}

func TestLastPageNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		totalCount int64
		perPage    int32
		want       int64
	}{
		{name: "0件でも1ページ目は存在する", totalCount: 0, perPage: 100, want: 1},
		{name: "1件", totalCount: 1, perPage: 100, want: 1},
		{name: "ちょうど1ページ分", totalCount: 100, perPage: 100, want: 1},
		{name: "1ページ分を1件超える", totalCount: 101, perPage: 100, want: 2},
		{name: "ちょうど2ページ分", totalCount: 200, perPage: 100, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := lastPageNumber(tt.totalCount, tt.perPage); got != tt.want {
				t.Errorf("lastPageNumber(%d, %d) = %d、期待値 = %d", tt.totalCount, tt.perPage, got, tt.want)
			}
		})
	}
}
