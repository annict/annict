// Package session はセッション管理機能を提供します
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Flash 一般的なフラッシュメッセージ（成功、情報、エラー、警告）
type Flash struct {
	Type    string `json:"type"`    // "success", "error", "info", "warning"
	Message string `json:"message"` // 表示するメッセージ
}

// FormErrors フォームバリデーションエラー
type FormErrors struct {
	Global []string            `json:"global,omitempty"` // フィールド横断のグローバルエラー
	Fields map[string][]string `json:"fields,omitempty"` // フィールドごとのエラー
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
	sessionKeyFormErrors = "form_errors"
)

// SetFlash はフラッシュメッセージをCookieに設定する（FlashManagerに委譲）
func (m *Manager) SetFlash(w http.ResponseWriter, flashType, message string) {
	m.flashMgr.setFlash(w, flashType, message)
}

// GetFlash はフラッシュメッセージをCookieから取得して削除する（FlashManagerに委譲）
func (m *Manager) GetFlash(w http.ResponseWriter, r *http.Request) *Flash {
	return m.flashMgr.GetFlash(w, r)
}

// SetFormErrors フォームバリデーションエラーを設定
func (m *Manager) SetFormErrors(ctx context.Context, w http.ResponseWriter, r *http.Request, errors FormErrors) error {
	data, err := json.Marshal(errors)
	if err != nil {
		return fmt.Errorf("failed to marshal form errors: %w", err)
	}

	return m.SetValue(ctx, w, r, sessionKeyFormErrors, string(data))
}

// GetFormErrors フォームエラーを取得して削除（1回限りの表示）
func (m *Manager) GetFormErrors(ctx context.Context, r *http.Request) (*FormErrors, error) {
	value, err := m.getAndDeleteSessionValue(ctx, r, sessionKeyFormErrors)
	if err != nil || value == "" {
		return nil, err
	}

	var errors FormErrors
	if err := json.Unmarshal([]byte(value), &errors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal form errors: %w", err)
	}

	return &errors, nil
}

// AddGlobalError グローバルエラーを追加するヘルパー
func (fe *FormErrors) AddGlobalError(message string) {
	if fe.Global == nil {
		fe.Global = []string{}
	}
	fe.Global = append(fe.Global, message)
}

// AddFieldError フィールドエラーを追加するヘルパー
func (fe *FormErrors) AddFieldError(field, message string) {
	if fe.Fields == nil {
		fe.Fields = make(map[string][]string)
	}
	fe.Fields[field] = append(fe.Fields[field], message)
}

// HasErrors エラーが存在するか確認
func (fe *FormErrors) HasErrors() bool {
	return len(fe.Global) > 0 || len(fe.Fields) > 0
}

// HasFieldError 特定のフィールドにエラーがあるか確認
func (fe *FormErrors) HasFieldError(field string) bool {
	_, exists := fe.Fields[field]
	return exists
}

// FieldError はフィールドエラーを表す構造体（テンプレート用）
type FieldError struct {
	Field   string
	Message string
}

// FieldErrors フィールドエラーを列挙可能な形式で取得
func (fe *FormErrors) FieldErrors() []FieldError {
	var errors []FieldError
	for field, messages := range fe.Fields {
		for _, message := range messages {
			errors = append(errors, FieldError{
				Field:   field,
				Message: message,
			})
		}
	}
	return errors
}
