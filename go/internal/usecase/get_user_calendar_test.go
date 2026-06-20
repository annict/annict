package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestGetUserCalendarUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("正常系: ユーザーのカレンダーデータを取得できる", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		testutil.NewUserBuilder(t, tx).
			WithUsername("calendar_test_user").
			WithEmail("calendar_test@example.com").
			Build()

		userCalendarRepo := repository.NewUserCalendarRepository(queries)
		uc := NewGetUserCalendarUsecase(userCalendarRepo)

		result, err := uc.Execute(context.Background(), GetUserCalendarInput{
			Username: "calendar_test_user",
			Now:      time.Now(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if result.UserCalendar == nil {
			t.Fatal("UserCalendar should not be nil")
		}

		if result.UserCalendar.Username != "calendar_test_user" {
			t.Errorf("Username = %q, want %q", result.UserCalendar.Username, "calendar_test_user")
		}
	})

	t.Run("異常系: 存在しないユーザーの場合はエラーを返す", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)

		userCalendarRepo := repository.NewUserCalendarRepository(queries)
		uc := NewGetUserCalendarUsecase(userCalendarRepo)

		_, err := uc.Execute(context.Background(), GetUserCalendarInput{
			Username: "nonexistent_user",
			Now:      time.Now(),
		})
		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}
