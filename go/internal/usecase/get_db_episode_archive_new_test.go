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

// newGetDBEpisodeArchiveNewUsecase wires the usecase against the test transaction. It is a
// read-only usecase that opens no transaction of its own, so the test uses SetupTx.
//
// [Ja] newGetDBEpisodeArchiveNewUsecase はテスト用トランザクション上に UseCase を組み立てる。本
// UseCase は読み取りのみで自らトランザクションを開かないため SetupTx を使う。
func newGetDBEpisodeArchiveNewUsecase(t *testing.T) (*GetDBEpisodeArchiveNewUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)

	return NewGetDBEpisodeArchiveNewUsecase(repository.NewEpisodeRepository(query.New(db).WithTx(tx))), tx
}

// TestGetDBEpisodeArchiveNewUsecase_Execute_ReturnsEpisodeAndWork verifies the usecase returns
// the episode the confirmation names together with the parent work its heading and subnav
// describe.
//
// [Ja] TestGetDBEpisodeArchiveNewUsecase_Execute_ReturnsEpisodeAndWork は、確認が名指しする
// エピソードと、その見出しとサブナビが示す親作品を返すことを検証する。
func TestGetDBEpisodeArchiveNewUsecase_Execute_ReturnsEpisodeAndWork(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodeArchiveNewUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソード非公開テスト").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第2話").WithTitle("二話目").Build()

	output, err := uc.Execute(context.Background(), GetDBEpisodeArchiveNewInput{EpisodeID: episodeID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Episode == nil || output.Episode.ID != episodeID {
		t.Fatalf("Episode = %+v, want ID %d", output.Episode, int64(episodeID))
	}
	if output.Episode.Number == nil || *output.Episode.Number != "第2話" {
		t.Errorf("Episode.Number = %v, want %q", output.Episode.Number, "第2話")
	}
	if output.Episode.Title == nil || *output.Episode.Title != "二話目" {
		t.Errorf("Episode.Title = %v, want %q", output.Episode.Title, "二話目")
	}
	if output.Work == nil || output.Work.ID != workID {
		t.Fatalf("Work = %+v, want ID %d", output.Work, int64(workID))
	}
	if output.Work.Title != "エピソード非公開テスト" {
		t.Errorf("Work.Title = %q, want %q", output.Work.Title, "エピソード非公開テスト")
	}
}

// TestGetDBEpisodeArchiveNewUsecase_Execute_RejectsNonArchivableEpisode verifies the confirmation
// page cannot be shown for an episode that is not currently published (already archived, or
// deleted), for one whose work was deleted, or for one that does not exist. The submit rejects
// the same set, so the two never disagree about which episodes are archivable.
//
// [Ja] TestGetDBEpisodeArchiveNewUsecase_Execute_RejectsNonArchivableEpisode は、現在公開中でない
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
				t.Fatalf("Execute() error = %v, want AppErrCodeResourceNotFound", err)
			}
		})
	}
}
