package db_episode_revision

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func newTestRouter(t *testing.T, db *sql.DB, tx *sql.Tx) chi.Router {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	handler := NewHandler(cfg, image.NewHelper(cfg), usecase.NewGetDBEpisodeRevisionsUsecase(
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRevisionRepository(queries),
	))

	r := chi.NewRouter()
	r.Get("/db/episodes/{id}/revisions", handler.Index)
	return r
}

func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソード履歴の作品").WithMedia(1).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第1話").WithTitle("はじまり").Build()
	animeID := testutil.CreateTestAnime(t, tx)
	if _, err := tx.Exec(`UPDATE episodes SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(episodeID)); err != nil {
		t.Fatalf("episodes.anime_idの設定に失敗: %v", err)
	}

	repo := repository.NewAnimeRevisionRepository(query.New(tx))
	before, after := "仮のタイトル", "はじまり"
	createdAt := time.Date(2026, 11, 8, 3, 4, 0, 0, time.UTC)
	createID := testutil.NewDBEditBuilder(t, tx).WithAction(model.DBEditActionCreate).WithCreatedAt(createdAt).Build()
	if err := repo.Create(ctx, createID, animeID, model.AnimeSnapshot{Title: &before, Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindEpisode}); err != nil {
		t.Fatalf("作成の版の作成に失敗: %v", err)
	}
	updateID := testutil.NewDBEditBuilder(t, tx).WithCreatedAt(createdAt.Add(time.Hour)).Build()
	if err := repo.Create(ctx, updateID, animeID, model.AnimeSnapshot{Title: &after, Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindEpisode}); err != nil {
		t.Fatalf("更新の版の作成に失敗: %v", err)
	}

	rr := httptest.NewRecorder()
	newTestRouter(t, db, tx).ServeHTTP(rr, httptest.NewRequest("GET", fmt.Sprintf("/db/episodes/%d/revisions", int64(episodeID)), nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"<title>エピソードの履歴 | 第1話 はじまり | エピソード履歴の作品 | Annict DB</title>",
		`<meta name="robots" content="noindex">`,
		"第1話 はじまりの履歴",
		`<time datetime="2026-11-08T03:04:00Z"`,
		"変更前",
		"仮のタイトル",
		// サブナビは作品のエピソードの項目に所属する。
		fmt.Sprintf(`href="/db/works/%d/episodes"`, int64(workID)),
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていない", expected)
		}
	}
	// ログインしていない閲覧者には、作品の履歴の項目と編集画面へ戻るリンクを出さない。
	for _, unexpected := range []string{
		fmt.Sprintf(`href="/db/works/%d/revisions"`, int64(workID)),
		fmt.Sprintf(`href="/db/episodes/%d/edit"`, int64(episodeID)),
	} {
		if strings.Contains(body, unexpected) {
			t.Errorf("レスポンスに%qが含まれている", unexpected)
		}
	}
}

func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		locale, emptyMessage, titlePrefix, heading string
	}{
		{"ja", "まだ履歴がありません", "<title>エピソードの履歴 | 名前の無いエピソード |", "エピソードの履歴"},
		{"en", "No history yet", "<title>Episode History | Untitled episode |", "Episode History"},
	} {
		t.Run(tt.locale, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			// animeにまだマッピングされていないエピソードは版を持たない。
			workID := testutil.NewWorkBuilder(t, tx).WithMedia(1).Build()
			episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("").WithTitle("").Build()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/db/episodes/%d/revisions", int64(episodeID)), nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), tt.locale))
			newTestRouter(t, db, tx).ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
			body := rr.Body.String()
			for _, expected := range []string{tt.emptyMessage, tt.titlePrefix, ">" + tt.heading + "</h2>"} {
				if !strings.Contains(body, expected) {
					t.Errorf("レスポンスに%qが含まれていない", expected)
				}
			}
		})
	}
}

func TestIndex_NotFound(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		target func(t *testing.T, tx *sql.Tx) string
	}{
		{name: "存在しないエピソード", target: func(t *testing.T, tx *sql.Tx) string {
			return fmt.Sprintf("/db/episodes/%d/revisions", int64(1)<<62)
		}},
		{name: "削除したエピソード", target: func(t *testing.T, tx *sql.Tx) string {
			workID := testutil.NewWorkBuilder(t, tx).WithMedia(1).Build()
			episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithDeletedAt(time.Now()).Build()
			return fmt.Sprintf("/db/episodes/%d/revisions", int64(episodeID))
		}},
		{name: "数値でないID", target: func(t *testing.T, tx *sql.Tx) string {
			return "/db/episodes/abc/revisions"
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			rr := httptest.NewRecorder()
			newTestRouter(t, db, tx).ServeHTTP(rr, httptest.NewRequest("GET", tt.target(t, tx), nil))

			if rr.Code != http.StatusNotFound {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNotFound)
			}
		})
	}
}
