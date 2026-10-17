package model_test

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

func TestPersonStatus_IsViewable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.PersonStatus
		want   bool
	}{
		{status: model.PersonStatusPublished, want: true},
		{status: model.PersonStatusArchived, want: true},
		{status: model.PersonStatusDelisted, want: false},
		{status: model.PersonStatusDeleted, want: false},
	}

	for _, tt := range tests {
		if got := tt.status.IsViewable(); got != tt.want {
			t.Errorf("PersonStatus(%q).IsViewable() = %v、期待値 = %v", tt.status, got, tt.want)
		}
	}
}

func TestPersonGender_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		gender model.PersonGender
		want   bool
	}{
		{gender: model.PersonGenderMale, want: true},
		{gender: model.PersonGenderFemale, want: true},
		{gender: "", want: false},
		{gender: "other", want: false},
	}

	for _, tt := range tests {
		if got := tt.gender.IsValid(); got != tt.want {
			t.Errorf("PersonGender(%q).IsValid() = %v、期待値 = %v", tt.gender, got, tt.want)
		}
	}
}

func TestPersonBloodType_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		bloodType model.PersonBloodType
		want      bool
	}{
		{bloodType: model.PersonBloodTypeA, want: true},
		{bloodType: model.PersonBloodTypeB, want: true},
		{bloodType: model.PersonBloodTypeAB, want: true},
		{bloodType: model.PersonBloodTypeO, want: true},
		{bloodType: "", want: false},
		{bloodType: "A", want: false},
	}

	for _, tt := range tests {
		if got := tt.bloodType.IsValid(); got != tt.want {
			t.Errorf("PersonBloodType(%q).IsValid() = %v、期待値 = %v", tt.bloodType, got, tt.want)
		}
	}
}
