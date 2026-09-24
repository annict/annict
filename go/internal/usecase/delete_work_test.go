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

// newDeleteWorkUsecaseは共有テストDBに対して作品削除UseCaseを組み立てる。
// アーカイブ / 更新UseCaseと同じく内部で自前のトランザクションを開くため、テストはSetupTx
// ではなくGetTestDBを使い、コミット済みの行がUseCaseの内側トランザクションと後続の同期
// 不変条件チェックから見えるようにする。
func newDeleteWorkUsecase(db *sql.DB) *DeleteWorkUsecase {
	queries := query.New(db)
	return NewDeleteWorkUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// TestDeleteWorkUsecase_Execute_DeletesWorkAndAnimeは、マッピング済みで公開中のworkを
// ソフトデリートするとworks.deleted_at (状態の正本) が立ち、導出されたanime.status = deleted
// が両書きされること、および直後のフェーズ2同期がUnchangedを報告することを検証する (削除と
// リコンシリエーションがdeleted_atから同じstatusを導出するため、同期は削除済みanimeを
// publishedに戻さない)。
func TestDeleteWorkUsecase_Execute_DeletesWorkAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "削除前アニメ_"+t.Name())
	animeID := *reloadSyncWork(t, db, workID).AnimeID

	if _, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	// works.deleted_atが立ち (workはソフトデリートされた)、マッピングは維持される。
	work := reloadSyncWork(t, db, workID)
	if work.DeletedAt == nil {
		t.Error("削除後のwork.DeletedAt = nil、期待値 = 値あり")
	}
	if work.DerivedStatus() != model.WorkStatusDeleted {
		t.Errorf("work.DerivedStatus() = %q、期待値 = deleted", work.DerivedStatus())
	}

	// マッピング済みanimeが導出された削除状態を持つ。
	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q、期待値 = deleted", anime.Status)
	}

	// 削除直後の同期が削除状態をpublishedに戻さないこと。
	result, err := newSyncUsecase(db).Execute(ctx, SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("同期の結果 = %+v、期待値 = Unchanged:1 Updated:0", result)
	}
}

// TestDeleteWorkUsecase_Execute_DeletesArchivedWorkは、アーカイブ済み (unpublished_at有)
// のworkが削除可能であることを検証する。Railsのscope Work.without_deleted (公開中または
// アーカイブ済み) に一致する。DerivedStatusではdeleted_atがunpublished_atより優先されるため、
// マッピング済みanimeはdeletedになる。
func TestDeleteWorkUsecase_Execute_DeletesArchivedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "アーカイブ済み削除_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET unpublished_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("unpublished_atの設定に失敗: %v", err)
	}
	animeID := *reloadSyncWork(t, db, workID).AnimeID

	if _, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	if work.DeletedAt == nil {
		t.Error("削除後のwork.DeletedAt = nil、期待値 = 値あり")
	}
	if work.DerivedStatus() != model.WorkStatusDeleted {
		t.Errorf("work.DerivedStatus() = %q、期待値 = deleted", work.DerivedStatus())
	}

	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q、期待値 = deleted", anime.Status)
	}
}

// TestDeleteWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWorkは、未マッピングのwork
// (anime_id NULL) がworks側だけソフトデリートされることを検証する。UseCaseは
// works.deleted_atを立てるがanimeを作らず、同期バッチに委ねる。
func TestDeleteWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "未マッピング削除_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET anime_id = NULL WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("anime_idのクリアに失敗: %v", err)
	}

	if _, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	if work.DeletedAt == nil {
		t.Error("削除後のwork.DeletedAt = nil、期待値 = 値あり")
	}
	if work.AnimeID != nil {
		t.Errorf("work.AnimeID = %v、期待値 = nil (対応付けの無い作品はそのままであること)", *work.AnimeID)
	}
}

// TestDeleteWorkUsecase_Execute_ReturnsNotFoundForDeletedWorkは、すでにソフトデリート
// 済みのworkの削除がnot foundとして弾かれることを検証する。Railsのscope
// Work.without_deletedに一致する。
func TestDeleteWorkUsecase_Execute_ReturnsNotFoundForDeletedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "二重削除_"+t.Name())
	if _, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID}); err != nil {
		t.Fatalf("最初のExecute()のエラー = %v", err)
	}

	output, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (削除済みの作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestDeleteWorkUsecase_Execute_ReturnsNotFoundForMissingWorkは、存在しないwork idが
// not foundとして報告されることを検証する。
func TestDeleteWorkUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)

	output, err := uc.Execute(context.Background(), DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: model.WorkID(1 << 62)})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (存在しない作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestDeleteWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWriteは、認可境界がworkの
// 読み書きより前にadmin以外の全ロールを拒否することを検証する。他の作品の状態変更がcommitter
// を受け付けるのに対し削除はadmin専用 (ADR 0009) であるため、編集者も対象に含める。
func TestDeleteWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWrite(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "削除認可テストアニメ_"+t.Name())

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "未認証", user: nil},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(ctx, DeleteWorkInput{User: tt.user, WorkID: workID})
			if output != nil {
				t.Errorf("output = %+v、期待値 = nil (権限の無いユーザーのため)", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("エラーコード = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}

	// workは公開中のまま。拒否は削除の書き込みより前に起きる。
	if got := reloadSyncWork(t, db, workID).DerivedStatus(); got != model.WorkStatusPublished {
		t.Errorf("DerivedStatus() = %q、期待値 = %q", got, model.WorkStatusPublished)
	}
}
