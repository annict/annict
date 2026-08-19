package db_episode_archive

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
	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// archiveTargetWorkSeasonYear keeps the works these tests commit out of the "no season" bucket
// the work list counts globally: the rows live in the shared test DB until the test ends and
// are visible to the sibling packages running at the same time.
//
// [Ja] archiveTargetWorkSeasonYear は、本テスト群がコミットする作品を、作品一覧が全体に対して
// 数える「シーズン未設定」の集合から外すためのもの。行はテストが終わるまで共有テスト DB に残り、
// 同時に走る他パッケージからも見えるため。
const archiveTargetWorkSeasonYear = 1904

// newTestHandler wires the handler against the shared test DB. The archive usecase opens its
// own transaction, so it is built on the pool rather than on the test transaction; the tests
// commit their fixtures and clean them up themselves.
//
// [Ja] newTestHandler は共有テスト DB に対してハンドラーを組み立てる。非公開 UseCase は自前の
// トランザクションを開くため、テスト用トランザクションではなくプールに対して組み立てる。
// フィクスチャは各テストがコミットし、後始末も行う。
func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) *Handler {
	t.Helper()

	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionManager := session.NewManager(repository.NewSessionRepository(query.New(db).WithTx(tx)), cfg)
	poolQueries := query.New(db)
	episodeRepo := repository.NewEpisodeRepository(poolQueries)

	return NewHandler(
		cfg,
		sessionManager,
		testutil.NewTestFlashManager(),
		usecase.NewGetDBEpisodeArchiveNewUsecase(episodeRepo),
		usecase.NewArchiveEpisodeUsecase(db, episodeRepo, repository.NewAnimeRepository(poolQueries)),
		usecase.NewUnarchiveEpisodeUsecase(db, episodeRepo, repository.NewAnimeRepository(poolQueries)),
	)
}

// insertArchiveTargetWork inserts the parent work of an archive test, committed to the shared
// pool rather than to the test transaction: the archive usecase opens its own transaction and
// would not see a work that is still uncommitted. Its cleanup also removes the episodes the
// test archived under it.
//
// [Ja] insertArchiveTargetWork は非公開テストの親作品を、テスト用トランザクションではなく共有
// プールにコミットして挿入する。非公開 UseCase は自前のトランザクションを開くため、未コミットの
// 作品は見えないからである。後始末では、そのテストが配下で非公開にしたエピソードも消す。
func insertArchiveTargetWork(t *testing.T, db *sql.DB) model.WorkID {
	t.Helper()

	var id int64
	if err := db.QueryRow(
		`INSERT INTO works (title, media, season_year, season_name, episodes_count, created_at, updated_at)
		 VALUES ($1, 1, $2, 1, 1, NOW(), NOW()) RETURNING id`,
		"非公開テストアニメ_"+t.Name(), archiveTargetWorkSeasonYear,
	).Scan(&id); err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}

	t.Cleanup(func() {
		for _, statement := range []string{
			`DELETE FROM episodes WHERE work_id = $1`,
			`DELETE FROM works WHERE id = $1`,
		} {
			if _, err := db.Exec(statement, id); err != nil {
				t.Logf("作品の後始末に失敗 (%s): %v", statement, err)
			}
		}
	})

	return model.WorkID(id)
}

// insertArchiveTargetEpisode inserts the episode an archive test unpublishes, committed to the
// shared pool for the reason insertArchiveTargetWork states. Its parent work's cleanup removes
// it.
//
// [Ja] insertArchiveTargetEpisode は非公開テストが非公開にするエピソードを、
// insertArchiveTargetWork が述べる理由により共有プールにコミットして挿入する。行の後始末は親作品
// の後始末が行う。
func insertArchiveTargetEpisode(t *testing.T, db *sql.DB, workID model.WorkID) model.EpisodeID {
	t.Helper()

	var id int64
	if err := db.QueryRow(`
		INSERT INTO episodes (work_id, number, sort_number, title, created_at, updated_at)
		VALUES ($1, '第2話', 200, 'もう、お婿にいけません', NOW(), NOW()) RETURNING id`,
		int64(workID),
	).Scan(&id); err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}

	return model.EpisodeID(id)
}

