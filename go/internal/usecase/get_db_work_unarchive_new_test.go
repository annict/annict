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

// TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsArchivedWork verifies the usecase returns
// a currently archived work for the publish confirmation. It is a read-only usecase that
// opens no transaction, so the test uses SetupTx.
//
// [Ja] TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsArchivedWork は、公開確認のために
// 現在アーカイブ済みの work を返すことを検証する。本 UseCase は読み取りのみでトランザクションを
// 開かないため SetupTx を使う。
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
				t.Fatalf("Execute() error = %v", err)
			}
			if output.Work == nil {
				t.Fatal("Work should not be nil")
			}
			if output.Work.ID != workID {
				t.Errorf("Work.ID = %d, want %d", output.Work.ID, workID)
			}
			if output.Work.Title != "公開確認テスト" {
				t.Errorf("Work.Title = %q, want %q", output.Work.Title, "公開確認テスト")
			}
		})
	}
}

// TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsNonPublishableWork verifies the usecase
// reports a work that is not currently archived (still published, or deleted) as not found,
// matching the publishable scope Work.without_deleted.unpublished.
//
// [Ja] TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsNonPublishableWork は、現在アーカイブ
// 済みでない (公開中、または削除済みの) work を not found として報告することを検証する。
// 公開可能な scope Work.without_deleted.unpublished に一致する。
func TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsNonPublishableWork(t *testing.T) {
	t.Parallel()

	now := time.Now()

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
			name: "削除済み (deleted_at あり)",
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
				t.Errorf("output = %+v, want nil for a non-publishable work", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
			}
		})
	}
}

// TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsNotFoundForMissingWork verifies a
// non-existent work id is reported as not found.
//
// [Ja] TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsNotFoundForMissingWork は、存在しない
// work id が not found として報告されることを検証する。
func TestGetDBWorkUnarchiveNewUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkUnarchiveNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

	output, err := uc.Execute(context.Background(), GetDBWorkUnarchiveNewInput{
		User:   &model.User{ID: 1, Role: model.RoleEditor},
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

// TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookup verifies the
// authorization boundary rejects unauthenticated and regular users before looking up either
// an existing or a missing work.
//
// [Ja] TestGetDBWorkUnarchiveNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookup は、
// 認可境界が既存・未存在どちらの work を取得するより前に、未認証と一般ユーザーを拒否する
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
				t.Errorf("output = %+v, want nil for an unauthorized user", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("expected AppErrCodeForbidden, got %v", err)
			}
		})
	}
}
