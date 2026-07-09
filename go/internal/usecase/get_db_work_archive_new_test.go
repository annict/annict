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

// TestGetDBWorkArchiveNewUsecase_Execute_ReturnsPublishedWork verifies the usecase returns
// a currently published work for the archive confirmation. It is a read-only usecase that
// opens no transaction, so the test uses SetupTx.
//
// [Ja] TestGetDBWorkArchiveNewUsecase_Execute_ReturnsPublishedWork は、非公開確認のために
// 現在公開中の work を返すことを検証する。本 UseCase は読み取りのみでトランザクションを
// 開かないため SetupTx を使う。
func TestGetDBWorkArchiveNewUsecase_Execute_ReturnsPublishedWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkArchiveNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("非公開確認テスト").
		Build()

	output, err := uc.Execute(context.Background(), GetDBWorkArchiveNewInput{WorkID: workID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Work == nil {
		t.Fatal("Work should not be nil")
	}
	if output.Work.ID != workID {
		t.Errorf("Work.ID = %d, want %d", output.Work.ID, workID)
	}
	if output.Work.Title != "非公開確認テスト" {
		t.Errorf("Work.Title = %q, want %q", output.Work.Title, "非公開確認テスト")
	}
}

// TestGetDBWorkArchiveNewUsecase_Execute_RejectsNonArchivableWork verifies the usecase
// reports a work that is not currently published (already archived, or deleted) as not
// found, matching the archivable scope Work.without_deleted.published.
//
// [Ja] TestGetDBWorkArchiveNewUsecase_Execute_RejectsNonArchivableWork は、現在公開中でない
// (すでにアーカイブ済み、または削除済みの) work を not found として報告することを検証する。
// アーカイブ可能な scope Work.without_deleted.published に一致する。
func TestGetDBWorkArchiveNewUsecase_Execute_RejectsNonArchivableWork(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name    string
		prepare func(b *testutil.WorkBuilder) *testutil.WorkBuilder
	}{
		{
			name: "アーカイブ済み (unpublished_at あり)",
			prepare: func(b *testutil.WorkBuilder) *testutil.WorkBuilder {
				return b.WithUnpublishedAt(now)
			},
		},
		{
			name: "削除済み (deleted_at あり)",
			prepare: func(b *testutil.WorkBuilder) *testutil.WorkBuilder {
				return b.WithDeletedAt(now)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			uc := NewGetDBWorkArchiveNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

			workID := tt.prepare(testutil.NewWorkBuilder(t, tx).WithTitle("非公開不可テスト")).Build()

			output, err := uc.Execute(context.Background(), GetDBWorkArchiveNewInput{WorkID: workID})
			if output != nil {
				t.Errorf("output = %+v, want nil for a non-archivable work", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
			}
		})
	}
}

// TestGetDBWorkArchiveNewUsecase_Execute_ReturnsNotFoundForMissingWork verifies a
// non-existent work id is reported as not found.
//
// [Ja] TestGetDBWorkArchiveNewUsecase_Execute_ReturnsNotFoundForMissingWork は、存在しない
// work id が not found として報告されることを検証する。
func TestGetDBWorkArchiveNewUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkArchiveNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))

	output, err := uc.Execute(context.Background(), GetDBWorkArchiveNewInput{WorkID: model.WorkID(1 << 62)})
	if output != nil {
		t.Errorf("output = %+v, want nil for a missing work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}
