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

// newGetDBEpisodeEditUsecaseはテスト用トランザクション上にUseCaseを組み立てる。本
// UseCaseは読み取りのみで自らトランザクションを開かないためSetupTxを使う。
func newGetDBEpisodeEditUsecase(t *testing.T) (*GetDBEpisodeEditUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)

	return NewGetDBEpisodeEditUsecase(repository.NewEpisodeRepository(query.New(db).WithTx(tx))), tx
}

// TestGetDBEpisodeEditUsecase_Execute_ReturnsEpisodeAndWorkは、フォームの初期値になる
// エピソードと、その見出しとサブナビが示す親作品を返すことを検証する。
func TestGetDBEpisodeEditUsecase_Execute_ReturnsEpisodeAndWork(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodeEditUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソード編集テスト").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第2話").WithTitle("二話目").Build()

	output, err := uc.Execute(context.Background(), GetDBEpisodeEditInput{User: &model.User{ID: 1, Role: model.RoleEditor}, EpisodeID: episodeID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.Episode == nil {
		t.Fatal("Episodeがnilだった")
	}
	if output.Episode.ID != episodeID {
		t.Errorf("Episode.ID = %d、期待値 = %d", int64(output.Episode.ID), int64(episodeID))
	}
	if output.Episode.Number == nil || *output.Episode.Number != "第2話" {
		t.Errorf("Episode.Number = %v、期待値 = %q", output.Episode.Number, "第2話")
	}
	if output.Work == nil {
		t.Fatal("Workがnilだった")
	}
	if output.Work.ID != workID {
		t.Errorf("Work.ID = %d、期待値 = %d", int64(output.Work.ID), int64(workID))
	}
	if output.Work.Title != "エピソード編集テスト" {
		t.Errorf("Work.Title = %q、期待値 = %q", output.Work.Title, "エピソード編集テスト")
	}
}

// TestGetDBEpisodeEditUsecase_Execute_NotFoundは、編集フォームを開けないエピソードが
// いずれもnot foundのAppErrorとして報告されることを検証する (Handlerはこれを404として
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
			output, err := uc.Execute(context.Background(), GetDBEpisodeEditInput{User: &model.User{ID: 1, Role: model.RoleEditor}, EpisodeID: tt.episodeID})
			if output != nil {
				t.Error("outputがnilでなかった")
			}
			ae := model.AsAppError(err)
			if ae == nil {
				t.Fatalf("AppErrorを期待したが%v", err)
			}
			if ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("Code = %v、期待値 = %v", ae.Code, model.AppErrCodeResourceNotFound)
			}
		})
	}
}

// TestGetDBEpisodeEditUsecase_Execute_RejectsUnauthorizedUserBeforeLookupは、認可境界が
// 既存・未存在どちらのエピソードを取得するより前に、未認証と一般ユーザーを拒否することを
// 検証する。
func TestGetDBEpisodeEditUsecase_Execute_RejectsUnauthorizedUserBeforeLookup(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBEpisodeEditUsecase(t)

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソード編集認可テスト").Build()
	existingEpisodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	missingEpisodeID := model.EpisodeID(1 << 62)

	tests := []struct {
		name      string
		user      *model.User
		episodeID model.EpisodeID
	}{
		{name: "未認証・既存エピソード", user: nil, episodeID: existingEpisodeID},
		{name: "未認証・未存在エピソード", user: nil, episodeID: missingEpisodeID},
		{name: "一般ユーザー・既存エピソード", user: &model.User{ID: 1, Role: model.RoleUser}, episodeID: existingEpisodeID},
		{name: "一般ユーザー・未存在エピソード", user: &model.User{ID: 1, Role: model.RoleUser}, episodeID: missingEpisodeID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBEpisodeEditInput{
				User:      tt.user,
				EpisodeID: tt.episodeID,
			})
			if output != nil {
				t.Errorf("output = %+v、期待値 = nil (権限の無いユーザーのため)", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("エラーコード = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}
}