// archiveTargetEpisode puts an inserted episode into the state the re-publish endpoint starts
// from, which is also the state the archive endpoints refuse.
//
// [Ja] archiveTargetEpisode は挿入したエピソードを、再公開エンドポイントが起点とする状態にする。
// これは非公開エンドポイントが拒否する状態でもある。
func archiveTargetEpisode(t *testing.T, db *sql.DB, episodeID model.EpisodeID) {
	t.Helper()

	if _, err := db.Exec(`UPDATE episodes SET unpublished_at = NOW() WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("エピソードの非公開化に失敗: %v", err)
	}
}

// readEpisodeUnpublishedAt returns the state column the archive and the re-publish write, so a
// test can tell an episode either of them changed from one the submit left alone.
//
// [Ja] readEpisodeUnpublishedAt は非公開と再公開が書く状態カラムを返す。どちらかが変更した
// エピソードと、送信が手を触れなかったエピソードをテストが区別できるようにするため。
func readEpisodeUnpublishedAt(t *testing.T, db *sql.DB, episodeID model.EpisodeID) sql.NullTime {
	t.Helper()

	var unpublishedAt sql.NullTime
	if err := db.QueryRow(`SELECT unpublished_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&unpublishedAt); err != nil {
		t.Fatalf("エピソードの状態の読み込みに失敗: %v", err)
	}

	return unpublishedAt
}

func getRequest(target string) *http.Request {
	return httptest.NewRequest("GET", target, nil)
}

func postRequest(target string) *http.Request {
	req := httptest.NewRequest("POST", target, nil)
	editor := &model.User{ID: 1, Role: model.RoleEditor}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, editor))
}

func deleteRequest(target string) *http.Request {
	req := httptest.NewRequest("DELETE", target, nil)
	editor := &model.User{ID: 1, Role: model.RoleEditor}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, editor))
}

// TestNew verifies the confirmation page names the episode, keeps the work's subnav and posts
// to the episode's archive endpoint with a CSRF token.
//
// [Ja] TestNew は確認ページがエピソードを名指しし、作品のサブナビを保ち、CSRF トークン付きで
// エピソードの非公開エンドポイントへ POST することを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertArchiveTargetWork(t, db)
	episodeID := insertArchiveTargetEpisode(t, db, workID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/episodes/{id}/archive/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, getRequest(fmt.Sprintf("/db/episodes/%d/archive/new", int64(episodeID))))

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	body := rr.Body.String()
	expectedContents := []string{
		"<form",
		fmt.Sprintf(`action="/db/episodes/%d/archive"`, int64(episodeID)),
		`method="POST"`,
		"csrf_token",
		"第2話「もう、お婿にいけません」を非公開にしますか？",
		fmt.Sprintf(`href="/db/works/%d/episodes"`, int64(workID)),
		// The document title identifies the episode, so two confirmation pages open side by
		// side can be told apart in tabs, history and assistive technology.
		//
		// [Ja] 文書タイトルはエピソードを識別する。2 つの確認ページを並べて開いても、タブ・
		// 履歴・支援技術で区別できるようにするため。
		fmt.Sprintf("<title>第2話 (ID: %d) | 非公開テストアニメ_%s | エピソードを非公開にする | Annict DB</title>", int64(episodeID), t.Name()),
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
	}
}

