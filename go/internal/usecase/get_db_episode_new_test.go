package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetDBEpisodeNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookupは、認可境界が
// 既存・未存在どちらの作品を取得するより前に、未認証と一般ユーザーを拒否することを検証する。
// 本UseCaseは読み取りのみでトランザクションを開かないためSetupTxを使う。
func TestGetDBEpisodeNewUsecase_Execute_RejectsUnauthorizedUserBeforeLookup(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBEpisodeNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))
	existingWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("一括作成認可テスト").Build()
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
			output, err := uc.Execute(context.Background(), GetDBEpisodeNewInput{
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

// TestGetDBEpisodeNewUsecase_Execute_ReportsRawNumberRequiredは、有効な放送枠がある作品で
// だけ、フォームが話数を求めるよう伝えることを検証する。
func TestGetDBEpisodeNewUsecase_Execute_ReportsRawNumberRequired(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBEpisodeNewUsecase(repository.NewWorkRepository(query.New(db).WithTx(tx)))
	editor := &model.User{ID: 1, Role: model.RoleEditor}

	slotWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("放送予定のある作品").Build()
	channelID := testutil.NewChannelBuilder(t, tx).Build()
	testutil.NewSlotBuilder(t, tx).WithWorkID(slotWorkID).WithChannelID(channelID).Build()
	plainWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("放送予定の無い作品").Build()

	for name, tt := range map[string]struct {
		workID model.WorkID
		want   bool
	}{
		"放送予定のある作品": {workID: slotWorkID, want: true},
		"放送予定の無い作品": {workID: plainWorkID, want: false},
	} {
		output, err := uc.Execute(context.Background(), GetDBEpisodeNewInput{User: editor, WorkID: tt.workID})
		if err != nil {
			t.Fatalf("%s: Execute()のエラー = %v", name, err)
		}
		if output.RawNumberRequired != tt.want {
			t.Errorf("%s: RawNumberRequired = %v、期待値 = %v", name, output.RawNumberRequired, tt.want)
		}
		if !output.ManualCreationState.Allowed() {
			t.Errorf("%s: 手動作成が許可されていない", name)
		}
	}
}
