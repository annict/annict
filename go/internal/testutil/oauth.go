package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestOAuthApplicationはテスト用のOAuthアプリケーションをデータベースに作成し、IDを返す。
// deletedがtrueなら論理削除済みにする。uidは一意制約の対象のため、UUIDから作る。
func CreateTestOAuthApplication(t *testing.T, tx *sql.Tx, ownerID model.UserID, deleted bool) int64 {
	t.Helper()

	now := time.Now()
	var deletedAt sql.NullTime
	if deleted {
		deletedAt = sql.NullTime{Time: now, Valid: true}
	}
	return createTestRow(t, tx, "OAuthアプリケーション",
		`INSERT INTO oauth_applications (name, uid, secret, redirect_uri, owner_id, owner_type, deleted_at, created_at, updated_at)
		VALUES ('アプリケーション', $1, 'secret', 'https://example.com/callback', $2, 'User', $3, $4, $4) RETURNING id`,
		uuid.New().String(), int64(ownerID), deletedAt, now,
	)
}

// CreateTestOAuthAccessTokenはテスト用のOAuthのアクセストークンをデータベースに作成し、IDを返す。
func CreateTestOAuthAccessToken(t *testing.T, tx *sql.Tx, resourceOwnerID model.UserID, applicationID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "OAuthのアクセストークン",
		`INSERT INTO oauth_access_tokens (resource_owner_id, application_id, token, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		int64(resourceOwnerID), applicationID, uuid.New().String(), time.Now(),
	)
}

// CreateTestOAuthAccessGrantはテスト用のOAuthの認可をデータベースに作成し、IDを返す。
func CreateTestOAuthAccessGrant(t *testing.T, tx *sql.Tx, resourceOwnerID model.UserID, applicationID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "OAuthの認可",
		`INSERT INTO oauth_access_grants (resource_owner_id, application_id, token, expires_in, redirect_uri, created_at)
		VALUES ($1, $2, $3, 600, 'https://example.com/callback', $4) RETURNING id`,
		int64(resourceOwnerID), applicationID, uuid.New().String(), time.Now(),
	)
}

// OAuthAccessTokenBuilderは、スコープ・期限・失効を指定できるアクセストークンのテストデータビルダー
type OAuthAccessTokenBuilder struct {
	t               *testing.T
	tx              *sql.Tx
	resourceOwnerID model.UserID
	applicationID   *int64
	token           string
	scopes          string
	expiresIn       *int32
	revokedAt       *time.Time
	createdAt       time.Time
}

// NewOAuthAccessTokenBuilderは、resourceOwnerIDの利用者の個人用アクセストークン (スコープは `read`) を作るビルダーを返す。
// tokenは一意制約の対象のため、UUIDから作る
func NewOAuthAccessTokenBuilder(t *testing.T, tx *sql.Tx, resourceOwnerID model.UserID) *OAuthAccessTokenBuilder {
	return &OAuthAccessTokenBuilder{
		t:               t,
		tx:              tx,
		resourceOwnerID: resourceOwnerID,
		token:           uuid.New().String(),
		scopes:          "read",
		createdAt:       time.Now().UTC(),
	}
}

// WithApplicationIDは、applicationIDのアプリケーションにOAuthで発行したトークンにする
func (b *OAuthAccessTokenBuilder) WithApplicationID(applicationID int64) *OAuthAccessTokenBuilder {
	b.applicationID = &applicationID
	return b
}

// WithTokenはトークンの文字列を設定する
func (b *OAuthAccessTokenBuilder) WithToken(token string) *OAuthAccessTokenBuilder {
	b.token = token
	return b
}

// WithScopesはスコープ (空白区切り) を設定する
func (b *OAuthAccessTokenBuilder) WithScopes(scopes string) *OAuthAccessTokenBuilder {
	b.scopes = scopes
	return b
}

// WithExpiresInは発行から失効までの秒数を設定する
func (b *OAuthAccessTokenBuilder) WithExpiresIn(seconds int32) *OAuthAccessTokenBuilder {
	b.expiresIn = &seconds
	return b
}

// WithRevokedAtは失効した時刻を設定する
func (b *OAuthAccessTokenBuilder) WithRevokedAt(revokedAt time.Time) *OAuthAccessTokenBuilder {
	b.revokedAt = &revokedAt
	return b
}

// WithCreatedAtは発行した時刻を設定する
func (b *OAuthAccessTokenBuilder) WithCreatedAt(createdAt time.Time) *OAuthAccessTokenBuilder {
	b.createdAt = createdAt
	return b
}

// Buildはアクセストークンをデータベースに作成し、IDを返す
func (b *OAuthAccessTokenBuilder) Build() model.OAuthAccessTokenID {
	b.t.Helper()
	return model.OAuthAccessTokenID(createTestRow(b.t, b.tx, "アクセストークン",
		`INSERT INTO oauth_access_tokens (resource_owner_id, application_id, token, scopes, expires_in, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		int64(b.resourceOwnerID), b.applicationID, b.token, b.scopes, b.expiresIn, b.revokedAt, b.createdAt,
	))
}