// TestSetNewTitle verifies both work-name branches use the localized document-title template.
// Every combination keeps the episode identifier and the Annict DB suffix.
//
// [Ja] TestSetNewTitle は作品名の有無の両分岐がローカライズ済みの文書タイトルテンプレートを
// 使うことを検証する。すべての組み合わせでエピソード識別子と Annict DB のサフィックスを保つ。
func TestSetNewTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		locale            string
		episodeIdentifier string
		workName          string
		want              string
	}{
		{
			name:              "日本語・作品名あり",
			locale:            "ja",
			episodeIdentifier: "第2話 (ID: 42)",
			workName:          "テストアニメ",
			want:              "第2話 (ID: 42) | テストアニメ | エピソードを非公開にする | Annict DB",
		},
		{
			name:              "日本語・作品名なし",
			locale:            "ja",
			episodeIdentifier: "第2話 (ID: 42)",
			want:              "第2話 (ID: 42) | エピソードを非公開にする | Annict DB",
		},
		{
			name:              "英語・作品名あり",
			locale:            "en",
			episodeIdentifier: "Episode 2 (ID: 42)",
			workName:          "Test Anime",
			want:              "Episode 2 (ID: 42) | Test Anime | Archive Episode | Annict DB",
		},
		{
			name:              "英語・作品名なし",
			locale:            "en",
			episodeIdentifier: "Episode 2 (ID: 42)",
			want:              "Episode 2 (ID: 42) | Archive Episode | Annict DB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			var meta viewmodel.PageMeta
			setNewTitle(ctx, &meta, tt.episodeIdentifier, tt.workName)

			if meta.Title != tt.want {
				t.Errorf("meta.Title = %q, want %q", meta.Title, tt.want)
			}
		})
	}
}

// TestNew_OGURL verifies that og:url names the page's own GET path built from the parsed
// episode ID, so that a link spelling the ID with leading zeros still declares the one
// representative URL of that page.
//
// [Ja] TestNew_OGURL は og:url がパース済みのエピソード ID から組み立てたページ自身の GET パス
// になることを検証する。ID を先頭ゼロ付きで書いたリンクでも、そのページの代表 URL は 1 つに
// 揃う。
func TestNew_OGURL(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertArchiveTargetWork(t, db)
	episodeID := insertArchiveTargetEpisode(t, db, workID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/episodes/{id}/archive/new", handler.New)

	want := fmt.Sprintf(`<meta property="og:url" content="https://test.annict.com/db/episodes/%d/archive/new">`, int64(episodeID))

	for _, target := range []string{
		fmt.Sprintf("/db/episodes/%d/archive/new", int64(episodeID)),
		fmt.Sprintf("/db/episodes/000%d/archive/new", int64(episodeID)),
	} {
		t.Run(target, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, getRequest(target))

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}
			if body := rr.Body.String(); !strings.Contains(body, want) {
				t.Errorf("レスポンスに %q が含まれていません", want)
			}
		})
	}
}

// TestNew_NotFound verifies the confirmation page is refused for an episode that is not
// currently published and for a malformed id.
//
// [Ja] TestNew_NotFound は、現在公開中でないエピソードと不正な id に対して確認ページが拒否
// されることを検証する。
func TestNew_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertArchiveTargetWork(t, db)
	archivedID := insertArchiveTargetEpisode(t, db, workID)
	archiveTargetEpisode(t, db, archivedID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Get("/db/episodes/{id}/archive/new", handler.New)

	targets := map[string]string{
		"非公開済みのエピソード": fmt.Sprintf("/db/episodes/%d/archive/new", int64(archivedID)),
		"存在しないエピソード":  "/db/episodes/999999999/archive/new",
		"数値でない id":    "/db/episodes/abc/archive/new",
	}
	for name, target := range targets {
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, getRequest(target))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("status = %d, want %d", status, http.StatusNotFound)
			}
		})
	}
}

