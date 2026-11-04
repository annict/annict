package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// OAuthAccessTokenRepositoryはOAuthのアクセストークン関連のデータアクセスを担当します
type OAuthAccessTokenRepository struct {
	queries *query.Queries
}

// NewOAuthAccessTokenRepositoryはOAuthAccessTokenRepositoryを作成します
func NewOAuthAccessTokenRepository(queries *query.Queries) *OAuthAccessTokenRepository {
	return &OAuthAccessTokenRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *OAuthAccessTokenRepository) WithTx(tx *sql.Tx) *OAuthAccessTokenRepository {
	return &OAuthAccessTokenRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーに発行したアクセストークンと、ユーザーが持つ論理削除済みのアプリケーションに残るアクセストークンを最大limit件削除し、削除した件数を返します。
func (r *OAuthAccessTokenRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteOAuthAccessTokensByUserID(ctx, query.DeleteOAuthAccessTokensByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// FindByTokenはトークンの文字列でアクセストークンを引きます。
// トークンが無いか、持ち主の利用者が退会・利用停止している場合は (nil, nil) を返します
func (r *OAuthAccessTokenRepository) FindByToken(ctx context.Context, token string) (*model.OAuthAccessToken, error) {
	row, err := r.queries.GetOAuthAccessTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	accessToken := &model.OAuthAccessToken{
		ID:              model.OAuthAccessTokenID(row.ID),
		ResourceOwnerID: model.UserID(row.ResourceOwnerID),
		Personal:        !row.ApplicationID.Valid,
		CreatedAt:       row.CreatedAt,
		// Rails版 (Doorkeeper) と同じく、スコープは空白区切りで保存している
		Scopes: strings.Fields(row.Scopes.String),
	}
	if row.ExpiresIn.Valid {
		expiresIn := time.Duration(row.ExpiresIn.Int32) * time.Second
		accessToken.ExpiresIn = &expiresIn
	}
	if row.RevokedAt.Valid {
		revokedAt := row.RevokedAt.Time
		accessToken.RevokedAt = &revokedAt
	}
	return accessToken, nil
}
