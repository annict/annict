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
	// constブロックに常に1つ以上の定数を維持するためのダミー定数
	// 実際のフラグがすべて削除されてもコンパイルエラーにならないようにする
	FeatureFlagExample    FeatureFlagName = "go_example"
	FeatureFlagGoAnnictDB FeatureFlagName = "go_annict_db"
)