// TestCreate_Success verifies a submit archives the episode and lands on the work's episode
// list, where the editor sees the archived row among the others.
//
// [Ja] TestCreate_Success は送信がエピソードを非公開にし、作品のエピソード一覧に着地すること
// を検証する。編集者はそこで他の行と並んだ非公開後の行を見る。
func TestCreate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertArchiveTargetWork(t, db)
	episodeID := insertArchiveTargetEpisode(t, db, workID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/episodes/{id}/archive", handler.Create)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, postRequest(fmt.Sprintf("/db/episodes/%d/archive", int64(episodeID))))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", status, http.StatusSeeOther)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != want {
		t.Errorf("Location = %q, want %q", location, want)
	}
	if unpublishedAt := readEpisodeUnpublishedAt(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL, want 非公開の時刻")
	}
}

// TestCreate_NotFound verifies a submit for an episode that cannot be archived returns 404
// rather than reporting a write that did not happen.
//
// [Ja] TestCreate_NotFound は、非公開にできないエピソードへの送信が、起きなかった書き込みを
// 報告せず 404 を返すことを検証する。
func TestCreate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/episodes/{id}/archive", handler.Create)

	for name, target := range map[string]string{
		"存在しないエピソード": "/db/episodes/999999999/archive",
		"数値でない id":   "/db/episodes/abc/archive",
	} {
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, postRequest(target))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("status = %d, want %d", status, http.StatusNotFound)
			}
		})
	}
}

// TestCreate_Forbidden verifies a direct Handler invocation maps the UseCase authorization
// failure to 403 instead of exposing it as an internal error.
//
// [Ja] TestCreate_Forbidden は Handler の直接呼び出しで UseCase の認可失敗を内部エラーとして
// 公開せず、403 に変換することを検証する。
func TestCreate_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/episodes/{id}/archive", handler.Create)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("POST", "/db/episodes/1/archive", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("status = %d, want %d", status, http.StatusForbidden)
	}
}

// TestDelete_Success verifies a re-publish submit clears the archived state and lands on the
// work's episode list, where the editor sees the published row among the others.
//
// [Ja] TestDelete_Success は再公開の送信が非公開の状態を解除し、作品のエピソード一覧に着地する
// ことを検証する。編集者はそこで他の行と並んだ公開後の行を見る。
func TestDelete_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertArchiveTargetWork(t, db)
	episodeID := insertArchiveTargetEpisode(t, db, workID)
	archiveTargetEpisode(t, db, episodeID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}/archive", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, deleteRequest(fmt.Sprintf("/db/episodes/%d/archive", int64(episodeID))))

	if status := rr.Code; status != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", status, http.StatusSeeOther)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != want {
		t.Errorf("Location = %q, want %q", location, want)
	}
	if unpublishedAt := readEpisodeUnpublishedAt(t, db, episodeID); unpublishedAt.Valid {
		t.Errorf("episodes.unpublished_at = %v, want NULL", unpublishedAt.Time)
	}
}

// TestDelete_HTMX verifies the submit the episode list's publish button makes gets HX-Redirect
// instead of a redirect htmx would follow and swap into the button.
//
// [Ja] TestDelete_HTMX は、エピソード一覧の公開ボタンが行う送信に対し、htmx が追ってボタンに
// スワップしてしまうリダイレクトではなく HX-Redirect を返すことを検証する。
func TestDelete_HTMX(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertArchiveTargetWork(t, db)
	episodeID := insertArchiveTargetEpisode(t, db, workID)
	archiveTargetEpisode(t, db, episodeID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}/archive", handler.Delete)

	req := deleteRequest(fmt.Sprintf("/db/episodes/%d/archive", int64(episodeID)))
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", status, http.StatusNoContent)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if redirect := rr.Header().Get("HX-Redirect"); redirect != want {
		t.Errorf("HX-Redirect = %q, want %q", redirect, want)
	}
	if location := rr.Header().Get("Location"); location != "" {
		t.Errorf("Location = %q, want 空 (htmx は HX-Redirect で遷移する)", location)
	}
	if unpublishedAt := readEpisodeUnpublishedAt(t, db, episodeID); unpublishedAt.Valid {
		t.Errorf("episodes.unpublished_at = %v, want NULL", unpublishedAt.Time)
	}
}

