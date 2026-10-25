package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetDBWorkFormOptionsUsecase_Execute_RequiresCommitterは、フォームの選択肢を
// committerにだけ返し、未認証と一般ユーザーを拒否することを検証する。本UseCaseは読み取り
// のみでトランザクションを開かないためSetupTxを使う。
func TestGetDBWorkFormOptionsUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := NewGetDBWorkFormOptionsUsecase(repository.NewNumberFormatRepository(query.New(db).WithTx(tx)))

	tests := []struct {
		name          string
		user          *model.User
		wantForbidden bool
	}{
		{name: "未認証", user: nil, wantForbidden: true},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}, wantForbidden: true},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBWorkFormOptionsInput{User: tt.user})
			if !tt.wantForbidden {
				if err != nil {
					t.Fatalf("Execute()のエラー = %v", err)
				}
				if output == nil {
					t.Fatal("outputがnilだった")
				}
				return
			}
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
