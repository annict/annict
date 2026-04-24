package model

import (
	"errors"
	"fmt"
)

// ValidationError はバリデーションエラーを表す。
// Handler はこのエラーを受け取ったらフォームを再描画する（422）。
type ValidationError struct {
	// Global はフォーム全体のエラーメッセージ（特定のフィールドに紐づかないエラー）
	Global []string `json:"global,omitempty"`
	// Fields はフィールドごとのエラーメッセージ
	Fields map[string][]string `json:"fields,omitempty"`
}

func (e *ValidationError) Error() string { return "validation failed" }

// AddGlobal はグローバルエラーを追加する
func (e *ValidationError) AddGlobal(message string) {
	e.Global = append(e.Global, message)
}

// AddField はフィールドエラーを追加する
func (e *ValidationError) AddField(field, message string) {
	if e.Fields == nil {
		e.Fields = make(map[string][]string)
	}
	e.Fields[field] = append(e.Fields[field], message)
}

// HasErrors はエラーがあるかどうかを返す
func (e *ValidationError) HasErrors() bool {
	if e == nil {
		return false
	}
	return len(e.Global) > 0 || len(e.Fields) > 0
}

// HasFieldError は指定されたフィールドにエラーがあるかどうかを返す
func (e *ValidationError) HasFieldError(field string) bool {
	if e == nil || e.Fields == nil {
		return false
	}
	return len(e.Fields[field]) > 0
}

// GetFieldErrors は指定されたフィールドのエラーメッセージを返す
func (e *ValidationError) GetFieldErrors(field string) []string {
	if e == nil || e.Fields == nil {
		return nil
	}
	return e.Fields[field]
}

// FieldError はフィールドエラーを表す構造体（テンプレート用）
type FieldError struct {
	Field   string
	Message string
}

// FieldErrors はフィールドエラーを列挙可能な形式で取得する
func (e *ValidationError) FieldErrors() []FieldError {
	if e == nil || e.Fields == nil {
		return nil
	}
	var errs []FieldError
	for field, messages := range e.Fields {
		for _, message := range messages {
			errs = append(errs, FieldError{
				Field:   field,
				Message: message,
			})
		}
	}
	return errs
}

// NewValidationError は新しい ValidationError を生成する
func NewValidationError() *ValidationError {
	return &ValidationError{
		Global: []string{},
		Fields: make(map[string][]string),
	}
}

// AppErrorCode はアプリケーションエラーの種別を表す型
type AppErrorCode int

const (
	AppErrCodeResourceNotFound AppErrorCode = iota + 1
	AppErrCodeForbidden
	AppErrCodeConflict
	AppErrCodeInternal
)

// AppError はアプリケーションエラーを表す（SafeError パターン）。
// Error() はユーザー安全なメッセージのみを返す。
type AppError struct {
	// Code はエラー種別。Handler がステータスコードを決定するために使用する
	Code AppErrorCode
	// UserMsg はユーザーに表示する安全なメッセージ。内部情報を含めてはならない
	UserMsg string
	// Internal はログ出力用の内部エラー。ユーザーには公開しない
	Internal error
	// Metadata は構造化ログ用のメタデータ（user_id, email 等）
	Metadata map[string]string
}

func (e *AppError) Error() string { return e.UserMsg }

// Unwrap は内部エラーを返す（errors.Is / errors.As チェーン用）
func (e *AppError) Unwrap() error { return e.Internal }

// LogString はログ出力用の詳細文字列を返す
func (e *AppError) LogString() string {
	return fmt.Sprintf("Code: %d | Msg: %s | Cause: %v | Meta: %v",
		e.Code, e.UserMsg, e.Internal, e.Metadata)
}

// NewAppError は新しい AppError を生成する
func NewAppError(code AppErrorCode, userMsg string, internal error) *AppError {
	return &AppError{
		Code:     code,
		UserMsg:  userMsg,
		Internal: internal,
	}
}

// AsValidationError は err から *ValidationError を取り出す。
// 取り出せない場合は nil を返す。
func AsValidationError(err error) *ValidationError {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve
	}
	return nil
}

// AsAppError は err から *AppError を取り出す。
// 取り出せない場合は nil を返す。
func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}
