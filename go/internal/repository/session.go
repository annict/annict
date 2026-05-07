// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// SessionRepository はSession関連のデータアクセスを担当します
type SessionRepository struct {
	queries *query.Queries
}

// NewSessionRepository はSessionRepositoryを作成します
func NewSessionRepository(queries *query.Queries) *SessionRepository {
	return &SessionRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *SessionRepository) WithTx(tx *sql.Tx) *SessionRepository {
	return &SessionRepository{queries: r.queries.WithTx(tx)}
}

// TouchSession はセッションのupdated_atを更新します
func (r *SessionRepository) TouchSession(ctx context.Context, sessionID string) error {
	// Rails/Rackの実装と互換性のある形式でprivate IDを生成
	privateID := r.generatePrivateID(sessionID)

	// セッションのupdated_atのみを更新
	return r.queries.TouchSession(ctx, privateID)
}

// GetSessionByID はセッションIDからセッションを取得します
func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*model.Session, error) {
	privateID := r.generatePrivateID(sessionID)
	row, err := r.queries.GetSessionByID(ctx, privateID)
	if err != nil {
		return nil, err
	}
	return toSessionModel(row), nil
}

// GetUserByID はユーザーIDからユーザー情報を取得します
func (r *SessionRepository) GetUserByID(ctx context.Context, userID model.UserID) (*model.User, error) {
	row, err := r.queries.GetUserByID(ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	var stripeSubID *model.StripeSubscriberID
	if row.StripeSubscriberID.Valid {
		id := model.StripeSubscriberID(row.StripeSubscriberID.Int64)
		stripeSubID = &id
	}
	var gumroadSubID *model.GumroadSubscriberID
	if row.GumroadSubscriberID.Valid {
		id := model.GumroadSubscriberID(row.GumroadSubscriberID.Int64)
		gumroadSubID = &id
	}
	return &model.User{
		ID:                  model.UserID(row.ID),
		Username:            row.Username,
		Email:               row.Email,
		Role:                row.Role,
		EncryptedPassword:   row.EncryptedPassword,
		Locale:              row.Locale,
		TimeZone:            row.TimeZone,
		StripeSubscriberID:  stripeSubID,
		GumroadSubscriberID: gumroadSubID,
		NotificationsCount:  row.NotificationsCount,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
		ProfileImageData:    row.ProfileImageData,
	}, nil
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
func (r *SessionRepository) CreateSession(ctx context.Context, sessionID string, data []byte) (*model.Session, error) {
	privateID := r.generatePrivateID(sessionID)
	row, err := r.queries.CreateSession(ctx, query.CreateSessionParams{
		SessionID: privateID,
		Data:      data,
	})
	if err != nil {
		return nil, err
	}
	return toSessionModel(row), nil
}

// DeleteSession はセッションを削除します
func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	privateID := r.generatePrivateID(sessionID)
	return r.queries.DeleteSession(ctx, privateID)
}

// generatePrivateID はpublic IDからprivate IDを生成
// Rails/Rackの実装と互換性のある形式: "2::" + SHA256(publicID)
func (r *SessionRepository) generatePrivateID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}

// toSessionModel はsqlcのSessionをmodel.Sessionに変換する
func toSessionModel(row query.Session) *model.Session {
	return &model.Session{
		ID:        row.ID,
		SessionID: row.SessionID,
		Data:      row.Data,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
