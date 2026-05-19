package model

import "time"

// FeatureFlag is the domain model that represents a single feature flag entry.
// [Ja] FeatureFlag はフィーチャーフラグを表すドメインモデル。
type FeatureFlag struct {
	ID          FeatureFlagID
	DeviceToken *string
	UserID      *UserID
	Name        FeatureFlagName
	CreatedAt   time.Time
}

// Flag names used for the Rails-to-Go migration are prefixed with `go_`. The
// const block keeps at least one constant defined at all times: even after
// every real flag has been cleaned up, FeatureFlagExample stays as a
// placeholder so the file still compiles.
//
// [Ja] Rails から Go への移行で使用するフラグ名には `go_` プレフィックスを付ける。
// 実際のフラグをすべて削除した後もファイルがコンパイルできるよう、
// FeatureFlagExample を const ブロックに常に 1 つ以上残しておく。
const (
	FeatureFlagExample    FeatureFlagName = "go_example"
	FeatureFlagGoAnnictDB FeatureFlagName = "go_annict_db"
)
