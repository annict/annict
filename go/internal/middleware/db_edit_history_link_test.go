package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates"
)

// recordingFeatureFlagCheckerは判定の呼び出しを数えて受け取った値を記録し、決めた結果を返す。
type recordingFeatureFlagChecker struct {
	enabled     bool
	err         error
	calls       int
	deviceToken string
	userID      model.UserID
	flag        model.FeatureFlagName
}

func (c *recordingFeatureFlagChecker) IsEnabledByDeviceOrUser(_ context.Context, deviceToken string, userID model.UserID, flag model.FeatureFlagName) (bool, error) {
	c.calls++
	c.deviceToken = deviceToken
	c.userID = userID
	c.flag = flag
	return c.enabled, c.err
}

func TestDBEditHistoryLinkMiddleware(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name      string
		path      string
		device    string
		user      *model.User
		enabled   bool
		err       error
		want      bool
		wantCalls int
	}{
		{name: "フラグが有効なデバイスのAnnict DBの画面", path: "/db/works", device: "device-1", enabled: true, want: true, wantCalls: 1},
		{name: "Annict DBのトップ", path: "/db", device: "device-1", enabled: true, want: true, wantCalls: 1},
		{name: "フラグが無効なデバイス", path: "/db/works", device: "device-1", want: false, wantCalls: 1},
		{name: "デバイスが無くてもフラグが有効なユーザーなら出す", path: "/db/works", user: &model.User{ID: 7}, enabled: true, want: true, wantCalls: 1},
		{name: "デバイスとユーザーの両方を判定に渡す", path: "/db/works", device: "device-1", user: &model.User{ID: 7}, enabled: true, want: true, wantCalls: 1},
		{name: "判定に失敗したときは出さない", path: "/db/works", device: "device-1", enabled: true, err: errors.New("失敗"), want: false, wantCalls: 1},
		{name: "Annict DB以外の画面では判定しない", path: "/works/1", device: "device-1", enabled: true, want: false, wantCalls: 0},
		{name: "似た接頭辞の画面では判定しない", path: "/dbx", device: "device-1", enabled: true, want: false, wantCalls: 0},
		{name: "デバイスもユーザーも無いときは判定しない", path: "/db/works", enabled: true, want: false, wantCalls: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checker := &recordingFeatureFlagChecker{enabled: tt.enabled, err: tt.err}
			var got bool
			handler := NewDBEditHistoryLinkMiddleware(checker).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = templates.IsDBEditHistoryLinkVisible(r.Context())
			}))
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.device != "" {
				req.AddCookie(&http.Cookie{Name: DeviceTokenCookieName, Value: tt.device})
			}
			if tt.user != nil {
				req = req.WithContext(context.WithValue(req.Context(), UserContextKey, tt.user))
			}
			handler.ServeHTTP(httptest.NewRecorder(), req)

			if got != tt.want {
				t.Errorf("編集履歴の項目を出すか = %v、期待値 = %v", got, tt.want)
			}
			if checker.calls != tt.wantCalls {
				t.Errorf("判定の回数 = %d、期待値 = %d", checker.calls, tt.wantCalls)
			}
			if tt.wantCalls == 0 {
				return
			}
			if checker.deviceToken != tt.device {
				t.Errorf("判定に渡したデバイスのトークン = %q、期待値 = %q", checker.deviceToken, tt.device)
			}
			var wantUserID model.UserID
			if tt.user != nil {
				wantUserID = tt.user.ID
			}
			if checker.userID != wantUserID {
				t.Errorf("判定に渡したユーザーのID = %d、期待値 = %d", checker.userID, wantUserID)
			}
			if checker.flag != model.FeatureFlagDBEditHistory {
				t.Errorf("判定したフラグ = %q、期待値 = %q", checker.flag, model.FeatureFlagDBEditHistory)
			}
		})
	}
}
