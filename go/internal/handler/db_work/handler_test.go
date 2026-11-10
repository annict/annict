package db_work

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	queries := query.New(db)
	if tx != nil {
		queries = queries.WithTx(tx)
	}
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	animeRepo := repository.NewAnimeRepository(queries)
	animeClassificationRepo := repository.NewAnimeClassificationRepository(queries)

	dbEditRecordRepos := usecase.DBEditRecordRepos{
		DBEdit:         repository.NewDBEditRepository(queries),
		AnimeRevision:  repository.NewAnimeRevisionRepository(queries),
		PersonRevision: repository.NewPersonRevisionRepository(queries),
	}
	satelliteRepos := usecase.WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}

	getDBWorksUC := usecase.NewGetDBWorksUsecase(workRepo, repository.NewFeatureFlagRepository(queries))
	getDBWorkFormOptionsUC := usecase.NewGetDBWorkFormOptionsUsecase(numberFormatRepo)
	getDBWorkEditUC := usecase.NewGetDBWorkEditUsecase(workRepo, numberFormatRepo, repository.NewFeatureFlagRepository(queries))
	createWorkUC := usecase.NewCreateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, dbEditRecordRepos, validator.NewDBWorkCreateValidator(workRepo, numberFormatRepo))
	updateWorkUC := usecase.NewUpdateWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, dbEditRecordRepos, validator.NewDBWorkCreateValidator(workRepo, numberFormatRepo))
	deleteWorkUC := usecase.NewDeleteWorkUsecase(db, workRepo, animeRepo, animeClassificationRepo, satelliteRepos, dbEditRecordRepos)

	return NewHandler(cfg, sessionManager, testutil.NewTestFlashManager(), testutil.NewTestImageHelper(), getDBWorksUC, getDBWorkFormOptionsUC, getDBWorkEditUC, createWorkUC, updateWorkUC, deleteWorkUC)
}

// withCommitterUserはルートのミドルウェア (RequireCommitter) を通過したcommitterの
// ユーザーをリクエストのコンテキストに載せる。ハンドラーを直接呼ぶテストでも、UseCaseの
// 認可を通過させるため。
func withCommitterUser(req *http.Request) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), authMiddleware.UserContextKey, &model.User{ID: 1, Role: model.RoleEditor}))
}

// newCommittedEditorは、作品の作成のようにUseCaseが自前のトランザクションでコミットする送信の編集者を、
// 共有テストDBにコミットして作る。編集履歴の操作がusersを外部キーで指すため、行の無いユーザーでは作成できない。
// テストの終わりに削除する。
func newCommittedEditor(t *testing.T, db *sql.DB) *model.User {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("編集者の作成トランザクションの開始に失敗: %v", err)
	}
	userID := testutil.NewUserBuilder(t, tx).WithRole(model.RoleEditor).Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("編集者の作成トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() { testutil.DeleteUser(t, db, userID) })

	return &model.User{ID: userID, Role: model.RoleEditor}
}

// assertNotFoundPageは404が共通のエラーページとして配信されることを検証する。古い
// リンクを辿った読み手が、何が起きたかを述べ戻る導線を持つページに着地するようにするため。
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

// assertForbiddenPageは認可失敗が共通の403ページとして配信されることを検証する。
func assertForbiddenPage(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if rr.Code != http.StatusForbidden {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusForbidden)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q、期待値 = text/html; charset=utf-8", contentType)
	}
	for _, expected := range []string{"<title>アクセスできません | Annict</title>", "この操作を行う権限がありません。", `href="/"`} {
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("403レスポンスに%qが含まれていません", expected)
		}
	}
}
