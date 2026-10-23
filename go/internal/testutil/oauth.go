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
