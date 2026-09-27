package model

import (
	"errors"
	"fmt"
)

// ValidationErrorはバリデーションエラーを表す。
// Handlerはこのエラーを受け取ったらフォームを再描画する (422)。
type ValidationError struct {
	// Globalはフォーム全体のエラーメッセージ (特定のフィールドに紐づかないエラー)
	Global []string `json:"global,omitempty"`
	// Fieldsはフィールドごとのエラーメッセージ
	Fields map[string][]string `json:"fields,omitempty"`
}

func (e *ValidationError) Error() string { return "validation failed" }

// AddGlobalはグローバルエラーを追加する
func (e *ValidationError) AddGlobal(message string) {
	e.Global = append(e.Global, message)
}

// AddFieldはフィールドエラーを追加する
func (e *ValidationError) AddField(field, message string) {
	if e.Fields == nil {
		e.Fields = make(map[string][]string)
	}
	e.Fields[field] = append(e.Fields[field], message)
}

// HasErrorsはエラーがあるかどうかを返す
func (e *ValidationError) HasErrors() bool {
	if e == nil {
		return false
	}
	return len(e.Global) > 0 || len(e.Fields) > 0
}

// HasFieldErrorは指定されたフィールドにエラーがあるかどうかを返す
func (e *ValidationError) HasFieldError(field string) bool {
	if e == nil || e.Fields == nil {
		return false
	}
	return len(e.Fields[field]) > 0
}

// GetFieldErrorsは指定されたフィールドのエラーメッセージを返す
func (e *ValidationError) GetFieldErrors(field string) []string {
	if e == nil || e.Fields == nil {
		return nil
	}
	return e.Fields[field]
}

// FieldErrorはフィールドエラーを表す構造体 (テンプレート用)
type FieldError struct {
	Field   string
	Message string
}

// FieldErrorsはフィールドエラーを列挙可能な形式で取得する
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

// NewValidationErrorは新しいValidationErrorを生成する
func NewValidationError() *ValidationError {
	return &ValidationError{
		Global: []string{},
		Fields: make(map[string][]string),
	}
}

// AppErrorCodeはアプリケーションエラーの種別を表す型
type AppErrorCode int

const (
	AppErrCodeResourceNotFound AppErrorCode = iota + 1
	AppErrCodeForbidden
	AppErrCodeConflict
	AppErrCodeInternal
	// AppErrCodeBusyは、リクエストが必要としたものを他の書き込みが保持していたため適用
	// されなかったことを表す。AppErrCodeConflictと違い、保存済みの状態と食い違ったことを意味
	// しない。何も書かれておらず、同じリクエストをやり直せば成功しうる。
	AppErrCodeBusy
)

// AppErrorはアプリケーションエラーを表す (SafeErrorパターン)。
// Error() はユーザー安全なメッセージのみを返す。
type AppError struct {
	// Codeはエラー種別。Handlerがステータスコードを決定するために使用する
	Code AppErrorCode
	// UserMsgはユーザーに表示する安全なメッセージ。内部情報を含めてはならない
	UserMsg string
	// Internalはログ出力用の内部エラー。ユーザーには公開しない
	Internal error
	// Metadataは構造化ログ用のメタデータ (user_id, email等)
	Metadata map[string]string
}

func (e *AppError) Error() string { return e.UserMsg }

// Unwrapは内部エラーを返す (errors.Is / errors.Asチェーン用)
func (e *AppError) Unwrap() error { return e.Internal }

// LogStringはログ出力用の詳細文字列を返す
func (e *AppError) LogString() string {
	return fmt.Sprintf("Code: %d | Msg: %s | Cause: %v | Meta: %v",
		e.Code, e.UserMsg, e.Internal, e.Metadata)
}

// NewAppErrorは新しいAppErrorを生成する
func NewAppError(code AppErrorCode, userMsg string, internal error) *AppError {
	return &AppError{
		Code:     code,
		UserMsg:  userMsg,
		Internal: internal,
	}
}

// AsValidationErrorはerrから *ValidationErrorを取り出す。
// 取り出せない場合はnilを返す。
func AsValidationError(err error) *ValidationError {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve
	}
	return nil
}

// AsAppErrorはerrから *AppErrorを取り出す。
// 取り出せない場合はnilを返す。
func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}
