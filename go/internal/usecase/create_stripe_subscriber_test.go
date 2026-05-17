package usecase

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

func TestParseUserIDFromMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		metadata  map[string]string
		wantID    model.UserID
		wantError bool
		errorType string
	}{
		{
			name:      "user_idが存在しない場合はエラー",
			metadata:  map[string]string{},
			wantID:    0,
			wantError: true,
			errorType: "MetadataUserIDMissingError",
		},
		{
			name:      "user_idが空の場合はエラー",
			metadata:  map[string]string{"user_id": ""},
			wantID:    0,
			wantError: true,
			errorType: "MetadataUserIDInvalidError",
		},
		{
			name:      "user_idが数値でない場合はエラー",
			metadata:  map[string]string{"user_id": "abc"},
			wantID:    0,
			wantError: true,
			errorType: "MetadataUserIDInvalidError",
		},
		{
			name:      "user_idが正しい場合は成功",
			metadata:  map[string]string{"user_id": "12345"},
			wantID:    12345,
			wantError: false,
		},
		{
			name:      "user_idが0でも成功",
			metadata:  map[string]string{"user_id": "0"},
			wantID:    0,
			wantError: false,
		},
		{
			name:      "user_idが負の値でも成功",
			metadata:  map[string]string{"user_id": "-1"},
			wantID:    -1,
			wantError: false,
		},
		{
			name:      "他のキーが含まれていても成功",
			metadata:  map[string]string{"user_id": "999", "plan": "monthly"},
			wantID:    999,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userID, err := ParseUserIDFromMetadata(tt.metadata)

			if tt.wantError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、nilが返されました")
					return
				}

				switch tt.errorType {
				case "MetadataUserIDMissingError":
					if !IsMetadataUserIDMissingError(err) {
						t.Errorf("MetadataUserIDMissingErrorが期待されましたが、%v が返されました", err)
					}
				case "MetadataUserIDInvalidError":
					if !IsMetadataUserIDInvalidError(err) {
						t.Errorf("MetadataUserIDInvalidErrorが期待されましたが、%v が返されました", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}

			if userID != tt.wantID {
				t.Errorf("userID: got %d, want %d", userID, tt.wantID)
			}
		})
	}
}

func TestInvalidSubscriptionStatusError(t *testing.T) {
	t.Parallel()

	err := &InvalidSubscriptionStatusError{Status: "unknown"}

	// Error()メソッドのテスト
	expected := "invalid subscription status: unknown"
	if err.Error() != expected {
		t.Errorf("Error(): got %q, want %q", err.Error(), expected)
	}

	// IsInvalidSubscriptionStatusErrorのテスト
	if !IsInvalidSubscriptionStatusError(err) {
		t.Error("IsInvalidSubscriptionStatusError: got false, want true")
	}
}

func TestMetadataUserIDMissingError(t *testing.T) {
	t.Parallel()

	err := &MetadataUserIDMissingError{}

	// Error()メソッドのテスト
	expected := "user_id is missing from metadata"
	if err.Error() != expected {
		t.Errorf("Error(): got %q, want %q", err.Error(), expected)
	}

	// IsMetadataUserIDMissingErrorのテスト
	if !IsMetadataUserIDMissingError(err) {
		t.Error("IsMetadataUserIDMissingError: got false, want true")
	}
}

func TestMetadataUserIDInvalidError(t *testing.T) {
	t.Parallel()

	err := &MetadataUserIDInvalidError{Value: "abc"}

	// Error()メソッドのテスト
	expected := "invalid user_id in metadata: abc"
	if err.Error() != expected {
		t.Errorf("Error(): got %q, want %q", err.Error(), expected)
	}

	// IsMetadataUserIDInvalidErrorのテスト
	if !IsMetadataUserIDInvalidError(err) {
		t.Error("IsMetadataUserIDInvalidError: got false, want true")
	}
}
