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

// ToJSON FlashをJSON文字列に変換
func (f *Flash) ToJSON() (string, error) {
	data, err := json.Marshal(f)
	if err != nil {
		return "", fmt.Errorf("failed to marshal flash: %w", err)
	}
	return string(data), nil
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
	sessionKeyFlash      = "flash"
	sessionKeyFormErrors = "form_errors"
)

// SetFlash 一般的なフラッシュメッセージを設定
func (m *Manager) SetFlash(ctx context.Context, w http.ResponseWriter, r *http.Request, flashType, message string) error {
	flash := Flash{
		Type:    flashType,
		Message: message,
	}

	data, err := json.Marshal(flash)
	if err != nil {
		return fmt.Errorf("failed to marshal flash: %w", err)
	}

	return m.SetValue(ctx, w, r, sessionKeyFlash, string(data))
}

// GetFlash フラッシュメッセージを取得して削除（1回限りの表示）
func (m *Manager) GetFlash(ctx context.Context, r *http.Request) (*Flash, error) {
	value, err := m.getAndDeleteSessionValue(ctx, r, sessionKeyFlash)
	if err != nil || value == "" {
		return nil, err
	}

	var flash Flash
	if err := json.Unmarshal([]byte(value), &flash); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flash: %w", err)
	}

	return &flash, nil
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

// FlashManager はフラッシュメッセージを管理するヘルパー
type FlashManager struct {
	manager *Manager
}

// NewFlashManager は新しいFlashManagerを作成します
func NewFlashManager(manager *Manager) *FlashManager {
	return &FlashManager{manager: manager}
}

// SetFlash フラッシュメッセージを設定します
func (fm *FlashManager) SetFlash(w http.ResponseWriter, r *http.Request, flashType, message string) error {
	return fm.manager.SetFlash(r.Context(), w, r, flashType, message)
}

// GetFlash フラッシュメッセージを取得します
func (fm *FlashManager) GetFlash(r *http.Request) (*Flash, error) {
	return fm.manager.GetFlash(r.Context(), r)
}

// SetFormErrors フォームエラーを設定します
func (fm *FlashManager) SetFormErrors(w http.ResponseWriter, r *http.Request, errors *FormErrors) error {
	return fm.manager.SetFormErrors(r.Context(), w, r, *errors)
}

// GetFormErrors フォームエラーを取得します
func (fm *FlashManager) GetFormErrors(w http.ResponseWriter, r *http.Request) *FormErrors {
	errors, err := fm.manager.GetFormErrors(r.Context(), r)
	if err != nil || errors == nil {
		return nil
	}
	return errors
}
