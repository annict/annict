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

// newUnarchiveWorkUsecase wires the un-archive-work usecase against the shared test DB.
// Like the archive usecase it opens its own transaction, so its tests use GetTestDB (not
// SetupTx) so the committed rows are visible to the usecase's inner transaction and to the
// follow-up sync invariant check.
//
// [Ja] newUnarchiveWorkUsecase は共有テスト DB に対して作品再公開 UseCase を組み立てる。
// 非公開 UseCase と同じく内部で自前のトランザクションを開くため、テストは SetupTx ではなく
// GetTestDB を使い、コミット済みの行が UseCase の内側トランザクションと後続の同期不変条件
// チェックから見えるようにする。
func newUnarchiveWorkUsecase(db *sql.DB) *UnarchiveWorkUsecase {
	queries := query.New(db)
	return NewUnarchiveWorkUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// TestUnarchiveWorkUsecase_Execute_UnarchivesWorkAndAnime verifies re-publishing a mapped,
// archived work clears works.unpublished_at (the state source of truth) and dual-writes the
// derived anime.status = published, and that a phase 2 sync right after reports Unchanged
// (the re-publish and the reconciliation derive the same status from the cleared
// unpublished_at, so the sync does not clobber the published anime back to archived).
//
// [Ja] TestUnarchiveWorkUsecase_Execute_UnarchivesWorkAndAnime は、マッピング済みで
// アーカイブ済みの work を再公開すると works.unpublished_at (状態の正本) がクリアされ、
// 導出された anime.status = published が両書きされること、および直後のフェーズ 2 同期が
// Unchanged を報告することを検証する (再公開とリコンシリエーションがクリアされた
// unpublished_at から同じ status を導出するため、同期は公開済み anime を archived に戻さない)。
func TestUnarchiveWorkUsecase_Execute_UnarchivesWorkAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()

	// Set up an archived, mapped work through the archive usecase so the anime is a
	// genuinely consistent archived state before re-publishing.
	//
	// [Ja] 再公開前に anime が真に整合したアーカイブ状態になるよう、非公開 UseCase 経由で
	// アーカイブ済みでマッピング済みの work を用意する。
	workID := createMappedWork(t, db, "再公開前アニメ_"+t.Name())
	animeID := *reloadSyncWork(t, db, workID).AnimeID
	if _, err := newArchiveWorkUsecase(db).Execute(ctx, ArchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("前提のアーカイブに失敗: %v", err)
	}

	if _, err := newUnarchiveWorkUsecase(db).Execute(ctx, UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// works.unpublished_at is cleared (the work is published again) and it stays mapped.
	//
	// [Ja] works.unpublished_at がクリアされ (work は再び公開された)、マッピングは維持される。
	work := reloadSyncWork(t, db, workID)
	if work.UnpublishedAt != nil {
		t.Errorf("work.UnpublishedAt should be nil after re-publishing, got %v", *work.UnpublishedAt)
	}
	if work.DeletedAt != nil {
		t.Error("work.DeletedAt should stay nil after re-publishing")
	}
	if work.DerivedStatus() != model.WorkStatusPublished {
		t.Errorf("work.DerivedStatus() = %q, want published", work.DerivedStatus())
	}

	// The mapped anime carries the derived published status.
	//
	// [Ja] マッピング済み anime が導出された公開状態を持つ。
	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q, want published", anime.Status)
	}

	// A sync right after the re-publish must not clobber the published state back to archived.
	//
	// [Ja] 再公開直後の同期が公開状態を archived に戻さないこと。
	result, err := newSyncUsecase(db).Execute(ctx, SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("sync result = %+v, want Unchanged:1 Updated:0", result)
	}
}

// TestUnarchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork verifies that an unmapped
// archived work (anime_id NULL) is re-published on the works side only: the usecase clears
// works.unpublished_at and does not create an anime, leaving that to the sync batch.
//
// [Ja] TestUnarchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork は、未マッピングの
// アーカイブ済み work (anime_id NULL) が works 側だけ再公開されることを検証する。UseCase は
// works.unpublished_at をクリアするが anime を作らず、同期バッチに委ねる。
func TestUnarchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()

	workID := createMappedWork(t, db, "未マッピング再公開_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET unpublished_at = NOW(), anime_id = NULL WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("非公開・未マッピング状態の設定に失敗: %v", err)
	}

	if _, err := newUnarchiveWorkUsecase(db).Execute(ctx, UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	if work.UnpublishedAt != nil {
		t.Errorf("work.UnpublishedAt should be nil after re-publishing, got %v", *work.UnpublishedAt)
	}
	if work.AnimeID != nil {
		t.Errorf("work.AnimeID = %v, want nil (unmapped work stays unmapped)", *work.AnimeID)
	}
}

// TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForPublishedWork verifies re-publishing a
// work that is not currently archived (already published) is rejected as not found,
// matching the Rails scope Work.without_deleted.unpublished.
//
// [Ja] TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForPublishedWork は、現在アーカイブ
// 済みでない (すでに公開中の) work の再公開が not found として弾かれることを検証する。
// Rails の scope Work.without_deleted.unpublished に一致する。
func TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForPublishedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveWorkUsecase(db)

	workID := createMappedWork(t, db, "公開中再公開_"+t.Name())

	output, err := uc.Execute(context.Background(), UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v, want nil for an already-published work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWork verifies re-publishing a
// soft-deleted work is rejected as not found (deleted works are outside without_deleted),
// even when unpublished_at is also set.
//
// [Ja] TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWork は、ソフトデリート済み
// の work の再公開が not found として弾かれることを検証する (unpublished_at も立っていても、
// 削除済みは without_deleted の対象外)。
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
		t.Errorf("output = %+v, want nil for a deleted work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWork verifies a non-existent
// work id is reported as not found.
//
// [Ja] TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWork は、存在しない work id
// が not found として報告されることを検証する。
func TestUnarchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveWorkUsecase(db)

	output, err := uc.Execute(context.Background(), UnarchiveWorkInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: model.WorkID(1 << 62)})
	if output != nil {
		t.Errorf("output = %+v, want nil for a missing work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestUnarchiveWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWrite verifies the authorization
// boundary rejects an unauthenticated and a regular user before the work is read or written, so
// a caller reaching the usecase outside the committer-gated route cannot re-publish a work.
//
// [Ja] TestUnarchiveWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWrite は、認可境界が work の
// 読み書きより前に未認証と一般ユーザーを拒否することを検証する。committer でゲートされた
// ルート以外から UseCase に到達した呼び出し元が作品を再公開できないようにするため。
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
				t.Errorf("output = %+v, want nil for an unauthorized user", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("expected AppErrCodeForbidden, got %v", err)
			}
		})
	}

	// The work is still archived: the rejection lands before the re-publish is written.
	//
	// [Ja] work はアーカイブ済みのまま。拒否は再公開の書き込みより前に起きる。
	if got := reloadSyncWork(t, db, workID).DerivedStatus(); got != model.WorkStatusArchived {
		t.Errorf("DerivedStatus() = %q, want %q", got, model.WorkStatusArchived)
	}
}
