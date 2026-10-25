package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetDBWorkEditUsecase_Execute_ReturnsWorkは対象workとフォーム選択肢を返すことを
// 検証する。本UseCaseは読み取りのみでトランザクションを開かないためSetupTxを使う。
func TestGetDBWorkEditUsecase_Execute_ReturnsWork(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	uc := NewGetDBWorkEditUsecase(workRepo, numberFormatRepo)

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
		t.Fatalf("worksのフィールド設定に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), GetDBWorkEditInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: workID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if output.Work == nil {
		t.Fatal("Workがnilだった")
	}
	if output.Work.ID != workID {
		t.Errorf("Work.ID = %d、期待値 = %d", output.Work.ID, workID)
	}
	if output.Work.Title != "編集UseCaseテスト" {
		t.Errorf("Work.Title = %q、期待値 = %q", output.Work.Title, "編集UseCaseテスト")
	}
	if output.Work.OfficialSiteURL != "https://example.dev" {
		t.Errorf("Work.OfficialSiteURL = %q、期待値 = %q", output.Work.OfficialSiteURL, "https://example.dev")
	}
	if output.Work.TwitterUsername == nil || *output.Work.TwitterUsername != "handle" {
		t.Errorf("Work.TwitterUsername = %v、期待値 = handle", output.Work.TwitterUsername)
	}
	if output.Work.ScTid == nil || *output.Work.ScTid != 42 {
		t.Errorf("Work.ScTid = %v、期待値 = 42", output.Work.ScTid)
	}
	if output.Work.SeasonYear == nil || *output.Work.SeasonYear != 2025 {
		t.Errorf("Work.SeasonYear = %v、期待値 = 2025", output.Work.SeasonYear)
	}
	if output.Work.StartedOn == nil {
		t.Error("Work.StartedOnがnilだった")
	}
	if output.NumberFormats == nil {
		t.Error("NumberFormatsがnilだった (空のスライスは可)")
	}
}

// TestGetDBWorkEditUsecase_Execute_NotFoundは存在しないworkで
// AppErrCodeResourceNotFoundを返すことを検証する。
func TestGetDBWorkEditUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	uc := NewGetDBWorkEditUsecase(workRepo, numberFormatRepo)

	_, err := uc.Execute(context.Background(), GetDBWorkEditInput{User: &model.User{ID: 1, Role: model.RoleEditor}, WorkID: model.WorkID(999999999)})
	if err == nil {
		t.Fatal("エラーを期待したが、nilだった")
	}
	ae := model.AsAppError(err)
	if ae == nil {
		t.Fatalf("エラーの型 = %T、期待値 = *model.AppError", err)
	}
	if ae.Code != model.AppErrCodeResourceNotFound {
		t.Errorf("AppError.Code = %v、期待値 = %v", ae.Code, model.AppErrCodeResourceNotFound)
	}
}

// TestGetDBWorkEditUsecase_Execute_RejectsUnauthorizedUserBeforeLookupは、認可境界が
// 既存・未存在どちらのworkを取得するより前に、未認証と一般ユーザーを拒否することを検証する。
func TestGetDBWorkEditUsecase_Execute_RejectsUnauthorizedUserBeforeLookup(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	uc := NewGetDBWorkEditUsecase(repository.NewWorkRepository(queries), repository.NewNumberFormatRepository(queries))
	existingWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("編集認可テスト").Build()
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
			output, err := uc.Execute(context.Background(), GetDBWorkEditInput{
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
