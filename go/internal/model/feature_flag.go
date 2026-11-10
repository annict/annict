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

	// FeatureFlagRESTAPIV2は、REST API V2をトークンの持ち主の利用者に応答するかを決めるフラグ。
	// 移行のフラグではないため `go_` を付けない
	FeatureFlagRESTAPIV2 FeatureFlagName = "rest_api_v2"

	// FeatureFlagDBEditHistoryは、Annict DBの編集履歴の画面と画面へのリンクを出すかを決めるフラグ。
	// 編集履歴の記録はフラグにかかわらず行う。移行のフラグではないため `go_` を付けない
	FeatureFlagDBEditHistory FeatureFlagName = "db_edit_history"
)
