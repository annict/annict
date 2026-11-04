package model_test

import (
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

func TestOAuthAccessToken_IsUsableAt(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	expiresIn := time.Hour
	revokedAt := createdAt.Add(time.Minute)

	tests := []struct {
		name  string
		token model.OAuthAccessToken
		now   time.Time
		want  bool
	}{
		{name: "期限も失効も無い", token: model.OAuthAccessToken{CreatedAt: createdAt}, now: createdAt.Add(24 * 365 * time.Hour), want: true},
		{name: "期限の前", token: model.OAuthAccessToken{CreatedAt: createdAt, ExpiresIn: &expiresIn}, now: createdAt.Add(59 * time.Minute), want: true},
		{name: "期限ちょうど", token: model.OAuthAccessToken{CreatedAt: createdAt, ExpiresIn: &expiresIn}, now: createdAt.Add(time.Hour), want: false},
		{name: "失効済み", token: model.OAuthAccessToken{CreatedAt: createdAt, RevokedAt: &revokedAt}, now: createdAt.Add(2 * time.Minute), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.token.IsUsableAt(tt.now); got != tt.want {
				t.Errorf("IsUsableAt() = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}
