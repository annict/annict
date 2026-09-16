package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetDBWorkDeletionNewUsecase_Execute_ReturnsDeletableWorkは、削除確認のために
// 公開中のworkもアーカイブ済みのworkも返すことを検証する (削除可能なscope
// Work.without_deletedに一致する)。本UseCaseは読み取りのみでトランザクションを開かないため
// SetupTxを使う。
func TestGetDBWorkDeletionNewUsecase_Execute_ReturnsDeletableWork(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.WorkBuilder) *testutil.WorkBuilder
	}{
		{
			name: "公開中 (unpublished_atなし)",
			prepare: func(b *testutil.WorkBuilder) *testutil.WorkBuilder {
				return b
			},
		},
		{
			name: "アーカイブ済み (unpublished_atあり)",
			prepare: func(b *testutil.WorkBuilder) *testutil.WorkBuilder {
				return b.WithUnpublishedAt(time.Now())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			uc := NewGetDBWorkDeletionNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

			workID := tt.prepare(testutil.NewWorkBuilder(t, tx).WithTitle("削除確認テスト")).Build()

			output, err := uc.Execute(context.Background(), GetDBWorkDeletionNewInput{
				User:   &model.User{ID: 1, Role: model.RoleAdmin},
				WorkID: workID,
			})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.Work == nil {
				t.Fatal("Workがnilだった")
			}
			if output.Work.ID != workID {
				t.Errorf("Work.ID = %d、期待値 = %d", output.Work.ID, workID)
			}
			if output.Work.Title != "削除確認テスト" {
				t.Errorf("Work.Title = %q、期待値 = %q", output.Work.Title, "削除確認テスト")
			}
		})
	}
}

// TestGetDBWorkDeletionNewUsecase_Execute_RejectsDeletedWorkは、すでに削除済みのworkを
// not foundとして報告することを検証する。古い削除リンクが、すでに失われた作品の確認画面では
// なく404になるようにする。
func TestGetDBWorkDeletionNewUsecase_Execute_RejectsDeletedWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkDeletionNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("削除不可テスト").
		WithDeletedAt(time.Now()).
		Build()

	output, err := uc.Execute(context.Background(), GetDBWorkDeletionNewInput{
		User:   &model.User{ID: 1, Role: model.RoleAdmin},
		WorkID: workID,
	})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (削除済みの作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestGetDBWorkDeletionNewUsecase_Execute_ReturnsNotFoundForMissingWorkは、存在しない
// work idがnot foundとして報告されることを検証する。
func TestGetDBWorkDeletionNewUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkDeletionNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

	output, err := uc.Execute(context.Background(), GetDBWorkDeletionNewInput{
		User:   &model.User{ID: 1, Role: model.RoleAdmin},
		WorkID: model.WorkID(1 << 62),
	})
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (存在しない作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestGetDBWorkDeletionNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookupは、
// 認可境界が既存・未存在どちらのworkを取得するより前に、admin以外の全ロールを拒否する
// ことを検証する。
func TestGetDBWorkDeletionNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookup(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkDeletionNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))
	existingWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("認可テスト").Build()
	missingWorkID := model.WorkID(1 << 62)

	tests := []struct {
		name   string
		user   *model.User
		workID model.WorkID
	}{
		{name: "未認証・既存作品", user: nil, workID: existingWorkID},
		{name: "未認証・未存在作品", user: nil, workID: missingWorkID},
		{name: "一般ユーザー・既存作品", user: &model.User{ID: 1, Role: model.RoleUser}, workID: existingWorkID},
		{name: "一般ユーザー・未存在作品", user: &model.User{ID: 1, Role: model.RoleUser}, workID: missingWorkID},
		{name: "編集者・既存作品", user: &model.User{ID: 1, Role: model.RoleEditor}, workID: existingWorkID},
		{name: "編集者・未存在作品", user: &model.User{ID: 1, Role: model.RoleEditor}, workID: missingWorkID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBWorkDeletionNewInput{
				User:   tt.user,
				WorkID: tt.workID,
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
