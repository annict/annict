package db_work_revision

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
	"github.com/annict/annict/go/internal/image"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
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
	handler := NewHandler(cfg, image.NewHelper(cfg), usecase.NewGetDBWorkRevisionsUsecase(
		repository.NewWorkRepository(queries),
		repository.NewAnimeRevisionRepository(queries),
		repository.NewNumberFormatRepository(queries),
	))

	r := chi.NewRouter()
	r.Get("/db/works/{id}/revisions", handler.Index)
	return r
}

// createWorkWithRevisionsは、animeにマッピングした作品と、作成と更新の2つの版と、配下のエピソードの作成の版を作る。
func createWorkWithRevisions(t *testing.T, tx *sql.Tx, username string) (model.WorkID, model.EpisodeID) {
	t.Helper()
	ctx := context.Background()

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("履歴テスト作品").WithMedia(1).Build()
	animeID := testutil.CreateTestAnime(t, tx)
	if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(workID)); err != nil {
		t.Fatalf("works.anime_idの設定に失敗: %v", err)
	}

	userID := testutil.NewUserBuilder(t, tx).WithUsername(username).WithProfileImageData(`{"original":{"id":"avatar.jpg","storage":"store"}}`).Build()
	repo := repository.NewAnimeRevisionRepository(query.New(tx))
	before := "履歴テスト作品"
	after := "履歴テスト作品 (改題)"
	createdAt := time.Date(2026, 11, 8, 3, 4, 0, 0, time.UTC)
	createID := testutil.NewDBEditBuilder(t, tx).WithUser(userID).WithAction(model.DBEditActionCreate).WithCreatedAt(createdAt).Build()
	if err := repo.Create(ctx, createID, animeID, model.AnimeSnapshot{Title: &before, Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("作成の版の作成に失敗: %v", err)
	}
	updateID := testutil.NewDBEditBuilder(t, tx).WithCreatedAt(createdAt.Add(time.Hour)).Build()
	if err := repo.Create(ctx, updateID, animeID, model.AnimeSnapshot{
		Title:    &after,
		Status:   model.AnimeStatusPublished,
		Kind:     model.AnimeClassificationKindWork,
		Hashtags: []model.AnimeSnapshotHashtag{{Hashtag: "rireki"}},
	}); err != nil {
		t.Fatalf("更新の版の作成に失敗: %v", err)
	}
	// 配下のエピソードの版も、作品の履歴に並べる。
	parentAnimeID := animeID
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	episodeAnimeID := testutil.CreateTestAnime(t, tx)
	if _, err := tx.Exec(`UPDATE episodes SET anime_id = $1 WHERE id = $2`, int64(episodeAnimeID), int64(episodeID)); err != nil {
		t.Fatalf("episodes.anime_idの設定に失敗: %v", err)
	}
	number, episodeTitle := "第1話", "配下のエピソード"
	episodeEditID := testutil.NewDBEditBuilder(t, tx).WithAction(model.DBEditActionCreate).WithCreatedAt(createdAt.Add(2 * time.Hour)).Build()
	if err := repo.Create(ctx, episodeEditID, episodeAnimeID, model.AnimeSnapshot{
		Title:         &episodeTitle,
		NumberText:    &number,
		Status:        model.AnimeStatusPublished,
		Kind:          model.AnimeClassificationKindEpisode,
		ParentAnimeID: &parentAnimeID,
	}); err != nil {
		t.Fatalf("エピソードの版の作成に失敗: %v", err)
	}
	return workID, episodeID
}

