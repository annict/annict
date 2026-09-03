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

// TestGetDBWorkDeletionNewUsecase_Execute_ReturnsDeletableWork verifies the usecase returns
// both a published and an archived work for the delete confirmation, matching the deletable
// scope Work.without_deleted. It is a read-only usecase that opens no transaction, so the
// test uses SetupTx.
//
// [Ja] TestGetDBWorkDeletionNewUsecase_Execute_ReturnsDeletableWork は、削除確認のために
// 公開中の work もアーカイブ済みの work も返すことを検証する (削除可能な scope
// Work.without_deleted に一致する)。本 UseCase は読み取りのみでトランザクションを開かないため
// SetupTx を使う。
func TestGetDBWorkDeletionNewUsecase_Execute_ReturnsDeletableWork(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.WorkBuilder) *testutil.WorkBuilder
	}{
		{
			name: "公開中 (unpublished_at なし)",
			prepare: func(b *testutil.WorkBuilder) *testutil.WorkBuilder {
				return b
			},
		},
		{
			name: "アーカイブ済み (unpublished_at あり)",
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
				t.Fatalf("Execute() error = %v", err)
			}
			if output.Work == nil {
				t.Fatal("Work should not be nil")
			}
			if output.Work.ID != workID {
				t.Errorf("Work.ID = %d, want %d", output.Work.ID, workID)
			}
			if output.Work.Title != "削除確認テスト" {
				t.Errorf("Work.Title = %q, want %q", output.Work.Title, "削除確認テスト")
			}
		})
	}
}

// TestGetDBWorkDeletionNewUsecase_Execute_RejectsDeletedWork verifies the usecase reports an
// already deleted work as not found, so a stale delete link turns into a 404 rather than a
// confirmation for a work that is already gone.
//
// [Ja] TestGetDBWorkDeletionNewUsecase_Execute_RejectsDeletedWork は、すでに削除済みの work を
// not found として報告することを検証する。古い削除リンクが、すでに失われた作品の確認画面では
// なく 404 になるようにする。
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
		t.Errorf("output = %+v, want nil for a deleted work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestGetDBWorkDeletionNewUsecase_Execute_ReturnsNotFoundForMissingWork verifies a
// non-existent work id is reported as not found.
//
// [Ja] TestGetDBWorkDeletionNewUsecase_Execute_ReturnsNotFoundForMissingWork は、存在しない
// work id が not found として報告されることを検証する。
func TestGetDBWorkDeletionNewUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkDeletionNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

	output, err := uc.Execute(context.Background(), GetDBWorkDeletionNewInput{
		User:   &model.User{ID: 1, Role: model.RoleAdmin},
		WorkID: model.WorkID(1 << 62),
	})
	if output != nil {
		t.Errorf("output = %+v, want nil for a missing work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}

// TestGetDBWorkDeletionNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookup verifies the
// authorization boundary rejects every non-admin role before looking up either an existing
// or a missing work.
//
// [Ja] TestGetDBWorkDeletionNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookup は、
// 認可境界が既存・未存在どちらの work を取得するより前に、admin 以外の全ロールを拒否する
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
				t.Errorf("output = %+v, want nil for an unauthorized user", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("expected AppErrCodeForbidden, got %v", err)
			}
		})
	}
}
