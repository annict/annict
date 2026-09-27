package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsArchivedWorkは、公開確認のために
// 現在アーカイブ済みのworkを返すことを検証する。本UseCaseは読み取りのみでトランザクションを
// 開かないためSetupTxを使う。
func TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsArchivedWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkUnarchiveNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("公開確認テスト").
		WithUnpublishedAt(time.Now()).
		Build()

	for _, role := range []int32{model.RoleAdmin, model.RoleEditor} {
		t.Run(fmt.Sprintf("role=%d", role), func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBWorkUnarchiveNewInput{
				User:   &model.User{ID: 1, Role: role},
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
			if output.Work.Title != "公開確認テスト" {
				t.Errorf("Work.Title = %q、期待値 = %q", output.Work.Title, "公開確認テスト")
			}
		})
	}
}

// TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsNonPublishableWorkは、現在アーカイブ
// 済みでない (公開中、または削除済みの) workをnot foundとして報告することを検証する。
// 公開可能なscope Work.without_deleted.unpublishedに一致する。
func TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsNonPublishableWork(t *testing.T) {
	t.Parallel()

	now := time.Now()

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
			name: "削除済み (deleted_atあり)",
			prepare: func(b *testutil.WorkBuilder) *testutil.WorkBuilder {
				return b.WithUnpublishedAt(now).WithDeletedAt(now)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			uc := NewGetDBWorkUnarchiveNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

			workID := tt.prepare(testutil.NewWorkBuilder(t, tx).WithTitle("公開不可テスト")).Build()

			output, err := uc.Execute(context.Background(), GetDBWorkUnarchiveNewInput{
				User:   &model.User{ID: 1, Role: model.RoleEditor},
				WorkID: workID,
			})
			if output != nil {
				t.Errorf("output = %+v、期待値 = nil (公開できない作品のため)", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}

// TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsNotFoundForMissingWorkは、存在しない
// work idがnot foundとして報告されることを検証する。
func TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkUnarchiveNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

	output, err := uc.Execute(context.Background(), GetDBWorkUnarchiveNewInput{
		User:   &model.User{ID: 1, Role: model.RoleEditor},
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

// TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookupは、
// 認可境界が既存・未存在どちらのworkを取得するより前に、未認証と一般ユーザーを拒否する
// ことを検証する。
func TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookup(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkUnarchiveNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))
	existingWorkID := testutil.NewWorkBuilder(t, tx).
		WithTitle("認可テスト").
		WithUnpublishedAt(time.Now()).
		Build()
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBWorkUnarchiveNewInput{
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
