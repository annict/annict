package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newArchiveWorkUsecaseは共有テストDBに対して作品非公開UseCaseを組み立てる。
// 作成 / 更新UseCaseと同じく内部で自前のトランザクションを開くため、テストはSetupTxでは
// なくGetTestDBを使い、コミット済みの行がUseCaseの内側トランザクションと後続の同期不変
// 条件チェックから見えるようにする。
func newArchiveWorkUsecase(db *sql.DB) *ArchiveWorkUsecase {
	queries := query.New(db)
	return NewArchiveWorkUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// TestArchiveWorkUsecase_Execute_ArchivesWorkAndAnimeは、マッピング済みで公開中の
// workを非公開にするとworks.unpublished_at (状態の正本) が立ち、導出された
// anime.status = archivedが両書きされること、および直後のフェーズ2同期がUnchangedを
// 報告することを検証する (非公開とリコンシリエーションがunpublished_atから同じstatusを
// 導出するため、同期はアーカイブ済みanimeをpublishedに戻さない)。
func TestArchiveWorkUsecase_Execute_ArchivesWorkAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "非公開前アニメ_"+t.Name())
	animeID := *reloadSyncWork(t, db, workID).AnimeID

	if _, err := uc.Execute(ctx, ArchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	// works.unpublished_atが立ち (workは非公開になった)、マッピングは維持される。
	work := reloadSyncWork(t, db, workID)
	if work.UnpublishedAt == nil {
		t.Error("アーカイブ後のwork.UnpublishedAt = nil、期待値 = 値あり")
	}
	if work.DeletedAt != nil {
		t.Error("アーカイブ後のwork.DeletedAtがnilでなくなった")
	}
	if work.DerivedStatus() != model.WorkStatusArchived {
		t.Errorf("work.DerivedStatus() = %q、期待値 = archived", work.DerivedStatus())
	}

	// マッピング済みanimeが導出されたアーカイブ状態を持つ。
	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = archived", anime.Status)
	}

	// 非公開直後の同期がアーカイブ状態をpublishedに戻さないこと。
	result, err := newSyncUsecase(db).Execute(ctx, SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("同期の結果 = %+v、期待値 = Unchanged:1 Updated:0", result)
	}
}

// TestArchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWorkは、未マッピングの
// work (anime_id NULL) がworks側だけ非公開になることを検証する。UseCaseは
// works.unpublished_atを立てるがanimeを作らず、同期バッチに委ねる。
func TestArchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "未マッピング非公開_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET anime_id = NULL WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("anime_idのクリアに失敗: %v", err)
	}

	if _, err := uc.Execute(ctx, ArchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	if work.UnpublishedAt == nil {
		t.Error("アーカイブ後のwork.UnpublishedAt = nil、期待値 = 値あり")
	}
	if work.AnimeID != nil {
		t.Errorf("work.AnimeID = %v、期待値 = nil (対応付けの無い作品はそのままであること)", *work.AnimeID)
	}
}

// TestArchiveWorkUsecase_Execute_ReturnsNotFoundForAlreadyArchivedWorkは、現在公開中
// でない (すでにアーカイブ済みの) workの非公開がnot foundとして弾かれることを検証する。
// Railsのscope Work.without_deleted.publishedに一致する。
func TestArchiveWorkUsecase_Execute_ReturnsNotFoundForAlreadyArchivedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "二重非公開_"+t.Name())
	if _, err := uc.Execute(ctx, ArchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("最初のExecute()のエラー = %v", err)
	}

	output, err := uc.Execute(ctx, ArchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (アーカイブ済みの作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestArchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWorkは、ソフトデリート済み
// のworkの非公開がnot foundとして弾かれることを検証する (削除済みはwithout_deletedの
// 対象外)。
func TestArchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "削除済み非公開_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET deleted_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("deleted_atの設定に失敗: %v", err)
	}

	output, err := uc.Execute(ctx, ArchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (削除済みの作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestArchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWorkは、存在しないwork id
// がnot foundとして報告されることを検証する。
func TestArchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)

	output, err := uc.Execute(context.Background(), ArchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: model.WorkID(1 << 62)})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (存在しない作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestArchiveWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWriteは、認可境界がworkの
// 読み書きより前に未認証と一般ユーザーを拒否することを検証する。committerでゲートされた
// ルート以外からUseCaseに到達した呼び出し元が作品を非公開にできないようにするため。
func TestArchiveWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWrite(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "非公開認可テストアニメ_"+t.Name())

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "未認証", user: nil},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(ctx, ArchiveWorkInput{User: tt.user, WorkID: workID})
			if output != nil {
				t.Errorf("output = %+v、期待値 = nil (権限の無いユーザーのため)", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("エラーコード = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}

	// workは公開中のまま。拒否は非公開の書き込みより前に起きる。
	if got := reloadSyncWork(t, db, workID).DerivedStatus(); got != model.WorkStatusPublished {
		t.Errorf("DerivedStatus() = %q、期待値 = %q", got, model.WorkStatusPublished)
	}
}
