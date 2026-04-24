// Package session はセッション管理機能を提供します
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/annict/annict/go/internal/model"
)

// Flash 一般的なフラッシュメッセージ（成功、情報、エラー、警告）
type Flash struct {
	Type    string `json:"type"`    // "success", "error", "info", "warning"
	Message string `json:"message"` // 表示するメッセージ
}

// FlashType フラッシュメッセージのタイプ定数
const (
	FlashSuccess = "success"
	FlashError   = "error"
	FlashInfo    = "info"
	FlashWarning = "warning"
)

// セッションキー
const (
	sessionKeyValidationError = "validation_error"
)

// SetFlash はフラッシュメッセージをCookieに設定する（FlashManagerに委譲）
func (m *Manager) SetFlash(w http.ResponseWriter, flashType, message string) {
	m.flashMgr.setFlash(w, flashType, message)
}

// GetFlash はフラッシュメッセージをCookieから取得して削除する（FlashManagerに委譲）
func (m *Manager) GetFlash(w http.ResponseWriter, r *http.Request) *Flash {
	return m.flashMgr.GetFlash(w, r)
}

// SetValidationError はバリデーションエラーをセッションに保存する
func (m *Manager) SetValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, ve model.ValidationError) error {
	data, err := json.Marshal(ve)
	if err != nil {
		return fmt.Errorf("failed to marshal validation error: %w", err)
	}

	return m.SetValue(ctx, w, r, sessionKeyValidationError, string(data))
}

// GetValidationError はバリデーションエラーをセッションから取得して削除する（1回限りの表示）
func (m *Manager) GetValidationError(ctx context.Context, r *http.Request) (*model.ValidationError, error) {
	value, err := m.getAndDeleteSessionValue(ctx, r, sessionKeyValidationError)
	if err != nil || value == "" {
		return nil, err
	}

	var ve model.ValidationError
	if err := json.Unmarshal([]byte(value), &ve); err != nil {
		return nil, fmt.Errorf("failed to unmarshal validation error: %w", err)
	}

	return &ve, nil
}
