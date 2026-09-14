package model

import "time"

// FeatureFlagはフィーチャーフラグを表すドメインモデル。
type FeatureFlag struct {
	ID          FeatureFlagID
	DeviceToken *string
	UserID      *UserID
	Name        FeatureFlagName
	CreatedAt   time.Time
}

// RailsからGoへの移行で使用するフラグ名には `go_` プレフィックスを付ける。
// 実際のフラグをすべて削除した後もファイルがコンパイルできるよう、
// FeatureFlagExampleをconstブロックに常に1つ以上残しておく。
const (
	FeatureFlagExample FeatureFlagName = "go_example"
)
