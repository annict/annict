package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetDbWorkEditUsecase_Execute_ReturnsWork verifies the usecase returns the
// target work and the form options. It is a read-only usecase that opens no
// transaction, so the test uses SetupTx.
//
// [Ja] TestGetDbWorkEditUsecase_Execute_ReturnsWork は対象 work とフォーム選択肢を返すことを
// 検証する。本 UseCase は読み取りのみでトランザクションを開かないため SetupTx を使う。
func TestGetDbWorkEditUsecase_Execute_ReturnsWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	uc := NewGetDbWorkEditUsecase(workRepo, numberFormatRepo)

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("編集UseCaseテスト").
		WithSeason(2025, testutil.SeasonSummer).
		Build()
	if _, err := tx.Exec(`
		UPDATE works SET
			official_site_url = 'https://example.dev',
			twitter_username = 'handle',
			sc_tid = 42,
			started_on = '2025-07-01'
		WHERE id = $1
	`, int64(workID)); err != nil {
		t.Fatalf("works のフィールド設定に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), GetDbWorkEditInput{WorkID: workID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.Work == nil {
		t.Fatal("Work should not be nil")
	}
	if output.Work.ID != workID {
		t.Errorf("Work.ID = %d, want %d", output.Work.ID, workID)
	}
	if output.Work.Title != "編集UseCaseテスト" {
		t.Errorf("Work.Title = %q, want %q", output.Work.Title, "編集UseCaseテスト")
	}
	if output.Work.OfficialSiteURL != "https://example.dev" {
		t.Errorf("Work.OfficialSiteURL = %q, want %q", output.Work.OfficialSiteURL, "https://example.dev")
	}
	if output.Work.TwitterUsername == nil || *output.Work.TwitterUsername != "handle" {
		t.Errorf("Work.TwitterUsername = %v, want handle", output.Work.TwitterUsername)
	}
	if output.Work.ScTid == nil || *output.Work.ScTid != 42 {
		t.Errorf("Work.ScTid = %v, want 42", output.Work.ScTid)
	}
	if output.Work.SeasonYear == nil || *output.Work.SeasonYear != 2025 {
		t.Errorf("Work.SeasonYear = %v, want 2025", output.Work.SeasonYear)
	}
	if output.Work.StartedOn == nil {
		t.Error("Work.StartedOn should not be nil")
	}
	if output.NumberFormats == nil {
		t.Error("NumberFormats should not be nil (empty slice is acceptable)")
	}
}

// TestGetDbWorkEditUsecase_Execute_NotFound verifies a nonexistent work returns
// AppErrCodeResourceNotFound.
//
// [Ja] TestGetDbWorkEditUsecase_Execute_NotFound は存在しない work で
// AppErrCodeResourceNotFound を返すことを検証する。
func TestGetDbWorkEditUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	uc := NewGetDbWorkEditUsecase(workRepo, numberFormatRepo)

	_, err := uc.Execute(context.Background(), GetDbWorkEditInput{WorkID: model.WorkID(999999999)})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ae := model.AsAppError(err)
	if ae == nil {
		t.Fatalf("expected *model.AppError, got %T", err)
	}
	if ae.Code != model.AppErrCodeResourceNotFound {
		t.Errorf("AppError.Code = %v, want %v", ae.Code, model.AppErrCodeResourceNotFound)
	}
}
