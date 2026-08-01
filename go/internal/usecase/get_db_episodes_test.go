package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newGetDBEpisodesUsecase wires the usecase against the test transaction. It is a read-only
// usecase that opens no transaction of its own, so the test uses SetupTx.
//
// [Ja] newGetDBEpisodesUsecase はテスト用トランザクション上に UseCase を組み立てる。本
// UseCase は読み取りのみで自らトランザクションを開かないため SetupTx を使う。
func newGetDBEpisodesUsecase(t *testing.T) (*GetDBEpisodesUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	return NewGetDBEpisodesUsecase(
		repository.NewWorkRepository(queries),
		repository.NewEpisodeRepository(queries),
	), tx
}

// TestGetDBEpisodesUsecase_Execute_ReturnsWorkAndEpisodes verifies the usecase returns the
// parent work together with its episodes, ordered as the list renders them (sort_number
// descending), and the total count the pagination needs.
//
// [Ja] TestGetDBEpisodesUsecase_Execute_ReturnsWorkAndEpisodes は、親作品とそのエピソードを
// 一覧が描画する順 (sort_number 降順) で返し、ページネーションに必要な総件数も返すことを
// 検証する。
func TestGetDBEpisodesUsecase_Execute_ReturnsWorkAndEpisodes(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodesUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソード一覧テスト").Build()
	firstID := insertEpisodeWithSortNumber(t, tx, workID, "第1話", 100)
	secondID := insertEpisodeWithSortNumber(t, tx, workID, "第2話", 200)

	output, err := uc.Execute(context.Background(), GetDBEpisodesInput{WorkID: workID, Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Work == nil {
		t.Fatal("Work should not be nil")
	}
	if output.Work.Title != "エピソード一覧テスト" {
		t.Errorf("Work.Title = %q, want %q", output.Work.Title, "エピソード一覧テスト")
	}
	if output.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", output.TotalCount)
	}

	wantIDs := []model.EpisodeID{secondID, firstID}
	if len(output.Episodes) != len(wantIDs) {
		t.Fatalf("len(Episodes) = %d, want %d", len(output.Episodes), len(wantIDs))
	}
	for i, want := range wantIDs {
		if output.Episodes[i].ID != want {
			t.Errorf("Episodes[%d].ID = %d, want %d", i, output.Episodes[i].ID, want)
		}
	}
}

// TestGetDBEpisodesUsecase_Execute_PaginatesEpisodes verifies that PerPage / Page select one
// page of episodes while TotalCount keeps reporting every listed episode, so the pagination
// can render the remaining pages.
//
// [Ja] TestGetDBEpisodesUsecase_Execute_PaginatesEpisodes は PerPage / Page がエピソードを
// 1 ページ分だけ切り出す一方、TotalCount は一覧対象すべてを報告し続けることを検証する。
// これによりページネーションが残りのページを描画できる。
func TestGetDBEpisodesUsecase_Execute_PaginatesEpisodes(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodesUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("ページングテスト").Build()
	insertEpisodeWithSortNumber(t, tx, workID, "第1話", 100)
	secondID := insertEpisodeWithSortNumber(t, tx, workID, "第2話", 200)

	output, err := uc.Execute(context.Background(), GetDBEpisodesInput{WorkID: workID, Page: 1, PerPage: 1})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(output.Episodes) != 1 {
		t.Fatalf("len(Episodes) = %d, want 1", len(output.Episodes))
	}
	if output.Episodes[0].ID != secondID {
		t.Errorf("Episodes[0].ID = %d, want %d", output.Episodes[0].ID, secondID)
	}
	if output.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", output.TotalCount)
	}
}

// TestGetDBEpisodesUsecase_Execute_ReturnsNotFoundForMissingWork verifies that a
// non-existent or deleted work is reported as not found, matching the Rails
// Work.without_deleted.find the episode list uses.
//
// [Ja] TestGetDBEpisodesUsecase_Execute_ReturnsNotFoundForMissingWork は、存在しない作品と
// 削除済みの作品が not found として報告されることを検証する。エピソード一覧が使う Rails の
// Work.without_deleted.find に一致する。
func TestGetDBEpisodesUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	t.Run("存在しない作品", func(t *testing.T) {
		t.Parallel()

		uc, _ := newGetDBEpisodesUsecase(t)

		output, err := uc.Execute(context.Background(), GetDBEpisodesInput{WorkID: model.WorkID(1 << 62), Page: 1, PerPage: 100})
		if output != nil {
			t.Errorf("output = %+v, want nil for a missing work", output)
		}
		ae := model.AsAppError(err)
		if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
			t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
		}
	})

	t.Run("削除済みの作品", func(t *testing.T) {
		t.Parallel()

		uc, tx := newGetDBEpisodesUsecase(t)

		workID := testutil.NewWorkBuilder(t, tx).
			WithTitle("削除済みテスト").
			WithDeletedAt(time.Now()).
			Build()

		output, err := uc.Execute(context.Background(), GetDBEpisodesInput{WorkID: workID, Page: 1, PerPage: 100})
		if output != nil {
			t.Errorf("output = %+v, want nil for a deleted work", output)
		}
		ae := model.AsAppError(err)
		if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
			t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
		}
	})
}

// insertEpisodeWithSortNumber creates an episode with an explicit sort_number, which decides
// the list order. The shared EpisodeBuilder fixes sort_number, so the ordering assertions
// insert their rows directly.
//
// [Ja] insertEpisodeWithSortNumber は sort_number を明示してエピソードを作成する。
// sort_number は一覧の並び順を決める。共有の EpisodeBuilder は sort_number を固定するため、
// 並び順を検証するテストは行を直接挿入する。
func insertEpisodeWithSortNumber(t *testing.T, tx *sql.Tx, workID model.WorkID, number string, sortNumber int32) model.EpisodeID {
	t.Helper()

	var id int64
	err := tx.QueryRow(`
		INSERT INTO episodes (work_id, number, sort_number, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`, int64(workID), number, sortNumber).Scan(&id)
	if err != nil {
		t.Fatalf("エピソードの作成に失敗: %v", err)
	}

	return model.EpisodeID(id)
}
