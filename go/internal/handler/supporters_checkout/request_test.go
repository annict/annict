package supporters_checkout

import "testing"

// TestCreateRequest_Validate はリクエストバリデーションのテスト
func TestCreateRequest_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		plan        string
		expectError bool
	}{
		{"monthly plan", "monthly", false},
		{"yearly plan", "yearly", false},
		{"empty plan", "", true},
		{"invalid plan", "weekly", true},
		{"uppercase plan", "MONTHLY", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &CreateRequest{Plan: tc.plan}
			errors := req.Validate()

			if tc.expectError && len(errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tc.expectError && len(errors) > 0 {
				t.Errorf("expected no validation error but got: %v", errors)
			}
		})
	}
}
