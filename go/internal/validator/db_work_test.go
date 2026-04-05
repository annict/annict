package validator

import (
	"context"
	"testing"
)

func TestCreateDbWorkValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      CreateDbWorkValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name: "正常系: 必須フィールドのみ",
			input: CreateDbWorkValidatorInput{
				Title: "テストアニメ",
				Media: "1",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 全フィールド入力",
			input: CreateDbWorkValidatorInput{
				Title:                 "テストアニメ",
				TitleKana:             "てすとあにめ",
				TitleAlter:            "別名",
				TitleEn:               "Test Anime",
				TitleAlterEn:          "Alt Name",
				Media:                 "1",
				SeasonYear:            "2024",
				SeasonName:            "2",
				StartedOn:             "2024-04-01",
				EndedOn:               "2024-06-30",
				OfficialSiteURL:       "https://example.com",
				OfficialSiteURLEn:     "https://example.com/en",
				WikipediaURL:          "https://ja.wikipedia.org/wiki/Test",
				WikipediaURLEn:        "https://en.wikipedia.org/wiki/Test",
				TwitterUsername:       "testanime",
				TwitterHashtag:        "テストアニメ",
				ScTid:                 "12345",
				MalAnimeID:            "54321",
				Synopsis:              "テストのあらすじ",
				SynopsisSource:        "公式サイト",
				SynopsisEn:            "Test synopsis",
				SynopsisSourceEn:      "Official site",
				ManualEpisodesCount:   "12",
				StartEpisodeRawNumber: "1",
				NumberFormatID:        "1",
				NoEpisodes:            "1",
			},
			wantErrors: false,
		},
		{
			name: "異常系: タイトルが空",
			input: CreateDbWorkValidatorInput{
				Title: "",
				Media: "1",
			},
			wantErrors: true,
			wantFields: []string{"title"},
		},
		{
			name: "異常系: タイトルがwhitespaceのみ",
			input: CreateDbWorkValidatorInput{
				Title: "   ",
				Media: "1",
			},
			wantErrors: true,
			wantFields: []string{"title"},
		},
		{
			name: "異常系: メディアが空",
			input: CreateDbWorkValidatorInput{
				Title: "テストアニメ",
				Media: "",
			},
			wantErrors: true,
			wantFields: []string{"media"},
		},
		{
			name: "異常系: メディアが不正な値",
			input: CreateDbWorkValidatorInput{
				Title: "テストアニメ",
				Media: "99",
			},
			wantErrors: true,
			wantFields: []string{"media"},
		},
		{
			name: "異常系: タイトルとメディアの両方が空",
			input: CreateDbWorkValidatorInput{
				Title: "",
				Media: "",
			},
			wantErrors: true,
			wantFields: []string{"title", "media"},
		},
		{
			name: "正常系: メディアが0（その他）",
			input: CreateDbWorkValidatorInput{
				Title: "テストアニメ",
				Media: "0",
			},
			wantErrors: false,
		},
	}

	validator := NewCreateDbWorkValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := validator.Validate(ctx, tt.input)

			if tt.wantErrors {
				if result.FormErrors == nil || !result.FormErrors.HasErrors() {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if !result.FormErrors.HasFieldError(field) {
						t.Errorf("フィールド %s のエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if result.FormErrors != nil && result.FormErrors.HasErrors() {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", result.FormErrors)
				}
			}
		})
	}
}

func TestCreateDbWorkValidatorValidate_URL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      CreateDbWorkValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name: "正常系: URLが空（スキップ）",
			input: CreateDbWorkValidatorInput{
				Title:           "テストアニメ",
				Media:           "1",
				OfficialSiteURL: "",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 有効なhttps URL",
			input: CreateDbWorkValidatorInput{
				Title:           "テストアニメ",
				Media:           "1",
				OfficialSiteURL: "https://example.com",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 有効なhttp URL",
			input: CreateDbWorkValidatorInput{
				Title:           "テストアニメ",
				Media:           "1",
				OfficialSiteURL: "http://example.com",
			},
			wantErrors: false,
		},
		{
			name: "異常系: スキームなしのURL",
			input: CreateDbWorkValidatorInput{
				Title:           "テストアニメ",
				Media:           "1",
				OfficialSiteURL: "example.com",
			},
			wantErrors: true,
			wantFields: []string{"official_site_url"},
		},
		{
			name: "異常系: 不正なURL",
			input: CreateDbWorkValidatorInput{
				Title:           "テストアニメ",
				Media:           "1",
				OfficialSiteURL: "not-a-url",
			},
			wantErrors: true,
			wantFields: []string{"official_site_url"},
		},
		{
			name: "異常系: ftpスキーム",
			input: CreateDbWorkValidatorInput{
				Title:           "テストアニメ",
				Media:           "1",
				OfficialSiteURL: "ftp://example.com/file",
			},
			wantErrors: true,
			wantFields: []string{"official_site_url"},
		},
		{
			name: "異常系: 複数のURL項目でエラー",
			input: CreateDbWorkValidatorInput{
				Title:             "テストアニメ",
				Media:             "1",
				OfficialSiteURL:   "invalid",
				OfficialSiteURLEn: "invalid",
				WikipediaURL:      "invalid",
				WikipediaURLEn:    "invalid",
			},
			wantErrors: true,
			wantFields: []string{"official_site_url", "official_site_url_en", "wikipedia_url", "wikipedia_url_en"},
		},
	}

	validator := NewCreateDbWorkValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := validator.Validate(ctx, tt.input)

			if tt.wantErrors {
				if result.FormErrors == nil || !result.FormErrors.HasErrors() {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if !result.FormErrors.HasFieldError(field) {
						t.Errorf("フィールド %s のエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if result.FormErrors != nil && result.FormErrors.HasErrors() {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", result.FormErrors)
				}
			}
		})
	}
}

