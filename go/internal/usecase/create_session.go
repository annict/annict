package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/session"
)

// SessionResult セッション作成の結果
type SessionResult struct {
	PublicID string // Cookie値として使用するID
	UserID   int64  // ログインユーザーID
}

// CreateSessionUsecase セッション作成のビジネスロジック
type CreateSessionUsecase struct {
	queries *query.Queries
}

// NewCreateSessionUsecase 新しいCreateSessionUsecaseを作成
func NewCreateSessionUsecase(queries *query.Queries) *CreateSessionUsecase {
	return &CreateSessionUsecase{
		queries: queries,
	}
}

// Execute セッションを作成する
// tx: オプションでトランザクションを渡せる（nilの場合は通常のクエリ実行）
// flashMessage: オプションでflashメッセージを設定可能（空の場合は設定しない）
// encryptedPassword: ユーザーのencrypted_password（authenticatable_salt生成に必要）
func (uc *CreateSessionUsecase) Execute(ctx context.Context, tx *sql.Tx, userID int64, encryptedPassword string, flashMessage string) (*SessionResult, error) {
	// Public ID（Cookie値）を生成
	publicID, err := generatePublicID()
	if err != nil {
		return nil, fmt.Errorf("public ID生成エラー: %w", err)
	}

	// Private ID（DB保存用）を生成
	privateID := generatePrivateID(publicID)

	// authenticatable_saltを生成（encrypted_passwordの最初の29文字）
	// これはDeviseのセキュリティ機能で、パスワード変更時にセッションを無効化するために使用される
	authenticatableSalt := ""
	if len(encryptedPassword) >= 29 {
		authenticatableSalt = encryptedPassword[:29]
	}

	// Rails互換のCSRFトークンを生成
	csrfToken, err := session.GenerateCSRFToken()
	if err != nil {
		return nil, fmt.Errorf("CSRFトークン生成エラー: %w", err)
	}

	// セッションデータを作成（Railsのwarden形式）
	sessionData := map[string]any{
		"warden.user.user.key": []any{
			[]any{userID},
			authenticatableSalt, // encrypted_passwordの最初の29文字
		},
		"_csrf_token": csrfToken,
	}

	// flashメッセージを追加（ログイン成功時など）
	if flashMessage != "" {
		sessionData["flash"] = flashMessage
	}

	// JSONにエンコード
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("セッションデータのエンコードエラー: %w", err)
	}

	// トランザクションが渡された場合はそれを使用、なければ通常のクエリ実行
	var queries *query.Queries
	if tx != nil {
		queries = uc.queries.WithTx(tx)
	} else {
		queries = uc.queries
	}

	// DBにセッションを保存
	_, err = queries.CreateSession(ctx, query.CreateSessionParams{
		SessionID: privateID,
		Data:      jsonData,
	})
	if err != nil {
		return nil, fmt.Errorf("セッション保存エラー: %w", err)
	}

	return &SessionResult{
		PublicID: publicID,
		UserID:   userID,
	}, nil
}

// generatePublicID ランダムなpublic IDを生成
func generatePublicID() (string, error) {
	// 32バイトのランダムデータを生成
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

// generatePrivateID はpublic IDからprivate IDを生成
// Rails/Rackの実装と互換性のある形式: "2::" + SHA256(publicID)
func generatePrivateID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}
