package model_test

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

func TestAnimeSeason_SortKey(t *testing.T) {
	t.Parallel()

	name := func(n model.SeasonName) *model.SeasonName { return &n }
	// 冬・春・夏・秋・不明の順で、翌年の冬より前になる
	seasons := []model.AnimeSeason{
		{Year: 2025, Name: name(model.SeasonNameWinter)},
		{Year: 2025, Name: name(model.SeasonNameSpring)},
		{Year: 2025, Name: name(model.SeasonNameSummer)},
		{Year: 2025, Name: name(model.SeasonNameFall)},
		{Year: 2025},
		{Year: 2026, Name: name(model.SeasonNameWinter)},
	}
	for i := 1; i < len(seasons); i++ {
		if prev, cur := seasons[i-1].SortKey(), seasons[i].SortKey(); prev >= cur {
			t.Errorf("%d番めのキー = %d、前のキー = %d、期待値 = 前より大きい", i, cur, prev)
		}
	}
	if got := seasons[1].SortKey(); got != 20252 {
		t.Errorf("2025年春のキー = %d、期待値 = 20252", got)
	}
}

func TestParseSeasonBucket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		wantErr bool
	}{
		{input: "2026-spring"},
		{input: "2026-unknown"},
		{input: "2026-autumn", wantErr: true},
		{input: "2026", wantErr: true},
		{input: "spring-2026", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			bucket, err := model.ParseSeasonBucket(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("エラーを期待したが、nilだった (結果 = %+v)", bucket)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSeasonBucket()のエラー = %v", err)
			}
			if got := bucket.String(); got != tt.input {
				t.Errorf("String() = %q、期待値 = %q", got, tt.input)
			}
		})
	}
}