func TestCreateDbWorkValidatorValidate_NumericFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      CreateDbWorkValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name: "正常系: sc_tidが空（スキップ）",
			input: CreateDbWorkValidatorInput{
				Title: "テストアニメ",
				Media: "1",
				ScTid: "",
			},
			wantErrors: false,
		},
		{
			name: "正常系: sc_tidが有効な整数",
			input: CreateDbWorkValidatorInput{
				Title: "テストアニメ",
				Media: "1",
				ScTid: "12345",
			},
			wantErrors: false,
		},
		{
			name: "異常系: sc_tidが整数でない",
			input: CreateDbWorkValidatorInput{
				Title: "テストアニメ",
				Media: "1",
				ScTid: "abc",
			},
			wantErrors: true,
			wantFields: []string{"sc_tid"},
		},
		{
			name: "異常系: sc_tidが小数",
			input: CreateDbWorkValidatorInput{
				Title: "テストアニメ",
				Media: "1",
				ScTid: "12.5",
			},
			wantErrors: true,
			wantFields: []string{"sc_tid"},
		},
		{
			name: "正常系: mal_anime_idが有効な整数",
			input: CreateDbWorkValidatorInput{
				Title:      "テストアニメ",
				Media:      "1",
				MalAnimeID: "54321",
			},
			wantErrors: false,
		},
		{
			name: "異常系: mal_anime_idが整数でない",
			input: CreateDbWorkValidatorInput{
				Title:      "テストアニメ",
				Media:      "1",
				MalAnimeID: "xyz",
			},
			wantErrors: true,
			wantFields: []string{"mal_anime_id"},
		},
	}

	validator := NewCreateDbWorkValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := validator.Validate(ctx, tt.input)

			if tt.wantErrors {
				if result.FormErrors == nil || !result.FormErrors.HasErrors() {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if !result.FormErrors.HasFieldError(field) {
						t.Errorf("フィールド %s のエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if result.FormErrors != nil && result.FormErrors.HasErrors() {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", result.FormErrors)
				}
			}
		})
	}
}

func TestCreateDbWorkValidatorValidate_PresencePair(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      CreateDbWorkValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name: "正常系: あらすじと出典の両方が空",
			input: CreateDbWorkValidatorInput{
				Title:    "テストアニメ",
				Media:    "1",
				Synopsis: "",
			},
			wantErrors: false,
		},
		{
			name: "正常系: あらすじと出典の両方がある",
			input: CreateDbWorkValidatorInput{
				Title:          "テストアニメ",
				Media:          "1",
				Synopsis:       "テストのあらすじ",
				SynopsisSource: "公式サイト",
			},
			wantErrors: false,
		},
		{
			name: "異常系: あらすじのみで出典がない",
			input: CreateDbWorkValidatorInput{
				Title:          "テストアニメ",
				Media:          "1",
				Synopsis:       "テストのあらすじ",
				SynopsisSource: "",
			},
			wantErrors: true,
			wantFields: []string{"synopsis_source"},
		},
		{
			name: "正常系: 出典のみ（あらすじなし）は許可",
			input: CreateDbWorkValidatorInput{
				Title:          "テストアニメ",
				Media:          "1",
				Synopsis:       "",
				SynopsisSource: "公式サイト",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 英語あらすじと出典の両方がある",
			input: CreateDbWorkValidatorInput{
				Title:            "テストアニメ",
				Media:            "1",
				SynopsisEn:       "Test synopsis",
				SynopsisSourceEn: "Official site",
			},
			wantErrors: false,
		},
		{
			name: "異常系: 英語あらすじのみで出典がない",
			input: CreateDbWorkValidatorInput{
				Title:            "テストアニメ",
				Media:            "1",
				SynopsisEn:       "Test synopsis",
				SynopsisSourceEn: "",
			},
			wantErrors: true,
			wantFields: []string{"synopsis_source_en"},
		},
		{
			name: "異常系: 日英両方のあらすじに出典がない",
			input: CreateDbWorkValidatorInput{
				Title:      "テストアニメ",
				Media:      "1",
				Synopsis:   "テストのあらすじ",
				SynopsisEn: "Test synopsis",
			},
			wantErrors: true,
			wantFields: []string{"synopsis_source", "synopsis_source_en"},
		},
	}

	validator := NewCreateDbWorkValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := validator.Validate(ctx, tt.input)

			if tt.wantErrors {
				if result.FormErrors == nil || !result.FormErrors.HasErrors() {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if !result.FormErrors.HasFieldError(field) {
						t.Errorf("フィールド %s のエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if result.FormErrors != nil && result.FormErrors.HasErrors() {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", result.FormErrors)
				}
			}
		})
	}
}
