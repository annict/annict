package validator

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestValidateNotNullTime(t *testing.T) {
	tests := []struct {
		name      string
		value     sql.NullTime
		fieldName string
		entityID  int64
		wantError bool
	}{
		{
			name: "正常系: Validなsql.NullTime",
			value: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			fieldName: "created_at",
			entityID:  1,
			wantError: false,
		},
		{
			name: "異常系: NULL（Valid=false）",
			value: sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			},
			fieldName: "updated_at",
			entityID:  123,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotNullTime(tt.value, tt.fieldName, tt.entityID)

			if tt.wantError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				} else {
					// エラーメッセージにフィールド名とIDが含まれているか確認
					errMsg := err.Error()
					if !strings.Contains(errMsg, tt.fieldName) {
						t.Errorf("エラーメッセージにフィールド名が含まれていません: %s", errMsg)
					}
					if !strings.Contains(errMsg, "データベース制約違反") {
						t.Errorf("エラーメッセージに「データベース制約違反」が含まれていません: %s", errMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("エラーが期待されていませんが、エラーが返されました: %v", err)
				}
			}
		})
	}
}

func TestValidateNotNullString(t *testing.T) {
	tests := []struct {
		name      string
		value     sql.NullString
		fieldName string
		entityID  int64
		wantError bool
	}{
		{
			name: "正常系: Validなsql.NullString",
			value: sql.NullString{
				String: "test",
				Valid:  true,
			},
			fieldName: "username",
			entityID:  1,
			wantError: false,
		},
		{
			name: "異常系: NULL（Valid=false）",
			value: sql.NullString{
				String: "",
				Valid:  false,
			},
			fieldName: "email",
			entityID:  456,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotNullString(tt.value, tt.fieldName, tt.entityID)

			if tt.wantError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("エラーが期待されていませんが、エラーが返されました: %v", err)
				}
			}
		})
	}
}

func TestValidateNotNullInt32(t *testing.T) {
	tests := []struct {
		name      string
		value     sql.NullInt32
		fieldName string
		entityID  int64
		wantError bool
	}{
		{
			name: "正常系: Validなsql.NullInt32",
			value: sql.NullInt32{
				Int32: 42,
				Valid: true,
			},
			fieldName: "count",
			entityID:  1,
			wantError: false,
		},
		{
			name: "異常系: NULL（Valid=false）",
			value: sql.NullInt32{
				Int32: 0,
				Valid: false,
			},
			fieldName: "status",
			entityID:  789,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotNullInt32(tt.value, tt.fieldName, tt.entityID)

			if tt.wantError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("エラーが期待されていませんが、エラーが返されました: %v", err)
				}
			}
		})
	}
}

func TestValidateNotNullInt64(t *testing.T) {
	tests := []struct {
		name      string
		value     sql.NullInt64
		fieldName string
		entityID  int64
		wantError bool
	}{
		{
			name: "正常系: Validなsql.NullInt64",
			value: sql.NullInt64{
				Int64: 12345,
				Valid: true,
			},
			fieldName: "total",
			entityID:  1,
			wantError: false,
		},
		{
			name: "異常系: NULL（Valid=false）",
			value: sql.NullInt64{
				Int64: 0,
				Valid: false,
			},
			fieldName: "amount",
			entityID:  999,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotNullInt64(tt.value, tt.fieldName, tt.entityID)

			if tt.wantError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("エラーが期待されていませんが、エラーが返されました: %v", err)
				}
			}
		})
	}
}
