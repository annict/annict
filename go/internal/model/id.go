package model

import "fmt"

// FeatureFlagID はフィーチャーフラグのID型
type FeatureFlagID int64

// String は文字列表現を返す
func (id FeatureFlagID) String() string { return fmt.Sprintf("%d", id) }

// FeatureFlagName はフィーチャーフラグ名の型
type FeatureFlagName string

// String は文字列表現を返す
func (n FeatureFlagName) String() string { return string(n) }
