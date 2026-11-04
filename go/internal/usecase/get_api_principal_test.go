package usecase

import (
	"context"
	"database/sql"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestGetAPIPrincipalUsecase_Execute(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		// prepareはトークンを作り、照合に使うトークンの文字列を返す
		prepare func(t *testing.T, tx *sql.Tx, userID model.UserID) string
		want    *model.APIPrincipal
	}{
		{
			name: "フラグの無い利用者の個人用アクセストークン",
			prepare: func(t *testing.T, tx *sql.Tx, userID model.UserID) string {
				token := uuid.New().String()
				testutil.NewOAuthAccessTokenBuilder(t, tx, userID).WithToken(token).WithScopes("read write").Build()
				return token
			},
			want: &model.APIPrincipal{TokenKind: model.APITokenKindPersonalAccessToken, Scopes: []string{"read", "write"}},
		},
		{
			name: "フラグの有効な利用者のOAuthのトークン",
			prepare: func(t *testing.T, tx *sql.Tx, userID model.UserID) string {
				token := uuid.New().String()
				applicationID := testutil.CreateTestOAuthApplication(t, tx, userID, false)
				testutil.NewOAuthAccessTokenBuilder(t, tx, userID).WithToken(token).WithApplicationID(applicationID).Build()
				testutil.NewFeatureFlagBuilder(t, tx).WithUserID(userID).WithName(string(model.FeatureFlagRESTAPIV2)).Build()
				return token
			},
			want: &model.APIPrincipal{TokenKind: model.APITokenKindOAuthAccessToken, Scopes: []string{"read"}, RESTAPIV2Enabled: true},
		},
		{
			name: "期限の切れたトークンは受け付けない",
			prepare: func(t *testing.T, tx *sql.Tx, userID model.UserID) string {
				token := uuid.New().String()
				testutil.NewOAuthAccessTokenBuilder(t, tx, userID).WithToken(token).WithCreatedAt(now.Add(-2 * time.Hour)).WithExpiresIn(3600).Build()
				return token
			},
		},
		{
			name: "失効したトークンは受け付けない",
			prepare: func(t *testing.T, tx *sql.Tx, userID model.UserID) string {
				token := uuid.New().String()
				testutil.NewOAuthAccessTokenBuilder(t, tx, userID).WithToken(token).WithRevokedAt(now.Add(-time.Minute)).Build()
				return token
			},
		},
		{
			name: "無いトークンは受け付けない",
			prepare: func(t *testing.T, tx *sql.Tx, userID model.UserID) string {
				return uuid.New().String()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			queries := query.New(db)
			uc := NewGetAPIPrincipalUsecase(
				repository.NewOAuthAccessTokenRepository(queries).WithTx(tx),
				repository.NewFeatureFlagRepository(queries).WithTx(tx),
			)
			uc.now = func() time.Time { return now }

			userID := testutil.NewUserBuilder(t, tx).Build()
			token := tt.prepare(t, tx, userID)

			got, err := uc.Execute(context.Background(), token)
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if tt.want == nil {
				if got != nil {
					t.Errorf("Execute() = %+v、期待値 = nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("Execute() = nil、呼び出し主体を期待")
			}
			if got.UserID != userID {
				t.Errorf("UserID = %v、期待値 = %v", got.UserID, userID)
			}
			if got.TokenKind != tt.want.TokenKind {
				t.Errorf("TokenKind = %v、期待値 = %v", got.TokenKind, tt.want.TokenKind)
			}
			if !slices.Equal(got.Scopes, tt.want.Scopes) {
				t.Errorf("Scopes = %v、期待値 = %v", got.Scopes, tt.want.Scopes)
			}
			if got.RESTAPIV2Enabled != tt.want.RESTAPIV2Enabled {
				t.Errorf("RESTAPIV2Enabled = %v、期待値 = %v", got.RESTAPIV2Enabled, tt.want.RESTAPIV2Enabled)
			}
		})
	}
}