func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	username := fmt.Sprintf("rireki%d", time.Now().UnixNano()%1000000)
	workID, episodeID := createWorkWithRevisions(t, tx, username)

	rr := httptest.NewRecorder()
	newTestRouter(t, db, tx).ServeHTTP(rr, httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/revisions", int64(workID)), nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"<title>作品の履歴 | 履歴テスト作品 | Annict DB</title>",
		`<meta name="robots" content="noindex">`,
		// 日時は機械可読の値をdatetimeに、表示は閲覧者のタイムゾーン (未ログインは日本時間) で示す。
		`<time datetime="2026-11-08T03:04:00Z"`,
		"2026-11-08 12:04",
		`<time datetime="2026-11-08T04:04:00Z"`,
		fmt.Sprintf(`href="/@%s"`, username),
		// 編集者のアバターは、プロフィール画像から作った正方形の画像を装飾として示す。
		`alt="" width="20" height="20" class="size-5 rounded-full object-cover"`,
		"resize:fill-down:40:40",
		"システム",
		"作成",
		"更新",
		"<caption",
		`<th scope="col"`,
		"変更前",
		"履歴テスト作品 (改題)",
		"ハッシュタグ",
		"ハッシュタグ: rireki",
		// サブナビは履歴の項目を持たず、現在地の印を作品の項目に付ける。
		fmt.Sprintf(`href="/db/works/%d/edit" class="block rounded-full px-3 py-2 text-sm whitespace-nowrap transition-colors hover:bg-accent aria-[current=page]:bg-primary aria-[current=page]:text-primary-foreground aria-[current=page]:hover:bg-primary" aria-current="page"`, int64(workID)),
		// エピソードの版は、どのエピソードの版かを示し、エピソードの履歴へリンクする。
		"エピソード:",
		fmt.Sprintf(`href="/db/episodes/%d/revisions"`, int64(episodeID)),
		"第1話 配下のエピソード: 作成",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていない", expected)
		}
	}
	// 新しい版から並べる。
	if strings.Index(body, "2026-11-08 13:04") > strings.Index(body, "2026-11-08 12:04") {
		t.Error("更新の版が作成の版より後ろに並んでいる")
	}
}

func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	// animeにまだマッピングされていない作品は版を持たない。
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("履歴の無い作品").WithMedia(1).Build()

	rr := httptest.NewRecorder()
	newTestRouter(t, db, tx).ServeHTTP(rr, httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/revisions", int64(workID)), nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "まだ履歴がありません") {
		t.Error("レスポンスに空の履歴の案内が含まれていない")
	}
}

func TestIndex_NotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target func(t *testing.T, tx *sql.Tx) string
	}{
		{
			name: "存在しない作品",
			target: func(t *testing.T, tx *sql.Tx) string {
				return fmt.Sprintf("/db/works/%d/revisions", int64(1)<<62)
			},
		},
		{
			name: "削除した作品",
			target: func(t *testing.T, tx *sql.Tx) string {
				workID := testutil.NewWorkBuilder(t, tx).WithMedia(1).WithDeletedAt(time.Now()).Build()
				return fmt.Sprintf("/db/works/%d/revisions", int64(workID))
			},
		},
		{
			name: "数値でないID",
			target: func(t *testing.T, tx *sql.Tx) string {
				return "/db/works/abc/revisions"
			},
		},
	}
	for _, tt := range tests {
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

// TestIndex_EditLinkは、作品の履歴の見出しに、編集者 (管理者または編集者) にだけ編集画面へ戻るリンクを出すことを検証する。
func TestIndex_EditLink(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		viewer *model.User
		want   bool
	}{
		{name: "ログインしていない閲覧者", viewer: nil, want: false},
		{name: "一般ユーザー", viewer: &model.User{ID: 1, Role: model.RoleUser}, want: false},
		{name: "編集者", viewer: &model.User{ID: 1, Role: model.RoleEditor}, want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			workID, _ := createWorkWithRevisions(t, tx, fmt.Sprintf("rirekiedit%d", time.Now().UnixNano()%1000000))
			req := httptest.NewRequest("GET", fmt.Sprintf("/db/works/%d/revisions", int64(workID)), nil)
			if tt.viewer != nil {
				req = req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, tt.viewer))
			}
			rr := httptest.NewRecorder()
			newTestRouter(t, db, tx).ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
			if got := strings.Contains(rr.Body.String(), "編集に戻る"); got != tt.want {
				t.Errorf("編集に戻るリンクの有無 = %v、期待値 = %v", got, tt.want)
			}
			// 「一覧に戻る」と同じアイコンを、装飾としてテキストの左に付ける。
			withIcon := fmt.Sprintf(`href="/db/works/%d/edit" class="btn rounded-full" data-variant="outline" data-size="sm"><svg data-icon="inline-start" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256"><path d="M232,200a8`, int64(workID))
			if got := strings.Contains(rr.Body.String(), withIcon); got != tt.want {
				t.Errorf("アイコン付きの編集に戻るリンクの有無 = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}
