package testutil

import (
	"database/sql"
	"testing"
)

// FeatureFlagBuilder はフィーチャーフラグのテストデータビルダー
type FeatureFlagBuilder struct {
	t           *testing.T
	tx          *sql.Tx
	deviceToken *string
	userID      *int64
	name        string
}

// NewFeatureFlagBuilder は新しいFeatureFlagBuilderを作成
func NewFeatureFlagBuilder(t *testing.T, tx *sql.Tx) *FeatureFlagBuilder {
	return &FeatureFlagBuilder{
		t:    t,
		tx:   tx,
		name: "test_flag",
	}
}

// WithDeviceToken はデバイストークンを設定
func (b *FeatureFlagBuilder) WithDeviceToken(token string) *FeatureFlagBuilder {
	b.deviceToken = &token
	return b
}

// WithUserID はユーザーIDを設定
func (b *FeatureFlagBuilder) WithUserID(userID int64) *FeatureFlagBuilder {
	b.userID = &userID
	return b
}

// WithName はフラグ名を設定
func (b *FeatureFlagBuilder) WithName(name string) *FeatureFlagBuilder {
	b.name = name
	return b
}

// Build はテスト用のフィーチャーフラグデータをデータベースに作成し、IDを返す
func (b *FeatureFlagBuilder) Build() int64 {
	b.t.Helper()

	var id int64
	err := b.tx.QueryRow(
		`INSERT INTO feature_flags (device_token, user_id, name) VALUES ($1, $2, $3) RETURNING id`,
		b.deviceToken,
		b.userID,
		b.name,
	).Scan(&id)
	if err != nil {
		b.t.Fatalf("フィーチャーフラグデータの作成に失敗しました: %v", err)
	}

	return id
}
