package model_test

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

func TestStripeSubscriptionStatus_String(t *testing.T) {
	tests := []struct {
		status   model.StripeSubscriptionStatus
		expected string
	}{
		{model.StripeSubscriptionStatusActive, "active"},
		{model.StripeSubscriptionStatusPastDue, "past_due"},
		{model.StripeSubscriptionStatusCanceled, "canceled"},
		{model.StripeSubscriptionStatusUnpaid, "unpaid"},
		{model.StripeSubscriptionStatusIncomplete, "incomplete"},
		{model.StripeSubscriptionStatusIncompleteExpired, "incomplete_expired"},
		{model.StripeSubscriptionStatusTrialing, "trialing"},
		{model.StripeSubscriptionStatusPaused, "paused"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStripeSubscriptionStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   model.StripeSubscriptionStatus
		expected bool
	}{
		{"active", model.StripeSubscriptionStatusActive, true},
		{"past_due", model.StripeSubscriptionStatusPastDue, true},
		{"canceled", model.StripeSubscriptionStatusCanceled, true},
		{"unpaid", model.StripeSubscriptionStatusUnpaid, true},
		{"incomplete", model.StripeSubscriptionStatusIncomplete, true},
		{"incomplete_expired", model.StripeSubscriptionStatusIncompleteExpired, true},
		{"trialing", model.StripeSubscriptionStatusTrialing, true},
		{"paused", model.StripeSubscriptionStatusPaused, true},
		{"invalid_status", model.StripeSubscriptionStatus("invalid"), false},
		{"empty_string", model.StripeSubscriptionStatus(""), false},
		{"random_string", model.StripeSubscriptionStatus("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStripeSubscriptionStatus_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   model.StripeSubscriptionStatus
		expected bool
	}{
		{
			name:     "active状態はアクティブ",
			status:   model.StripeSubscriptionStatusActive,
			expected: true,
		},
		{
			name:     "past_due状態はアクティブ（猶予期間）",
			status:   model.StripeSubscriptionStatusPastDue,
			expected: true,
		},
		{
			name:     "canceled状態は非アクティブ",
			status:   model.StripeSubscriptionStatusCanceled,
			expected: false,
		},
		{
			name:     "unpaid状態は非アクティブ",
			status:   model.StripeSubscriptionStatusUnpaid,
			expected: false,
		},
		{
			name:     "incomplete状態は非アクティブ",
			status:   model.StripeSubscriptionStatusIncomplete,
			expected: false,
		},
		{
			name:     "incomplete_expired状態は非アクティブ",
			status:   model.StripeSubscriptionStatusIncompleteExpired,
			expected: false,
		},
		{
			name:     "trialing状態は非アクティブ",
			status:   model.StripeSubscriptionStatusTrialing,
			expected: false,
		},
		{
			name:     "paused状態は非アクティブ",
			status:   model.StripeSubscriptionStatusPaused,
			expected: false,
		},
		{
			name:     "無効な状態は非アクティブ",
			status:   model.StripeSubscriptionStatus("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsActive(); got != tt.expected {
				t.Errorf("IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}