// TestDelete_NotFound verifies a submit for an episode that cannot be re-published returns 404
// rather than reporting a write that did not happen.
//
// [Ja] TestDelete_NotFound は、再公開できないエピソードへの送信が、起きなかった書き込みを報告
// せず 404 を返すことを検証する。
func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertArchiveTargetWork(t, db)
	publishedID := insertArchiveTargetEpisode(t, db, workID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}/archive", handler.Delete)

	for name, target := range map[string]string{
		"公開中のエピソード":  fmt.Sprintf("/db/episodes/%d/archive", int64(publishedID)),
		"存在しないエピソード": "/db/episodes/999999999/archive",
		"数値でない id":   "/db/episodes/abc/archive",
	} {
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, deleteRequest(target))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("status = %d, want %d", status, http.StatusNotFound)
			}
		})
	}
}

// TestDelete_Forbidden verifies a direct Handler invocation maps the UseCase authorization
// failure to 403 instead of exposing it as an internal error.
//
// [Ja] TestDelete_Forbidden は Handler の直接呼び出しで UseCase の認可失敗を内部エラーとして
// 公開せず、403 に変換することを検証する。
func TestDelete_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}/archive", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("DELETE", "/db/episodes/1/archive", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("status = %d, want %d", status, http.StatusForbidden)
	}
}

// TestRequiresCommitter verifies every endpoint is gated by the committer role at the HTTP
// boundary. The archive and re-publish write usecases repeat the role check for non-HTTP entry
// points.
//
// [Ja] TestRequiresCommitter は HTTP 境界ですべてのエンドポイントが committer ロールによりゲート
// されていることを検証する。非公開・再公開の書き込み UseCase は HTTP 以外の entry point に対して
// もロール検査を繰り返す。
func TestRequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	workID := insertArchiveTargetWork(t, db)
	episodeID := insertArchiveTargetEpisode(t, db, workID)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireCommitter)
		r.Get("/db/episodes/{id}/archive/new", handler.New)
		r.Post("/db/episodes/{id}/archive", handler.Create)
		r.Delete("/db/episodes/{id}/archive", handler.Delete)
	})

	routes := map[string]struct {
		method string
		target string
	}{
		"確認ページ":  {method: "GET", target: fmt.Sprintf("/db/episodes/%d/archive/new", int64(episodeID))},
		"非公開の送信": {method: "POST", target: fmt.Sprintf("/db/episodes/%d/archive", int64(episodeID))},
		"再公開の送信": {method: "DELETE", target: fmt.Sprintf("/db/episodes/%d/archive", int64(episodeID))},
	}
	users := []struct {
		name       string
		user       *model.User
		wantStatus int
	}{
		{name: "未認証はサインインへリダイレクト", user: nil, wantStatus: http.StatusSeeOther},
		{name: "一般ユーザーは403", user: &model.User{ID: 1, Role: model.RoleUser}, wantStatus: http.StatusForbidden},
	}

	for routeName, route := range routes {
		for _, tt := range users {
			t.Run(routeName+"/"+tt.name, func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.target, nil)
				if tt.user != nil {
					req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.user))
				}
				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, req)

				if rr.Code != tt.wantStatus {
					t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
				}
			})
		}
	}

	// A committer passes the middleware and must not be rejected. The confirmation page is
	// used rather than the submit so the assertion does not depend on the episode still being
	// archivable after another subtest has archived it.
	//
	// [Ja] committer はミドルウェアを通過し、弾かれてはならない。送信ではなく確認ページで確かめ
	// るのは、他のサブテストが非公開にした後でもエピソードが非公開にできる状態かどうかに、検証が
	// 左右されないようにするため。
	t.Run("編集者は通過", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/db/episodes/%d/archive/new", int64(episodeID)), nil)
		req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}
	})
}
