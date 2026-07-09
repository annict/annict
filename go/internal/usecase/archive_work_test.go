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

// newArchiveWorkUsecase wires the archive-work usecase against the shared test DB. Like
// the create / update usecases it opens its own transaction, so its tests use GetTestDB
// (not SetupTx) so the committed rows are visible to the usecase's inner transaction and
// to the follow-up sync invariant check.
//
// [Ja] newArchiveWorkUsecase は共有テスト DB に対して作品非公開 UseCase を組み立てる。
// 作成 / 更新 UseCase と同じく内部で自前のトランザクションを開くため、テストは SetupTx では
// なく GetTestDB を使い、コミット済みの行が UseCase の内側トランザクションと後続の同期不変
// 条件チェックから見えるようにする。
func newArchiveWorkUsecase(db *sql.DB) *ArchiveWorkUsecase {
	queries := query.New(db)
	return NewArchiveWorkUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// TestArchiveWorkUsecase_Execute_ArchivesWorkAndAnime verifies archiving a mapped,
// published work sets works.unpublished_at (the state source of truth) and dual-writes the
// derived anime.status = archived, and that a phase 2 sync right after reports Unchanged
// (the archive and the reconciliation derive the same status from unpublished_at, so the
// sync does not clobber the archived anime back to published).
//
// [Ja] TestArchiveWorkUsecase_Execute_ArchivesWorkAndAnime は、マッピング済みで公開中の
// work を非公開にすると works.unpublished_at (状態の正本) が立ち、導出された
// anime.status = archived が両書きされること、および直後のフェーズ 2 同期が Unchanged を
// 報告することを検証する (非公開とリコンシリエーションが unpublished_at から同じ status を
// 導出するため、同期はアーカイブ済み anime を published に戻さない)。
func TestArchiveWorkUsecase_Execute_ArchivesWorkAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "非公開前アニメ_"+t.Name())
	animeID := *reloadSyncWork(t, db, workID).AnimeID

	if _, err := uc.Execute(ctx, ArchiveWorkInput{WorkID: workID}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// works.unpublished_at is set (the work is now archived) and it stays mapped.
	//
	// [Ja] works.unpublished_at が立ち (work は非公開になった)、マッピングは維持される。
	work := reloadSyncWork(t, db, workID)
	if work.UnpublishedAt == nil {
		t.Error("work.UnpublishedAt should be set after archiving, got nil")
	}
	if work.DeletedAt != nil {
		t.Error("work.DeletedAt should stay nil after archiving")
	}
	if work.DerivedStatus() != model.WorkStatusArchived {
		t.Errorf("work.DerivedStatus() = %q, want archived", work.DerivedStatus())
	}

	// The mapped anime carries the derived archived status.
	//
	// [Ja] マッピング済み anime が導出されたアーカイブ状態を持つ。
	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want archived", anime.Status)
	}

	// A sync right after the archive must not clobber the archived state back to published.
	//
	// [Ja] 非公開直後の同期がアーカイブ状態を published に戻さないこと。
	result, err := newSyncUsecase(db).Execute(ctx, SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("sync result = %+v, want Unchanged:1 Updated:0", result)
	}
}

// TestArchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork verifies that an unmapped
// work (anime_id NULL) is archived on the works side only: the usecase sets
// works.unpublished_at and does not create an anime, leaving that to the sync batch.
//
// [Ja] TestArchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork は、未マッピングの
// work (anime_id NULL) が works 側だけ非公開になることを検証する。UseCase は
// works.unpublished_at を立てるが anime を作らず、同期バッチに委ねる。
func TestArchiveWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "未マッピング非公開_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET anime_id = NULL WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("anime_id のクリアに失敗: %v", err)
	}

	if _, err := uc.Execute(ctx, ArchiveWorkInput{WorkID: workID}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	if work.UnpublishedAt == nil {
		t.Error("work.UnpublishedAt should be set after archiving, got nil")
	}
	if work.AnimeID != nil {
		t.Errorf("work.AnimeID = %v, want nil (unmapped work stays unmapped)", *work.AnimeID)
	}
}

// TestArchiveWorkUsecase_Execute_ReturnsNotFoundForAlreadyArchivedWork verifies archiving
// a work that is not currently published (already archived) is rejected as not found,
// matching the Rails scope Work.without_deleted.published.
//
// [Ja] TestArchiveWorkUsecase_Execute_ReturnsNotFoundForAlreadyArchivedWork は、現在公開中
// でない (すでにアーカイブ済みの) work の非公開が not found として弾かれることを検証する。
// Rails の scope Work.without_deleted.published に一致する。
func TestArchiveWorkUsecase_Execute_ReturnsNotFoundForAlreadyArchivedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "二重非公開_"+t.Name())
	if _, err := uc.Execute(ctx, ArchiveWorkInput{WorkID: workID}); err != nil {
		t.Fatalf("最初の Execute() error = %v", err)
	}

	output, err := uc.Execute(ctx, ArchiveWorkInput{WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v, want nil for an already-archived work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestArchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWork verifies archiving a
// soft-deleted work is rejected as not found (deleted works are outside without_deleted).
//
// [Ja] TestArchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWork は、ソフトデリート済み
// の work の非公開が not found として弾かれることを検証する (削除済みは without_deleted の
// 対象外)。
func TestArchiveWorkUsecase_Execute_ReturnsNotFoundForDeletedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "削除済み非公開_"+t.Name())
	if _, err := db.ExecContext(ctx, "UPDATE works SET deleted_at = NOW() WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("deleted_at の設定に失敗: %v", err)
	}

	output, err := uc.Execute(ctx, ArchiveWorkInput{WorkID: workID})
	if output != nil {
		t.Errorf("output = %+v, want nil for a deleted work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestArchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWork verifies a non-existent
// work id is reported as not found.
//
// [Ja] TestArchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWork は、存在しない work id
// が not found として報告されることを検証する。
func TestArchiveWorkUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveWorkUsecase(db)

	output, err := uc.Execute(context.Background(), ArchiveWorkInput{WorkID: model.WorkID(1 << 62)})
	if output != nil {
		t.Errorf("output = %+v, want nil for a missing work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}
