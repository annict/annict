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

// newGetDBEpisodesUsecaseはテスト用トランザクション上にUseCaseを組み立てる。本
// UseCaseは読み取りのみで自らトランザクションを開かないためSetupTxを使う。
func newGetDBEpisodesUsecase(t *testing.T) (*GetDBEpisodesUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	return NewGetDBEpisodesUsecase(
		repository.NewWorkRepository(queries),
		repository.NewEpisodeRepository(queries),
	), tx
}

// TestGetDBEpisodesUsecase_Execute_ReturnsWorkAndEpisodesは、親作品とそのエピソードを
// 一覧が描画する順 (sort_number降順) で返し、ページネーションに必要な総件数も返すことを
// 検証する。
func TestGetDBEpisodesUsecase_Execute_ReturnsWorkAndEpisodes(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodesUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソード一覧テスト").Build()
	firstID := insertEpisodeWithSortNumber(t, tx, workID, "第1話", 100)
	secondID := insertEpisodeWithSortNumber(t, tx, workID, "第2話", 200)

	output, err := uc.Execute(context.Background(), GetDBEpisodesInput{WorkID: workID, Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.Work == nil {
		t.Fatal("Workがnilだった")
	}
	if output.Work.Title != "エピソード一覧テスト" {
		t.Errorf("Work.Title = %q、期待値 = %q", output.Work.Title, "エピソード一覧テスト")
	}
	if output.TotalCount != 2 {
		t.Errorf("TotalCount = %d、期待値 = 2", output.TotalCount)
	}

	wantIDs := []model.EpisodeID{secondID, firstID}
	if len(output.Episodes) != len(wantIDs) {
		t.Fatalf("len(Episodes) = %d、期待値 = %d", len(output.Episodes), len(wantIDs))
	}
	for i, want := range wantIDs {
		if output.Episodes[i].ID != want {
			t.Errorf("Episodes[%d].ID = %d、期待値 = %d", i, output.Episodes[i].ID, want)
		}
	}
}

// TestGetDBEpisodesUsecase_Execute_ReturnsGenerationValuesは、ページの自動生成の
// 案内が報告する3つの値をUseCaseが運ぶこと、およびそれが一覧自体の総件数とは別物で
// あることを検証する。
// 案内は公開中のエピソードだけを数え、一覧は表示する非公開のエピソードも総件数に含める。
func TestGetDBEpisodesUsecase_Execute_ReturnsGenerationValues(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodesUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("自動生成の案内テスト").
		WithManualEpisodesCount(12).
		Build()
	insertEpisodeWithSortNumber(t, tx, workID, "第1話", 100)
	testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("第2話").
		WithUnpublishedAt(time.Now()).
		Build()

	channelID := testutil.NewChannelBuilder(t, tx).Build()
	testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).WithNumber(9).Build()

	output, err := uc.Execute(context.Background(), GetDBEpisodesInput{WorkID: workID, Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.Work.ManualEpisodesCount == nil || *output.Work.ManualEpisodesCount != 12 {
		t.Errorf("Work.ManualEpisodesCount = %v、期待値 = 12", output.Work.ManualEpisodesCount)
	}
	if output.PublishedEpisodeCount != 1 {
		t.Errorf("PublishedEpisodeCount = %d、期待値 = 1", output.PublishedEpisodeCount)
	}
	if output.MaxGeneratableEpisodeNumber != 9 {
		t.Errorf("MaxGeneratableEpisodeNumber = %d、期待値 = 9", output.MaxGeneratableEpisodeNumber)
	}
	if output.TotalCount != 2 {
		t.Errorf("TotalCount = %d、期待値 = 2 (一覧は非公開のエピソードも含む)", output.TotalCount)
	}
}

// TestGetDBEpisodesUsecase_Execute_PaginatesEpisodesはPerPage / Pageがエピソードを
// 1ページ分だけ切り出す一方、TotalCountは一覧対象すべてを報告し続けることを検証する。
// これによりページネーションが残りのページを描画できる。
func TestGetDBEpisodesUsecase_Execute_PaginatesEpisodes(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodesUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("ページングテスト").Build()
	insertEpisodeWithSortNumber(t, tx, workID, "第1話", 100)
	secondID := insertEpisodeWithSortNumber(t, tx, workID, "第2話", 200)

	output, err := uc.Execute(context.Background(), GetDBEpisodesInput{WorkID: workID, Page: 1, PerPage: 1})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if len(output.Episodes) != 1 {
		t.Fatalf("len(Episodes) = %d、期待値 = 1", len(output.Episodes))
	}
	if output.Episodes[0].ID != secondID {
		t.Errorf("Episodes[0].ID = %d、期待値 = %d", output.Episodes[0].ID, secondID)
	}
	if output.TotalCount != 2 {
		t.Errorf("TotalCount = %d、期待値 = 2", output.TotalCount)
	}
}

// TestGetDBEpisodesUsecase_Execute_ReturnsNotFoundForMissingWorkは、存在しない作品と
// 削除済みの作品がnot foundとして報告されることを検証する。エピソード一覧が使うRailsの
// Work.without_deleted.findに一致する。
func TestGetDBEpisodesUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	t.Run("存在しない作品", func(t *testing.T) {
		t.Parallel()

		uc, _ := newGetDBEpisodesUsecase(t)

		output, err := uc.Execute(context.Background(), GetDBEpisodesInput{WorkID: model.WorkID(1 << 62), Page: 1, PerPage: 100})
		if output != nil {
			t.Errorf("output = %+v、期待値 = nil (存在しない作品のため)", output)
		}
		ae := model.AsAppError(err)
		if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
			t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
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
			t.Errorf("output = %+v、期待値 = nil (削除済みの作品のため)", output)
		}
		ae := model.AsAppError(err)
		if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
			t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
		}
	})
}

// insertEpisodeWithSortNumberはsort_numberを明示してエピソードを作成する。
// sort_numberは一覧の並び順を決める。共有のEpisodeBuilderはsort_numberを固定するため、
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
