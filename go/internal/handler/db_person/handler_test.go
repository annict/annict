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

	// 一括登録・更新・削除のUseCaseは編集履歴と同じトランザクションで書くため自前のトランザクションを開く。
	// テスト用トランザクションではなくプールに対して組み立て、前提の人物はコミットして作る。
	poolQueries := query.New(db)
	poolPersonRepo := repository.NewPersonRepository(poolQueries)
	poolPrefectureRepo := repository.NewPrefectureRepository(poolQueries)
	dbEditRecordRepos := usecase.DBEditRecordRepos{
		DBEdit:         repository.NewDBEditRepository(poolQueries),
		AnimeRevision:  repository.NewAnimeRevisionRepository(poolQueries),
		PersonRevision: repository.NewPersonRevisionRepository(poolQueries),
	}
	createPeopleUC := usecase.NewCreatePeopleUsecase(db, poolPersonRepo, validator.NewDBPersonCreateValidator(poolPersonRepo), dbEditRecordRepos)

	// 編集の取得は、テスト用トランザクションに対して組み立てる。
	prefectureRepo := repository.NewPrefectureRepository(queries)

	return NewHandler(
		cfg,
		sessionManager,
		testutil.NewTestFlashManager(),
		usecase.NewGetDBPeopleUsecase(personRepo, repository.NewFeatureFlagRepository(queries)),
		usecase.NewGetDBPersonNewUsecase(),
		createPeopleUC,
		usecase.NewGetDBPersonEditUsecase(personRepo, prefectureRepo, repository.NewFeatureFlagRepository(queries)),
		usecase.NewUpdatePersonUsecase(db, poolPersonRepo, validator.NewDBPersonUpdateValidator(poolPersonRepo, poolPrefectureRepo), dbEditRecordRepos),
		usecase.NewDeletePersonUsecase(db, poolPersonRepo, dbEditRecordRepos),
	)
}
