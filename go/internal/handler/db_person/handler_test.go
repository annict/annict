package db_person

import (
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/config"
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
	sessionManager := session.NewManager(repository.NewSessionRepository(queries), cfg)
	personRepo := repository.NewPersonRepository(queries)

	// 一括登録UseCaseは自前のトランザクションを開くため、テスト用トランザクションではなく
	// プールに対して組み立てる。コミットされた行はそのテスト側で後始末する。
	poolPersonRepo := repository.NewPersonRepository(query.New(db))
	createPeopleUC := usecase.NewCreatePeopleUsecase(db, poolPersonRepo, validator.NewDBPersonCreateValidator(poolPersonRepo))

	// 編集の取得・更新と削除のUseCaseは自前のトランザクションを開かないため、テスト用
	// トランザクションに対して組み立て、書き込みはテスト終了時に巻き戻す。
	prefectureRepo := repository.NewPrefectureRepository(queries)

	return NewHandler(
		cfg,
		sessionManager,
		testutil.NewTestFlashManager(),
		usecase.NewGetDBPeopleUsecase(personRepo),
		usecase.NewGetDBPersonNewUsecase(),
		createPeopleUC,
		usecase.NewGetDBPersonEditUsecase(personRepo, prefectureRepo),
		usecase.NewUpdatePersonUsecase(personRepo, validator.NewDBPersonUpdateValidator(personRepo, prefectureRepo)),
		usecase.NewDeletePersonUsecase(personRepo),
	)
}
