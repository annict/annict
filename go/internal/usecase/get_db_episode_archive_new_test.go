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

// newGetDBEpisodeArchiveNewUsecaseはテスト用トランザクション上にUseCaseを組み立てる。本
// UseCaseは読み取りのみで自らトランザクションを開かないためSetupTxを使う。
func newGetDBEpisodeArchiveNewUsecase(t *testing.T) (*GetDBEpisodeArchiveNewUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)

	return NewGetDBEpisodeArchiveNewUsecase(repository.NewEpisodeRepository(query.New(db).WithTx(tx))), tx
}

// TestGetDBEpisodeArchiveNewUsecase_Execute_ReturnsEpisodeAndWorkは、確認が名指しする
// エピソードと、その見出しとサブナビが示す親作品を返すことを検証する。
func TestGetDBEpisodeArchiveNewUsecase_Execute_ReturnsEpisodeAndWork(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodeArchiveNewUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソード非公開テスト").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第2話").WithTitle("二話目").Build()

	output, err := uc.Execute(context.Background(), GetDBEpisodeArchiveNewInput{EpisodeID: episodeID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.Episode == nil || output.Episode.ID != episodeID {
		t.Fatalf("Episode = %+v、期待値 = ID %d", output.Episode, int64(episodeID))
	}
	if output.Episode.Number == nil || *output.Episode.Number != "第2話" {
		t.Errorf("Episode.Number = %v、期待値 = %q", output.Episode.Number, "第2話")
	}
	if output.Episode.Title == nil || *output.Episode.Title != "二話目" {
		t.Errorf("Episode.Title = %v、期待値 = %q", output.Episode.Title, "二話目")
	}
	if output.Work == nil || output.Work.ID != workID {
		t.Fatalf("Work = %+v、期待値 = ID %d", output.Work, int64(workID))
	}
	if output.Work.Title != "エピソード非公開テスト" {
		t.Errorf("Work.Title = %q、期待値 = %q", output.Work.Title, "エピソード非公開テスト")
	}
}

// TestGetDBEpisodeArchiveNewUsecase_Execute_RejectsNonArchivableEpisodeは、現在公開中でない
// (すでに非公開、または削除済みの) エピソード、作品が削除されたエピソード、存在しないエピソード
// に対して確認ページを出せないことを検証する。送信も同じ集合を拒否するため、非公開にできる
// エピソードの判断が両者でずれない。
func TestGetDBEpisodeArchiveNewUsecase_Execute_RejectsNonArchivableEpisode(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := map[string]func(t *testing.T, tx *sql.Tx) model.EpisodeID{
		"非公開済みのエピソード": func(t *testing.T, tx *sql.Tx) model.EpisodeID {
			workID := testutil.NewWorkBuilder(t, tx).Build()
			return testutil.NewEpisodeBuilder(t, tx, workID).WithUnpublishedAt(now).Build()
		},
		"削除済みのエピソード": func(t *testing.T, tx *sql.Tx) model.EpisodeID {
			workID := testutil.NewWorkBuilder(t, tx).Build()
			return testutil.NewEpisodeBuilder(t, tx, workID).WithDeletedAt(now).Build()
		},
		"削除済み作品のエピソード": func(t *testing.T, tx *sql.Tx) model.EpisodeID {
			workID := testutil.NewWorkBuilder(t, tx).WithDeletedAt(now).Build()
			return testutil.NewEpisodeBuilder(t, tx, workID).Build()
		},
		"存在しないエピソード": func(t *testing.T, tx *sql.Tx) model.EpisodeID {
			return model.EpisodeID(-1)
		},
	}

	for name, prepare := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uc, tx := newGetDBEpisodeArchiveNewUsecase(t)
			episodeID := prepare(t, tx)

			_, err := uc.Execute(context.Background(), GetDBEpisodeArchiveNewInput{EpisodeID: episodeID})
			appErr := model.AsAppError(err)
			if appErr == nil || appErr.Code != model.AppErrCodeResourceNotFound {
				t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}
