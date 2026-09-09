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

// newGetDBEpisodeEditUsecase wires the usecase against the test transaction. It is a
// read-only usecase that opens no transaction of its own, so the test uses SetupTx.
//
// [Ja] newGetDBEpisodeEditUsecase はテスト用トランザクション上に UseCase を組み立てる。本
// UseCase は読み取りのみで自らトランザクションを開かないため SetupTx を使う。
func newGetDBEpisodeEditUsecase(t *testing.T) (*GetDBEpisodeEditUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)

	return NewGetDBEpisodeEditUsecase(repository.NewEpisodeRepository(query.New(db).WithTx(tx))), tx
}

// TestGetDBEpisodeEditUsecase_Execute_ReturnsEpisodeAndWork verifies the usecase returns the
// episode the form starts from together with the parent work its heading and subnav describe.
//
// [Ja] TestGetDBEpisodeEditUsecase_Execute_ReturnsEpisodeAndWork は、フォームの初期値になる
// エピソードと、その見出しとサブナビが示す親作品を返すことを検証する。
func TestGetDBEpisodeEditUsecase_Execute_ReturnsEpisodeAndWork(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodeEditUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソード編集テスト").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第2話").WithTitle("二話目").Build()

	output, err := uc.Execute(context.Background(), GetDBEpisodeEditInput{EpisodeID: episodeID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Episode == nil {
		t.Fatal("Episode should not be nil")
	}
	if output.Episode.ID != episodeID {
		t.Errorf("Episode.ID = %d, want %d", int64(output.Episode.ID), int64(episodeID))
	}
	if output.Episode.Number == nil || *output.Episode.Number != "第2話" {
		t.Errorf("Episode.Number = %v, want %q", output.Episode.Number, "第2話")
	}
	if output.Work == nil {
		t.Fatal("Work should not be nil")
	}
	if output.Work.ID != workID {
		t.Errorf("Work.ID = %d, want %d", int64(output.Work.ID), int64(workID))
	}
	if output.Work.Title != "エピソード編集テスト" {
		t.Errorf("Work.Title = %q, want %q", output.Work.Title, "エピソード編集テスト")
	}
}

// TestGetDBEpisodeEditUsecase_Execute_NotFound verifies every episode the edit form cannot be
// opened for is reported as a not-found AppError, which the handler renders as 404.
//
// [Ja] TestGetDBEpisodeEditUsecase_Execute_NotFound は、編集フォームを開けないエピソードが
// いずれも not found の AppError として報告されることを検証する (Handler はこれを 404 として
// 描画する)。
func TestGetDBEpisodeEditUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodeEditUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("編集不可テスト").Build()
	deletedEpisodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithDeletedAt(time.Now()).Build()

	deletedWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("削除済み作品").WithDeletedAt(time.Now()).Build()
	episodeOfDeletedWorkID := testutil.NewEpisodeBuilder(t, tx, deletedWorkID).Build()

	tests := []struct {
		name      string
		episodeID model.EpisodeID
	}{
		{name: "存在しないエピソード", episodeID: model.EpisodeID(999999999)},
		{name: "削除済みのエピソード", episodeID: deletedEpisodeID},
		{name: "削除済み作品のエピソード", episodeID: episodeOfDeletedWorkID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBEpisodeEditInput{EpisodeID: tt.episodeID})
			if output != nil {
				t.Error("output should be nil")
			}
			ae := model.AsAppError(err)
			if ae == nil {
				t.Fatalf("AppError を期待したが %v", err)
			}
			if ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("Code = %v, want %v", ae.Code, model.AppErrCodeResourceNotFound)
			}
		})
	}
}
