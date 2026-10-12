package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestGetAccountDeletionNewUsecase_Execute(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name    string
		prepare func(t *testing.T, tx *sql.Tx, user *model.User)
		want    GetAccountDeletionNewOutput
	}{
		{
			name:    "サポーター契約もアプリケーションも無い",
			prepare: func(t *testing.T, tx *sql.Tx, user *model.User) {},
			want:    GetAccountDeletionNewOutput{},
		},
		{
			name: "有効なStripeのサブスクリプションがある",
			prepare: func(t *testing.T, tx *sql.Tx, user *model.User) {
				id := testutil.NewStripeSubscriberBuilder(t, tx).WithStripeStatus("active").Build()
				user.StripeSubscriberID = &id
			},
			want: GetAccountDeletionNewOutput{IsStripeSupporter: true},
		},
		{
			name: "解約済みのStripeのサブスクリプションは対象にしない",
			prepare: func(t *testing.T, tx *sql.Tx, user *model.User) {
				id := testutil.NewStripeSubscriberBuilder(t, tx).WithStripeStatus("canceled").Build()
				user.StripeSubscriberID = &id
			},
			want: GetAccountDeletionNewOutput{},
		},
		{
			name: "有効なGumroadのサブスクリプションがある",
			prepare: func(t *testing.T, tx *sql.Tx, user *model.User) {
				id := testutil.NewGumroadSubscriberBuilder(t, tx).Build()
				user.GumroadSubscriberID = &id
			},
			want: GetAccountDeletionNewOutput{IsGumroadSupporter: true},
		},
		{
			name: "終了したGumroadのサブスクリプションは対象にしない",
			prepare: func(t *testing.T, tx *sql.Tx, user *model.User) {
				id := testutil.NewGumroadSubscriberBuilder(t, tx).WithGumroadEndedAt(past).Build()
				user.GumroadSubscriberID = &id
			},
			want: GetAccountDeletionNewOutput{},
		},
		{
			name: "論理削除されていないOAuthアプリケーションがある",
			prepare: func(t *testing.T, tx *sql.Tx, user *model.User) {
				testutil.CreateTestOAuthApplication(t, tx, user.ID, false)
			},
			want: GetAccountDeletionNewOutput{HasKeptOAuthApplications: true},
		},
		{
			name: "論理削除済みのOAuthアプリケーションだけなら退会を妨げない",
			prepare: func(t *testing.T, tx *sql.Tx, user *model.User) {
				testutil.CreateTestOAuthApplication(t, tx, user.ID, true)
			},
			want: GetAccountDeletionNewOutput{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			queries := query.New(db).WithTx(tx)
			uc := NewGetAccountDeletionNewUsecase(
				repository.NewStripeSubscriberRepository(queries),
				repository.NewGumroadSubscriberRepository(queries),
				repository.NewOAuthApplicationRepository(queries),
			)

			user := &model.User{ID: testutil.NewUserBuilder(t, tx).Build()}
			tt.prepare(t, tx, user)

			got, err := uc.Execute(context.Background(), GetAccountDeletionNewInput{User: user})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if *got != tt.want {
				t.Errorf("出力 = %+v、期待値 = %+v", *got, tt.want)
			}
		})
	}
}

func TestGetAccountDeletionNewUsecase_Execute_NilUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	uc := NewGetAccountDeletionNewUsecase(
		repository.NewStripeSubscriberRepository(queries),
		repository.NewGumroadSubscriberRepository(queries),
		repository.NewOAuthApplicationRepository(queries),
	)

	if _, err := uc.Execute(context.Background(), GetAccountDeletionNewInput{}); err == nil {
		t.Error("エラーを期待したが、nilだった")
	}
}
