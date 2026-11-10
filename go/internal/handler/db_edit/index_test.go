package db_edit

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func newTestHandler(db *sql.DB, tx *sql.Tx) *Handler {
	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	return NewHandler(cfg, image.NewHelper(cfg), usecase.NewGetDBEditsUsecase(
		repository.NewDBEditRepository(queries),
		repository.NewAnimeRevisionRepository(queries),
		repository.NewPersonRevisionRepository(queries),
		repository.NewNumberFormatRepository(queries),
		repository.NewPrefectureRepository(queries),
	))
}

func newRequest(user *model.User) *http.Request {
	req := httptest.NewRequest("GET", "/db/edits", nil)
	if user == nil {
		return req
	}
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, user))
}

// createEditsは、作品の更新の操作と、人物の掲載停止の操作を作る。人物は掲載を再開して公開中に戻っている。
func createEdits(t *testing.T, tx *sql.Tx, username string) (model.WorkID, model.PersonID) {
	t.Helper()
	ctx := context.Background()

	animeID := testutil.CreateTestAnime(t, tx)
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("全体の履歴の作品").WithMedia(1).Build()
	if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(workID)); err != nil {
		t.Fatalf("works.anime_idの設定に失敗: %v", err)
	}
	userID := testutil.NewUserBuilder(t, tx).WithUsername(username).WithProfileImageData(`{"original":{"id":"avatar.jpg","storage":"store"}}`).Build()
	animeRepo := repository.NewAnimeRevisionRepository(query.New(tx))
	before, after := "全体の履歴の作品", "全体の履歴の作品 (改題)"
	createdAt := time.Date(2026, 11, 8, 3, 4, 0, 0, time.UTC)
	if err := animeRepo.CreateInitial(ctx, animeID, model.AnimeSnapshot{Title: &before, Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("起点の版の作成に失敗: %v", err)
	}
	editID := testutil.NewDBEditBuilder(t, tx).WithUser(userID).WithCreatedAt(createdAt).Build()
	if err := animeRepo.Create(ctx, editID, animeID, model.AnimeSnapshot{Title: &after, Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("作品の版の作成に失敗: %v", err)
	}

	personID := testutil.NewPersonBuilder(t, tx).WithName("掲載を再開した人物").Build()
	delistID := testutil.NewDBEditBuilder(t, tx).WithUser(userID).WithAction(model.DBEditActionDelist).WithCreatedAt(createdAt.Add(time.Hour)).Build()
	if err := repository.NewPersonRevisionRepository(query.New(tx)).Create(ctx, delistID, personID, model.PersonSnapshot{Name: "掲載を再開した人物", Status: model.PersonStatusDelisted}); err != nil {
		t.Fatalf("人物の版の作成に失敗: %v", err)
	}
	return workID, personID
}

func TestIndex(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	username := fmt.Sprintf("dbedits%d", time.Now().UnixNano()%1000000)
	workID, personID := createEdits(t, tx, username)

	rr := httptest.NewRecorder()
	newTestHandler(db, tx).Index(rr, newRequest(nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"<title>編集履歴 | Annict DB</title>",
		`<meta name="robots" content="noindex">`,
		`<time datetime="2026-11-08T03:04:00Z"`,
		fmt.Sprintf(`href="/@%s"`, username),
		// 編集者のアバターは、プロフィール画像から作った正方形の画像を装飾として示す。
		`alt="" width="20" height="20" class="size-5 rounded-full object-cover"`,
		"resize:fill-down:40:40",
		"作品:",
		fmt.Sprintf(`href="/db/works/%d/revisions"`, int64(workID)),
		"全体の履歴の作品 (改題)",
		"変更前",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていない", expected)
		}
	}
	// 管理者以外には、掲載を再開した人物の掲載停止の操作を出さない。
	if strings.Contains(body, fmt.Sprintf(`href="/db/people/%d/revisions"`, int64(personID))) {
		t.Error("管理者以外に掲載停止の操作が出ている")
	}
}

func TestIndex_Admin(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	_, personID := createEdits(t, tx, fmt.Sprintf("dbeditsadm%d", time.Now().UnixNano()%1000000))

	rr := httptest.NewRecorder()
	newTestHandler(db, tx).Index(rr, newRequest(&model.User{ID: 1, Role: model.RoleAdmin}))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"人物:",
		fmt.Sprintf(`href="/db/people/%d/revisions"`, int64(personID)),
		"掲載停止",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに%qが含まれていない", expected)
		}
	}
}
