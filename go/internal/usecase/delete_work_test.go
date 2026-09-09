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

// newDeleteWorkUsecase wires the delete-work usecase against the shared test DB. Like the
// archive / update usecases it opens its own transaction, so its tests use GetTestDB (not
// SetupTx) so the committed rows are visible to the usecase's inner transaction and to the
// follow-up sync invariant check.
//
// [Ja] newDeleteWorkUsecase は共有テスト DB に対して作品削除 UseCase を組み立てる。
// アーカイブ / 更新 UseCase と同じく内部で自前のトランザクションを開くため、テストは SetupTx
// ではなく GetTestDB を使い、コミット済みの行が UseCase の内側トランザクションと後続の同期
// 不変条件チェックから見えるようにする。
func newDeleteWorkUsecase(db *sql.DB) *DeleteWorkUsecase {
	queries := query.New(db)
	return NewDeleteWorkUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// TestDeleteWorkUsecase_Execute_DeletesWorkAndAnime verifies soft-deleting a mapped,
// published work sets works.deleted_at (the state source of truth) and dual-writes the
// derived anime.status = deleted, and that a phase 2 sync right after reports Unchanged (the
// delete and the reconciliation derive the same status from deleted_at, so the sync does not
// clobber the deleted anime back to published).
//
// [Ja] TestDeleteWorkUsecase_Execute_DeletesWorkAndAnime は、マッピング済みで公開中の work を
// ソフトデリートすると works.deleted_at (状態の正本) が立ち、導出された anime.status = deleted
// が両書きされること、および直後のフェーズ 2 同期が Unchanged を報告することを検証する (削除と
// リコンシリエーションが deleted_at から同じ status を導出するため、同期は削除済み anime を
// published に戻さない)。
func TestDeleteWorkUsecase_Execute_DeletesWorkAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "削除前アニメ_"+t.Name())
	animeID := *reloadSyncWork(t, db, workID).AnimeID

	if _, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// works.deleted_at is set (the work is now soft-deleted) and it stays mapped.
	//
	// [Ja] works.deleted_at が立ち (work はソフトデリートされた)、マッピングは維持される。
	work := reloadSyncWork(t, db, workID)
	if work.DeletedAt == nil {
		t.Error("work.DeletedAt should be set after deleting, got nil")
	}
	if work.DerivedStatus() != model.WorkStatusDeleted {
		t.Errorf("work.DerivedStatus() = %q, want deleted", work.DerivedStatus())
	}

	// The mapped anime carries the derived deleted status.
	//
	// [Ja] マッピング済み anime が導出された削除状態を持つ。
	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q, want deleted", anime.Status)
	}

	// A sync right after the delete must not clobber the deleted state back to published.
	//
	// [Ja] 削除直後の同期が削除状態を published に戻さないこと。
	result, err := newSyncUsecase(db).Execute(ctx, SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("sync result = %+v, want Unchanged:1 Updated:0", result)
	}
}

// TestDeleteWorkUsecase_Execute_DeletesArchivedWork verifies an archived work (unpublished_at
// set) is deletable, matching the Rails scope Work.without_deleted (published or archived).
// deleted_at wins over unpublished_at in DerivedStatus, so the mapped anime becomes deleted.
//
// [Ja] TestDeleteWorkUsecase_Execute_DeletesArchivedWork は、アーカイブ済み (unpublished_at 有)
// の work が削除可能であることを検証する。Rails の scope Work.without_deleted (公開中または
// アーカイブ済み) に一致する。DerivedStatus では deleted_at が unpublished_at より優先されるため、
// マッピング済み anime は deleted になる。
func TestDeleteWorkUsecase_Execute_DeletesArchivedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "アーカイブ済み削除_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET unpublished_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("unpublished_at の設定に失敗: %v", err)
	}
	animeID := *reloadSyncWork(t, db, workID).AnimeID

	if _, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	if work.DeletedAt == nil {
		t.Error("work.DeletedAt should be set after deleting, got nil")
	}
	if work.DerivedStatus() != model.WorkStatusDeleted {
		t.Errorf("work.DerivedStatus() = %q, want deleted", work.DerivedStatus())
	}

	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q, want deleted", anime.Status)
	}
}

// TestDeleteWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork verifies that an unmapped
// work (anime_id NULL) is soft-deleted on the works side only: the usecase sets
// works.deleted_at and does not create an anime, leaving that to the sync batch.
//
// [Ja] TestDeleteWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork は、未マッピングの work
// (anime_id NULL) が works 側だけソフトデリートされることを検証する。UseCase は
// works.deleted_at を立てるが anime を作らず、同期バッチに委ねる。
func TestDeleteWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "未マッピング削除_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET anime_id = NULL WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("anime_id のクリアに失敗: %v", err)
	}

	if _, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	if work.DeletedAt == nil {
		t.Error("work.DeletedAt should be set after deleting, got nil")
	}
	if work.AnimeID != nil {
		t.Errorf("work.AnimeID = %v, want nil (unmapped work stays unmapped)", *work.AnimeID)
	}
}

// TestDeleteWorkUsecase_Execute_ReturnsNotFoundForDeletedWork verifies deleting a work that
// is already soft-deleted is rejected as not found, matching the Rails scope
// Work.without_deleted.
//
// [Ja] TestDeleteWorkUsecase_Execute_ReturnsNotFoundForDeletedWork は、すでにソフトデリート
// 済みの work の削除が not found として弾かれることを検証する。Rails の scope
// Work.without_deleted に一致する。
func TestDeleteWorkUsecase_Execute_ReturnsNotFoundForDeletedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "二重削除_"+t.Name())
	if _, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID}); err != nil {
		t.Fatalf("最初の Execute() error = %v", err)
	}

	output, err := uc.Execute(ctx, DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v, want nil for an already-deleted work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestDeleteWorkUsecase_Execute_ReturnsNotFoundForMissingWork verifies a non-existent work
// id is reported as not found.
//
// [Ja] TestDeleteWorkUsecase_Execute_ReturnsNotFoundForMissingWork は、存在しない work id が
// not found として報告されることを検証する。
func TestDeleteWorkUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteWorkUsecase(db)

	output, err := uc.Execute(context.Background(), DeleteWorkInput{User: &model.User{ID: 1, Role: model.RoleAdmin}, WorkID: model.WorkID(1 << 62)})
	if output != nil {
		t.Errorf("output = %+v, want nil for a missing work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestDeleteWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWrite verifies the authorization
// boundary rejects every non-admin role before the work is read or written. An editor is
// included because deletion is admin-only (ADR 0009) while the other work state changes accept
// a committer.
//
// [Ja] TestDeleteWorkUsecase_Execute_RejectsUnauthorizedUserBeforeWrite は、認可境界が work の
// 読み書きより前に admin 以外の全ロールを拒否することを検証する。他の作品の状態変更が committer
// を受け付けるのに対し削除は admin 専用 (ADR 0009) であるため、編集者も対象に含める。
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
				t.Errorf("output = %+v, want nil for an unauthorized user", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("expected AppErrCodeForbidden, got %v", err)
			}
		})
	}

	// The work is still published: the rejection lands before the delete is written.
	//
	// [Ja] work は公開中のまま。拒否は削除の書き込みより前に起きる。
	if got := reloadSyncWork(t, db, workID).DerivedStatus(); got != model.WorkStatusPublished {
		t.Errorf("DerivedStatus() = %q, want %q", got, model.WorkStatusPublished)
	}
}
