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

// archiveTargetWorkSeasonYearは、本テスト群がコミットする作品を、作品一覧が全体に対して
// 数える「シーズン未設定」の集合から外すためのもの。行はテストが終わるまで共有テストDBに残り、
// 同時に走る他パッケージからも見えるため。
const archiveTargetWorkSeasonYear = 1904

// newTestHandlerは共有テストDBに対してハンドラーを組み立てる。非公開UseCaseは自前の
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

// insertArchiveTargetWorkは非公開テストの親作品を、テスト用トランザクションではなく共有
// プールにコミットして挿入する。非公開UseCaseは自前のトランザクションを開くため、未コミットの
// 作品は見えないからである。後始末では、そのテストが配下で非公開にしたエピソードも消す。
func insertArchiveTargetWork(t *testing.T, db *sql.DB) model.WorkID {
	t.Helper()

	var id int64
	if err := db.QueryRow(
		`INSERT INTO works (title, media, season_year, season_name, episodes_count, created_at, updated_at)
		 VALUES ($1, 1, $2, 1, 1, NOW(), NOW()) RETURNING id`,
		"非公開テストアニメ_"+t.Name(), archiveTargetWorkSeasonYear,
	).Scan(&id); err != nil {
		t.Fatalf("worksの挿入に失敗: %v", err)
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

// insertArchiveTargetEpisodeは非公開テストが非公開にするエピソードを、
// insertArchiveTargetWorkが述べる理由により共有プールにコミットして挿入する。行の後始末は親作品
// の後始末が行う。
func insertArchiveTargetEpisode(t *testing.T, db *sql.DB, workID model.WorkID) model.EpisodeID {
	t.Helper()

	var id int64
	if err := db.QueryRow(`
		INSERT INTO episodes (work_id, number, sort_number, title, created_at, updated_at)
		VALUES ($1, '第2話', 200, 'もう、お婿にいけません', NOW(), NOW()) RETURNING id`,
		int64(workID),
	).Scan(&id); err != nil {
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}

	return model.EpisodeID(id)
}

// archiveTargetEpisodeは挿入したエピソードを、再公開エンドポイントが起点とする状態にする。
// これは非公開エンドポイントが拒否する状態でもある。
func archiveTargetEpisode(t *testing.T, db *sql.DB, episodeID model.EpisodeID) {
	t.Helper()

	if _, err := db.Exec(`UPDATE episodes SET unpublished_at = NOW() WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("エピソードの非公開化に失敗: %v", err)
	}
}

// readEpisodeUnpublishedAtは非公開と再公開が書く状態カラムを返す。どちらかが変更した
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

// TestNewは確認ページがエピソードを名指しし、作品のサブナビを保ち、CSRFトークン付きで
// エピソードの非公開エンドポイントへPOSTすることを検証する。
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusOK)
	}

	body := rr.Body.String()
	expectedContents := []string{
		"<form",
		fmt.Sprintf(`action="/db/episodes/%d/archive"`, int64(episodeID)),
		`method="POST"`,
		"csrf_token",
		"第2話「もう、お婿にいけません」を非公開にしますか？",
		fmt.Sprintf(`href="/db/works/%d/episodes"`, int64(workID)),
		// 文書タイトルはエピソードのラベルを含む。ラベルが異なる確認ページを並べて
		// 開いたときに、タブ・履歴・支援技術で区別できるようにするため。
		fmt.Sprintf("<title>エピソード非公開 | 第2話 | 非公開テストアニメ_%s | Annict DB</title>", t.Name()),
	}
	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q、期待値 = %q", ct, "text/html; charset=utf-8")
	}
}

// TestSetNewTitleは作品名の有無の両分岐がローカライズ済みの文書タイトルテンプレートを
// 使うことを検証する。すべての組み合わせでエピソード識別子とAnnict DBのサフィックスを保つ。
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
			episodeIdentifier: "第2話",
			workName:          "テストアニメ",
			want:              "エピソード非公開 | 第2話 | テストアニメ | Annict DB",
		},
		{
			name:              "日本語・作品名なし",
			locale:            "ja",
			episodeIdentifier: "第2話",
			want:              "エピソード非公開 | 第2話 | Annict DB",
		},
		{
			name:              "英語・作品名あり",
			locale:            "en",
			episodeIdentifier: "Episode 2",
			workName:          "Test Anime",
			want:              "Archive Episode | Episode 2 | Test Anime | Annict DB",
		},
		{
			name:              "英語・作品名なし",
			locale:            "en",
			episodeIdentifier: "Episode 2",
			want:              "Archive Episode | Episode 2 | Annict DB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			var meta viewmodel.PageMeta
			setNewTitle(ctx, &meta, tt.episodeIdentifier, tt.workName)

			if meta.Title != tt.want {
				t.Errorf("meta.Title = %q、期待値 = %q", meta.Title, tt.want)
			}
		})
	}
}

// TestNew_OGURLはog:urlがパース済みのエピソードIDから組み立てたページ自身のGETパス
// になることを検証する。IDを先頭ゼロ付きで書いたリンクでも、そのページの代表URLは1つに
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
				t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusOK)
			}
			if body := rr.Body.String(); !strings.Contains(body, want) {
				t.Errorf("レスポンスに%qが含まれていません", want)
			}
		})
	}
}

// TestNew_NotFoundは、現在公開中でないエピソードと不正なidに対して確認ページが拒否
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
		"数値でないid":     "/db/episodes/abc/archive/new",
	}
	for name, target := range targets {
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, getRequest(target))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("ステータスコード = %d、期待値 = %d", status, http.StatusNotFound)
			}
		})
	}
}

// TestCreate_Successは送信がエピソードを非公開にし、作品のエピソード一覧に着地すること
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusSeeOther)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != want {
		t.Errorf("リダイレクト先 = %q、期待値 = %q", location, want)
	}
	if unpublishedAt := readEpisodeUnpublishedAt(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL、期待値 = 非公開の時刻")
	}
}

// TestCreate_NotFoundは、非公開にできないエピソードへの送信が、起きなかった書き込みを
// 報告せず404を返すことを検証する。
func TestCreate_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/episodes/{id}/archive", handler.Create)

	for name, target := range map[string]string{
		"存在しないエピソード": "/db/episodes/999999999/archive",
		"数値でないid":    "/db/episodes/abc/archive",
	} {
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, postRequest(target))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("ステータスコード = %d、期待値 = %d", status, http.StatusNotFound)
			}
		})
	}
}

// TestCreate_ForbiddenはHandlerの直接呼び出しでUseCaseの認可失敗を内部エラーとして
// 公開せず、403に変換することを検証する。
func TestCreate_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Post("/db/episodes/{id}/archive", handler.Create)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("POST", "/db/episodes/1/archive", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %d、期待値 = %d", status, http.StatusForbidden)
	}
}

// TestDelete_Successは再公開の送信が非公開の状態を解除し、作品のエピソード一覧に着地する
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusSeeOther)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if location := rr.Header().Get("Location"); location != want {
		t.Errorf("リダイレクト先 = %q、期待値 = %q", location, want)
	}
	if unpublishedAt := readEpisodeUnpublishedAt(t, db, episodeID); unpublishedAt.Valid {
		t.Errorf("episodes.unpublished_at = %v、期待値 = NULL", unpublishedAt.Time)
	}
}

// TestDelete_HTMXは、エピソード一覧の公開ボタンが行う送信に対し、htmxが追ってボタンに
// スワップしてしまうリダイレクトではなくHX-Redirectを返すことを検証する。
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
		t.Fatalf("ステータスコード = %d、期待値 = %d", status, http.StatusNoContent)
	}
	want := fmt.Sprintf("/db/works/%d/episodes", int64(workID))
	if redirect := rr.Header().Get("HX-Redirect"); redirect != want {
		t.Errorf("HX-Redirect = %q、期待値 = %q", redirect, want)
	}
	if location := rr.Header().Get("Location"); location != "" {
		t.Errorf("リダイレクト先 = %q、期待値 = 空 (htmxはHX-Redirectで遷移する)", location)
	}
	if unpublishedAt := readEpisodeUnpublishedAt(t, db, episodeID); unpublishedAt.Valid {
		t.Errorf("episodes.unpublished_at = %v、期待値 = NULL", unpublishedAt.Time)
	}
}

// TestDelete_NotFoundは、再公開できないエピソードへの送信が、起きなかった書き込みを報告
// せず404を返すことを検証する。
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
		"数値でないid":    "/db/episodes/abc/archive",
	} {
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, deleteRequest(target))

			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("ステータスコード = %d、期待値 = %d", status, http.StatusNotFound)
			}
		})
	}
}

// TestDelete_ForbiddenはHandlerの直接呼び出しでUseCaseの認可失敗を内部エラーとして
// 公開せず、403に変換することを検証する。
func TestDelete_Forbidden(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.Delete("/db/episodes/{id}/archive", handler.Delete)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest("DELETE", "/db/episodes/1/archive", nil))

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("ステータスコード = %d、期待値 = %d", status, http.StatusForbidden)
	}
}

// TestRequiresCommitterはHTTP境界ですべてのエンドポイントがcommitterロールによりゲート
// されていることを検証する。非公開・再公開の書き込みUseCaseはHTTP以外のentry pointに対して
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
					t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
				}
			})
		}
	}

	// committerはミドルウェアを通過し、弾かれてはならない。送信ではなく確認ページで確かめ
	// るのは、他のサブテストが非公開にした後でもエピソードが非公開にできる状態かどうかに、検証が
	// 左右されないようにするため。
	t.Run("編集者は通過", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/db/episodes/%d/archive/new", int64(episodeID)), nil)
		req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
		}
	})
}
