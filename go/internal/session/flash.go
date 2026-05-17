// Package session はセッション管理機能を提供します
package session

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
