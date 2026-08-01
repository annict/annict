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
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	workRepo := repository.NewWorkRepository(queries)
	episodeRepo := repository.NewEpisodeRepository(queries)

	return NewHandler(cfg, usecase.NewGetDBEpisodesUsecase(workRepo, episodeRepo))
}

// newIndexRequest builds a GET request for a work's episode list with the work_id URL
// parameter chi would have extracted from the route pattern.
//
// [Ja] newIndexRequest はある作品のエピソード一覧への GET リクエストを、chi がルートパターン
// から取り出す work_id の URL パラメータ付きで組み立てる。
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
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	expectedContents := []string{
		// The DB pages carry the " | Annict DB" title suffix so they stay distinguishable
		// from the public pages in browser tabs.
		//
		// [Ja] DB のページはブラウザのタブで公開画面と区別できるよう " | Annict DB" の
		// タイトルサフィックスを持つ。
		"<title>テストアニメ | エピソード | Annict DB</title>",
		// The heading names the parent work, and the subnav links back to its form.
		//
		// [Ja] 見出しは親作品を名指しし、サブナビはそのフォームへ戻るリンクを持つ。
		"テストアニメ",
		fmt.Sprintf(`href="/db/works/%d/edit"`, int64(workID)),
		"<table",
		"<thead",
		"<tbody",
		"第1話",
		"はじまり",
		// The ID column links to the episode's public page in a new tab.
		//
		// [Ja] ID 列はエピソードの公開ページを新しいタブで開くリンクになる。
		fmt.Sprintf(`href="/works/%d/episodes/%d"`, int64(workID), int64(episodeID)),
		`target="_blank"`,
		fmt.Sprintf(`aria-label="エピソード %d を新しいタブで開く"`, int64(episodeID)),
		// A published episode shows the published badge.
		//
		// [Ja] 公開中のエピソードには公開のバッジが出る。
		`<span class="badge" data-variant="success">公開</span>`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
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
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
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
			t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})

	t.Run("削除済みの作品", func(t *testing.T) {
		rr := httptest.NewRecorder()
		handler.Index(rr, newIndexRequest(deletedWorkID, ""))

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})

	t.Run("数値でない work_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/db/works/abc/episodes", nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("work_id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

		rr := httptest.NewRecorder()
		handler.Index(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("status code: got %v want %v", status, http.StatusNotFound)
		}
		assertNotFoundPage(t, rr)
	})
}

func assertNotFoundPage(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", contentType)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>ページが見つかりません | Annict</title>",
		"ページが見つかりません",
		`href="/"`,
		"ホームに戻る",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("404 レスポンスに %q が含まれていません", expected)
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
			want:      "テストアニメ | エピソード | Annict DB",
		},
		{
			name:      "日本語の2ページ目",
			locale:    "ja",
			workTitle: "テストアニメ",
			page:      2,
			want:      "テストアニメ | エピソード (2ページ目) | Annict DB",
		},
		{
			name:      "英語の2ページ目",
			locale:    "en",
			workTitle: "Test Anime",
			page:      2,
			want:      "Test Anime | Episodes (Page 2) | Annict DB",
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
				t.Errorf("meta.Title = %q, want %q", meta.Title, tt.want)
			}
		})
	}
}

// canonicalTag builds the og:url tag a work's episode list carries for the given query. The
// DB pages carry their representative URL as og:url (DBHead leaves out the canonical link,
// since robots.txt disallows these pages), so the assertions read that tag.
//
// [Ja] canonicalTag は、ある作品のエピソード一覧が指定のクエリで持つ og:url タグを組み立てる。
// DB のページは代表 URL を og:url として持つ (robots.txt でクロールを禁止しているため DBHead
// は canonical のリンクを出さない)。そのためこのタグを検証する。
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
		// The first page and its ?page=1 form share one representative URL.
		//
		// [Ja] 1 ページ目と ?page=1 の形は 1 つの代表 URL を共有する。
		{name: "ページ指定なし", query: "", wantCanonical: canonicalTag(workID, ""), wantTitle: "テストアニメ | エピソード | Annict DB"},
		{name: "page=1", query: "?page=1", wantCanonical: canonicalTag(workID, ""), wantTitle: "テストアニメ | エピソード | Annict DB"},
		// A page number that is not a positive integer falls back to the first page.
		//
		// [Ja] 正の整数でないページ番号は 1 ページ目にフォールバックする。
		{name: "不正なページ番号", query: "?page=abc", wantCanonical: canonicalTag(workID, ""), wantTitle: "テストアニメ | エピソード | Annict DB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.Index(rr, newIndexRequest(workID, tt.query))

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("status code: got %v want %v", status, http.StatusOK)
			}

			if !strings.Contains(rr.Body.String(), tt.wantCanonical) {
				t.Errorf("canonical URL に %q が含まれていません", tt.wantCanonical)
			}

			if !strings.Contains(rr.Body.String(), "<title>"+tt.wantTitle+"</title>") {
				t.Errorf("title に %q が含まれていません", tt.wantTitle)
			}
		})
	}
}

// TestIndex_SecondPage covers a page number the work really has: the representative URL and
// the document title name that page, and only the rows belonging to it are listed.
//
// [Ja] TestIndex_SecondPage は作品が実際に持つページ番号を検証する。代表 URL と文書タイトルが
// そのページを名乗り、そのページに属する行だけが並ぶ。
func TestIndex_SecondPage(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()

	// One episode past the 100-per-page boundary, so the second page holds exactly the
	// lowest-sorted one. The rows go in with a single statement: the builder inserts one row
	// per call, and the list order here depends on sort_number, which it fixes.
	//
	// [Ja] 1 ページ 100 件の境界をちょうど 1 件超える件数を入れ、2 ページ目に並び順が最も
	// 小さい 1 件だけが載るようにする。行は 1 文で投入する。ビルダーは 1 回の呼び出しにつき
	// 1 行を挿入するうえ、ここでの一覧の並びはビルダーが固定する sort_number に依存するため。
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
		t.Fatalf("status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()

	for _, expected := range []string{
		canonicalTag(workID, "?page=2"),
		"<title>テストアニメ | エピソード (2ページ目) | Annict DB</title>",
		// sort_number ascends with the episode number and the list is sorted descending, so
		// the second page holds the first episode alone.
		//
		// [Ja] sort_number は話数とともに増え、一覧は降順に並ぶため、2 ページ目には第1話が
		// 1 件だけ載る。
		"第1話",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	if strings.Contains(body, "第101話") {
		t.Error("1 ページ目のエピソードが 2 ページ目に含まれています")
	}
}

// TestIndex_PageBeyondLastRedirects covers page numbers past the end of the list. Rendering
// them would leave the reader on an empty list that names a page the work does not have, with
// the pagination dropped alongside the table, so they are sent to the last page instead.
//
// [Ja] TestIndex_PageBeyondLastRedirects は一覧の末尾を超えるページ番号を検証する。そのまま
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
		// A work with no episodes still has a first page, so its empty list stays reachable
		// at the path without a page number.
		//
		// [Ja] エピソードが無い作品も 1 ページ目は持つため、空の一覧はページ番号なしのパスで
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
				t.Fatalf("status code: got %v want %v", status, http.StatusFound)
			}
			if location := rr.Header().Get("Location"); location != tt.wantLocation {
				t.Errorf("Location = %q, want %q", location, tt.wantLocation)
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
				t.Errorf("lastPageNumber(%d, %d) = %d, want %d", tt.totalCount, tt.perPage, got, tt.want)
			}
		})
	}
}
