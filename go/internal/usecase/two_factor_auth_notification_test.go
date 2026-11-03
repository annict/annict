package usecase

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/model"
)

// recordingJobInserterは投入されたジョブの引数を記録するJobInserter。
// 同時に実行するテストからも呼ばれるため、記録はロックで守る。
type recordingJobInserter struct {
	mu   sync.Mutex
	args []river.JobArgs
}

func (r *recordingJobInserter) Insert(_ context.Context, args river.JobArgs, _ *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.args = append(r.args, args)
	return &rivertype.JobInsertResult{}, nil
}

// twoFactorAuthNotificationEventsは、userIDのユーザーのメールアドレスへ投入した二要素認証の通知の種類を、投入した順に返す。
func (r *recordingJobInserter) twoFactorAuthNotificationEvents(t *testing.T, db *sql.DB, userID model.UserID) []model.TwoFactorAuthNotificationEvent {
	t.Helper()

	var email string
	if err := db.QueryRow("SELECT email FROM users WHERE id = $1", int64(userID)).Scan(&email); err != nil {
		t.Fatalf("ユーザーのメールアドレスの取得に失敗: %v", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	var events []model.TwoFactorAuthNotificationEvent
	for _, args := range r.args {
		if notification, ok := args.(dispatcher.SendTwoFactorAuthNotificationEmailArgs); ok && notification.Email == email {
			events = append(events, model.TwoFactorAuthNotificationEvent(notification.Event))
		}
	}
	return events
}

// assertTwoFactorAuthNotificationsは、ユーザーへ投入した通知の種類が期待どおりかを確かめる。
func assertTwoFactorAuthNotifications(t *testing.T, recorder *recordingJobInserter, db *sql.DB, userID model.UserID, want ...model.TwoFactorAuthNotificationEvent) {
	t.Helper()

	got := recorder.twoFactorAuthNotificationEvents(t, db, userID)
	if len(got) != len(want) {
		t.Errorf("通知の種類 = %v、期待値 = %v", got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("通知の種類 = %v、期待値 = %v", got, want)
			return
		}
	}
}
