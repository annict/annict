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

// newUnarchiveWorkUsecaseは共有テストDBに対して作品再公開UseCaseを組み立てる。
// 非公開UseCaseと同じく内部で自前のトランザクションを開くため、テストはSetupTxではなく
// GetTestDBを使い、コミット済みの行がUseCaseの内側トランザクションと後続の同期不変条件
// チェックから見えるようにする。
func newUnarchiveWorkUsecase(db *sql.DB) *UnarchiveWorkUsecase {
	queries := query.New(db)
	return NewUnarchiveWorkUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// TestUnarchiveWorkUsecase_Execute_UnarchivesWorkAndAnimeは、マッピング済みで
// アーカイブ済みのworkを再公開するとworks.unpublished_at (状態の正本) がクリアされ、
// 導出されたanime.status = publishedが両書きされること、および直後のフェーズ2同期が
// Unchangedを報告することを検証する (再公開とリコンシリエーションがクリアされた
// unpublished_atから同じstatusを導出するため、同期は公開済みanimeをarchivedに戻さない)。
func TestUnarchiveWorkUsecase_Execute_UnarchivesWorkAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()

	// 再公開前にanimeが真に整合したアーカイブ状態になるよう、非公開UseCase経由で
	// アーカイブ済みでマッピング済みのworkを用意する。
	workID := createMappedWork(t, db, "再公開前アニメ_"+t.Name())
	animeID := *reloadSyncWork(t, db, workID).AnimeID
	if _, err := newArchiveWorkUsecase(db).Execute(ctx, ArchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("前提のアーカイブに失敗: %v", err)
	}

	if _, err := newUnarchiveWorkUsecase(db).Execute(ctx, UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	// works.unpublished_atがクリアされ (workは再び公開された)、マッピングは維持される。
	work := reloadSyncWork(t, db, workID)
	if work.UnpublishedAt != nil {
		t.Errorf("再公開後のwork.UnpublishedAt = %v、期待値 = nil", *work.UnpublishedAt)
	}
	if work.DeletedAt != nil {
		t.Error("再公開後のwork.DeletedAtがnilでなくなった")
	}
	if work.DerivedStatus() != model.WorkStatusPublished {
		t.Errorf("work.DerivedStatus() = %q、期待値 = published", work.DerivedStatus())
	}

	// マッピング済みanimeが導出された公開状態を持つ。
	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q、期待値 = published", anime.Status)
	}

	// 再公開直後の同期が公開状態をarchivedに戻さないこと。
	result, err := newSyncUsecase(db).Execute(ctx, SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("同期の結果 = %+v、期待値 = Unchanged:1 Updated:0", result)
	}
}

// TestUnarchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWorkは、未マッピングの
// アーカイブ済みwork (anime_id NULL) がworks側だけ再公開されることを検証する。UseCaseは
// works.unpublished_atをクリアするがanimeを作らず、同期バッチに委ねる。
func TestUnarchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()

	workID := createMappedWork(t, db, "未マッピング再公開_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET unpublished_at = NOW(), anime_id = NULL WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("非公開・未マッピング状態の設定に失敗: %v", err)
	}

	if _, err := newUnarchiveWorkUsecase(db).Execute(ctx, UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	if work.UnpublishedAt != nil {
		t.Errorf("再公開後のwork.UnpublishedAt = %v、期待値 = nil", *work.UnpublishedAt)
	}
	if work.AnimeID != nil {
		t.Errorf("work.AnimeID = %v、期待値 = nil (対応付けの無い作品はそのままであること)", *work.AnimeID)
	}
}

// TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForPublishedWorkは、現在アーカイブ
// 済みでない (すでに公開中の) workの再公開がnot foundとして弾かれることを検証する。
// Railsのscope Work.without_deleted.unpublishedに一致する。
func TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForPublishedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveWorkUsecase(db)

	workID := createMappedWork(t, db, "公開中再公開_"+t.Name())

	output, err := uc.Execute(context.Background(), UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (公開済みの作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWorkは、ソフトデリート済み
// のworkの再公開がnot foundとして弾かれることを検証する (unpublished_atも立っていても、
// 削除済みはwithout_deletedの対象外)。
func TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "削除済み再公開_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET unpublished_at = NOW(), deleted_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("非公開・削除済み状態の設定に失敗: %v", err)
	}

	output, err := uc.Execute(ctx, UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (削除済みの作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWorkは、存在しないwork id
// がnot foundとして報告されることを検証する。
func TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveWorkUsecase(db)

	output, err := uc.Execute(context.Background(), UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: model.WorkID(1 << 62)})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (存在しない作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestUnarchiveWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWriteは、認可境界がworkの
// 読み書きより前に未認証と一般ユーザーを拒否することを検証する。committerでゲートされた
// ルート以外からUseCaseに到達した呼び出し元が作品を再公開できないようにするため。
func TestUnarchiveWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWrite(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()

	workID := createMappedWork(t, db, "公開認可テストアニメ_"+t.Name())
	if _, err := newArchiveWorkUsecase(db).Execute(ctx, ArchiveWorkInput{
		User:   &model.User{ID: 1, Role: model.RoleEditor},
		WorkID: workID,
	}); err != nil {
		t.Fatalf("前提の非公開に失敗: %v", err)
	}

	uc := newUnarchiveWorkUsecase(db)
	tests := []struct {
		name string
		user *model.User
	}{
		{name: "未認証", user: nil},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(ctx, UnarchiveWorkInput{User: tt.user, WorkID: workID})
			if output != nil {
				t.Errorf("output = %+v、期待値 = nil (権限の無いユーザーのため)", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("エラーコード = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}

	// workはアーカイブ済みのまま。拒否は再公開の書き込みより前に起きる。
	if got := reloadSyncWork(t, db, workID).DerivedStatus(); got != model.WorkStatusArchived {
		t.Errorf("DerivedStatus() = %q、期待値 = %q", got, model.WorkStatusArchived)
	}
}
