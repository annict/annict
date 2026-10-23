package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

func TestGetDBPersonNewUsecase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		user          *model.User
		wantForbidden bool
	}{
		{name: "未ログインは拒否", user: nil, wantForbidden: true},
		{name: "一般ユーザーは拒否", user: &model.User{ID: 1, Role: model.RoleUser}, wantForbidden: true},
		{name: "編集者は許可", user: &model.User{ID: 1, Role: model.RoleEditor}},
		{name: "管理者は許可", user: &model.User{ID: 1, Role: model.RoleAdmin}},
	}

	uc := NewGetDBPersonNewUsecase()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := uc.Execute(context.Background(), GetDBPersonNewInput{User: tt.user})
			if !tt.wantForbidden {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				return
			}

			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("エラー = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}
}
