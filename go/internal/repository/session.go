// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/annict/annict/internal/query"
)

// SessionRepository はSession関連のデータアクセスを担当します
type SessionRepository struct {
	queries *query.Queries
}

// NewSessionRepository はSessionRepositoryを作成します
func NewSessionRepository(queries *query.Queries) *SessionRepository {
	return &SessionRepository{queries: queries}
}

// TouchSession はセッションのupdated_atを更新します
func (r *SessionRepository) TouchSession(ctx context.Context, sessionID string) error {
	// Rails/Rackの実装と互換性のある形式でprivate IDを生成
	privateID := r.generatePrivateID(sessionID)

	// セッションのupdated_atのみを更新
	return r.queries.TouchSession(ctx, privateID)
}

// GetSessionByID はセッションIDからセッションを取得します
func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*query.Session, error) {
	privateID := r.generatePrivateID(sessionID)
	session, err := r.queries.GetSessionByID(ctx, privateID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserByID はユーザーIDからユーザー情報を取得します
func (r *SessionRepository) GetUserByID(ctx context.Context, userID int64) (*query.GetUserByIDRow, error) {
	user, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateSession はセッションを更新します
func (r *SessionRepository) UpdateSession(ctx context.Context, sessionID string, data []byte) error {
	privateID := r.generatePrivateID(sessionID)
	return r.queries.UpdateSession(ctx, query.UpdateSessionParams{
		SessionID: privateID,
		Data:      data,
	})
}

// CreateSession はセッションを作成します
func (r *SessionRepository) CreateSession(ctx context.Context, sessionID string, data []byte) (query.Session, error) {
	privateID := r.generatePrivateID(sessionID)
	return r.queries.CreateSession(ctx, query.CreateSessionParams{
		SessionID: privateID,
		Data:      data,
	})
}

// generatePrivateID はpublic IDからprivate IDを生成
// Rails/Rackの実装と互換性のある形式: "2::" + SHA256(publicID)
func (r *SessionRepository) generatePrivateID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}
