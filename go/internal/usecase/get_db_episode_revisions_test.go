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

func newGetDBEpisodeRevisionsUsecase(t *testing.T) (*GetDBEpisodeRevisionsUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	return NewGetDBEpisodeRevisionsUsecase(repository.NewEpisodeRepository(queries), repository.NewAnimeRevisionRepository(queries)), tx
}

func TestGetDBEpisodeRevisionsUsecase_Execute_ReturnsRevisions(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodeRevisionsUsecase(t)
	ctx := context.Background()
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("履歴の親作品").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第1話").Build()
	animeID := testutil.CreateTestAnime(t, tx)
	if _, err := tx.Exec(`UPDATE episodes SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(episodeID)); err != nil {
		t.Fatalf("episodes.anime_idの設定に失敗: %v", err)
	}
	title := "第1話のタイトル"
	if err := repository.NewAnimeRevisionRepository(query.New(tx)).Create(ctx, testutil.NewDBEditBuilder(t, tx).Build(), animeID, model.AnimeSnapshot{
		Title:  &title,
		Status: model.AnimeStatusPublished,
		Kind:   model.AnimeClassificationKindEpisode,
	}); err != nil {
		t.Fatalf("版の作成に失敗: %v", err)
	}

	output, err := uc.Execute(ctx, GetDBEpisodeRevisionsInput{EpisodeID: episodeID, Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.Episode.ID != episodeID || output.Work.ID != workID || output.Work.Title != "履歴の親作品" {
		t.Errorf("エピソードと作品 = %d・%+v、期待値 = %d・作品%d", output.Episode.ID, output.Work, episodeID, workID)
	}
	if output.TotalCount != 1 || len(output.Revisions) != 1 || *output.Revisions[0].Snapshot.Title != title {
		t.Errorf("版 = %d件 (総数%d)、期待値 = %qの1件", len(output.Revisions), output.TotalCount, title)
	}
}

func TestGetDBEpisodeRevisionsUsecase_Execute_EmptyForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodeRevisionsUsecase(t)
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()

	output, err := uc.Execute(context.Background(), GetDBEpisodeRevisionsInput{EpisodeID: episodeID, Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.TotalCount != 0 || output.Revisions == nil || len(output.Revisions) != 0 {
		t.Errorf("版 = %v (総数%d)、期待値 = 空の一覧", output.Revisions, output.TotalCount)
	}
}

func TestGetDBEpisodeRevisionsUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		target func(t *testing.T, tx *sql.Tx) model.EpisodeID
	}{
		{name: "存在しないエピソード", target: func(t *testing.T, tx *sql.Tx) model.EpisodeID { return model.EpisodeID(1 << 62) }},
		{name: "削除したエピソード", target: func(t *testing.T, tx *sql.Tx) model.EpisodeID {
			workID := testutil.NewWorkBuilder(t, tx).Build()
			return testutil.NewEpisodeBuilder(t, tx, workID).WithDeletedAt(time.Now()).Build()
		}},
		{name: "削除した作品のエピソード", target: func(t *testing.T, tx *sql.Tx) model.EpisodeID {
			workID := testutil.NewWorkBuilder(t, tx).WithDeletedAt(time.Now()).Build()
			return testutil.NewEpisodeBuilder(t, tx, workID).Build()
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, tx := newGetDBEpisodeRevisionsUsecase(t)
			_, err := uc.Execute(context.Background(), GetDBEpisodeRevisionsInput{EpisodeID: tt.target(t, tx), Page: 1, PerPage: 10})
			if ae := model.AsAppError(err); ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}
