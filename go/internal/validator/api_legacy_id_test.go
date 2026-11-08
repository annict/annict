package validator

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

func TestAPILegacyIDIndexValidator_Validate(t *testing.T) {
	t.Parallel()

	manyWorkIDs := make([]model.WorkID, 60)
	manyEpisodeIDs := make([]model.EpisodeID, 40)
	tests := []struct {
		name        string
		input       APILegacyIDIndexValidatorInput
		wantMessage string
	}{
		{name: "作品だけを指定できる", input: APILegacyIDIndexValidatorInput{WorkIDs: []model.WorkID{1}}},
		{name: "エピソードだけを指定できる", input: APILegacyIDIndexValidatorInput{EpisodeIDs: []model.EpisodeID{1}}},
		{name: "種類をまたいで合計100件まで指定できる", input: APILegacyIDIndexValidatorInput{WorkIDs: manyWorkIDs, EpisodeIDs: manyEpisodeIDs}},
		{name: "旧IDが1つも無いとエラー", input: APILegacyIDIndexValidatorInput{}, wantMessage: "旧IDを1つ以上指定してください (work_ids・episode_ids)"},
		{
			name:        "種類をまたいで合計100件を超えるとエラー",
			input:       APILegacyIDIndexValidatorInput{WorkIDs: manyWorkIDs, EpisodeIDs: append(manyEpisodeIDs, 1)},
			wantMessage: "旧IDは合計100件まで指定できます",
		},
	}

	v := NewAPILegacyIDIndexValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")
			err := v.Validate(ctx, tt.input)
			if tt.wantMessage == "" {
				if err != nil {
					t.Errorf("Validate()のエラー = %v", err)
				}
				return
			}
			ve := model.AsValidationError(err)
			if ve == nil || len(ve.Global) != 1 || ve.Global[0] != tt.wantMessage {
				t.Errorf("エラー = %+v、期待値 = %q のグローバルエラー", ve, tt.wantMessage)
			}
		})
	}
}
