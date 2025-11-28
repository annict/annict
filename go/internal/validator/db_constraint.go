// Package validator はバリデーション機能を提供します
package validator

import (
	"database/sql"
	"fmt"
)

// ValidateNotNullTime はsql.NullTimeフィールドがNULLでないことを検証します。
// データベーススキーマでNOT NULL制約があるにも関わらず、Goの型がsql.NullTimeの場合に使用します。
//
// # 引数
//   - value: 検証するsql.NullTime値
//   - fieldName: フィールド名（エラーメッセージに使用）
//   - entityID: エンティティのID（エラーメッセージに使用）
//
// # 戻り値
//   - error: NULLの場合はエラー、それ以外はnil
//
// # 使用例
//
//	if err := validator.ValidateNotNullTime(user.CreatedAt, "created_at", user.ID); err != nil {
//	    return nil, err
//	}
func ValidateNotNullTime(value sql.NullTime, fieldName string, entityID int64) error {
	if !value.Valid {
		return fmt.Errorf("%sがNULLです（データベース制約違反）: id=%d", fieldName, entityID)
	}
	return nil
}

// ValidateNotNullString はsql.NullStringフィールドがNULLでないことを検証します。
// データベーススキーマでNOT NULL制約があるにも関わらず、Goの型がsql.NullStringの場合に使用します。
//
// # 引数
//   - value: 検証するsql.NullString値
//   - fieldName: フィールド名（エラーメッセージに使用）
//   - entityID: エンティティのID（エラーメッセージに使用）
//
// # 戻り値
//   - error: NULLの場合はエラー、それ以外はnil
func ValidateNotNullString(value sql.NullString, fieldName string, entityID int64) error {
	if !value.Valid {
		return fmt.Errorf("%sがNULLです（データベース制約違反）: id=%d", fieldName, entityID)
	}
	return nil
}

// ValidateNotNullInt32 はsql.NullInt32フィールドがNULLでないことを検証します。
// データベーススキーマでNOT NULL制約があるにも関わらず、Goの型がsql.NullInt32の場合に使用します。
//
// # 引数
//   - value: 検証するsql.NullInt32値
//   - fieldName: フィールド名（エラーメッセージに使用）
//   - entityID: エンティティのID（エラーメッセージに使用）
//
// # 戻り値
//   - error: NULLの場合はエラー、それ以外はnil
func ValidateNotNullInt32(value sql.NullInt32, fieldName string, entityID int64) error {
	if !value.Valid {
		return fmt.Errorf("%sがNULLです（データベース制約違反）: id=%d", fieldName, entityID)
	}
	return nil
}

// ValidateNotNullInt64 はsql.NullInt64フィールドがNULLでないことを検証します。
// データベーススキーマでNOT NULL制約があるにも関わらず、Goの型がsql.NullInt64の場合に使用します。
//
// # 引数
//   - value: 検証するsql.NullInt64値
//   - fieldName: フィールド名（エラーメッセージに使用）
//   - entityID: エンティティのID（エラーメッセージに使用）
//
// # 戻り値
//   - error: NULLの場合はエラー、それ以外はnil
func ValidateNotNullInt64(value sql.NullInt64, fieldName string, entityID int64) error {
	if !value.Valid {
		return fmt.Errorf("%sがNULLです（データベース制約違反）: id=%d", fieldName, entityID)
	}
	return nil
}
