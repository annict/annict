package model

import "time"

// FeatureFlag はフィーチャーフラグを表すドメインモデル
type FeatureFlag struct {
	ID          FeatureFlagID
	DeviceToken *string
	UserID      *int64
	Name        FeatureFlagName
	CreatedAt   time.Time
}

// フラグ名の定数
// Go版への移行で使用するフラグには go_ プレフィックスを付ける
const (
// 今後のGo移行タスクで追加される
// 例: FeatureFlagGoPageEdit FeatureFlagName = "go_page_edit"
)
