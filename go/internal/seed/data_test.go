package seed

import (
	"math/rand"
	"strings"
	"testing"
)

func TestGenerateAnimeTitle(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	// 100件のタイトルを生成して、すべてが空でないことを確認
	titles := make(map[string]bool)
	for i := 0; i < 100; i++ {
		title := GenerateAnimeTitle(r)

		// タイトルが空でないことを確認
		if title == "" {
			t.Errorf("GenerateAnimeTitle() returned empty string")
		}

		// 重複がないことを確認（ランダム性のチェック）
		titles[title] = true
	}

	// 少なくとも50件以上の異なるタイトルが生成されることを確認
	if len(titles) < 50 {
		t.Errorf("GenerateAnimeTitle() generated too few unique titles: got %d, want at least 50", len(titles))
	}
}

func TestGenerateSeasonYear(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	// 100回生成して、すべて2020〜2025年の範囲内であることを確認
	for i := 0; i < 100; i++ {
		year := GenerateSeasonYear(r)

		if year < 2020 || year > 2025 {
			t.Errorf("GenerateSeasonYear() = %d, want between 2020 and 2025", year)
		}
	}
}

func TestGenerateSeasonName(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	// すべてのシーズンが生成される可能性があることを確認
	seasons := make(map[SeasonName]bool)
	for i := 0; i < 1000; i++ {
		season := GenerateSeasonName(r)
		seasons[season] = true

		// 有効なシーズン名であることを確認
		found := false
		for _, s := range AllSeasons {
			if season == s {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GenerateSeasonName() = %s, not in AllSeasons", season)
		}
	}

	// すべてのシーズンが生成されたことを確認
	if len(seasons) != len(AllSeasons) {
		t.Errorf("GenerateSeasonName() did not generate all seasons: got %d, want %d", len(seasons), len(AllSeasons))
	}
}

func TestGenerateMediaType(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	t.Run("weighted=false", func(t *testing.T) {
		// すべてのメディアタイプが均等に生成される可能性があることを確認
		mediaTypes := make(map[MediaType]bool)
		for i := 0; i < 1000; i++ {
			mediaType := GenerateMediaType(r, false)
			mediaTypes[mediaType] = true

			// 有効なメディアタイプであることを確認
			found := false
			for _, mt := range AllMediaTypes {
				if mediaType == mt {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GenerateMediaType() = %s, not in AllMediaTypes", mediaType)
			}
		}

		// すべてのメディアタイプが生成されたことを確認
		if len(mediaTypes) != len(AllMediaTypes) {
			t.Errorf("GenerateMediaType(weighted=false) did not generate all media types: got %d, want %d", len(mediaTypes), len(AllMediaTypes))
		}
	})

	t.Run("weighted=true", func(t *testing.T) {
		// 加重ランダム: TVアニメが最も多く生成されることを確認
		counts := make(map[MediaType]int)
		total := 10000
		for i := 0; i < total; i++ {
			mediaType := GenerateMediaType(r, true)
			counts[mediaType]++
		}

		// TVアニメが最も多いことを確認（目安として50%以上）
		tvRatio := float64(counts[MediaTV]) / float64(total)
		if tvRatio < 0.5 {
			t.Errorf("GenerateMediaType(weighted=true): TV ratio = %.2f, want >= 0.5", tvRatio)
		}

		// すべてのメディアタイプが少なくとも1回は生成されることを確認
		if len(counts) != len(AllMediaTypes) {
			t.Errorf("GenerateMediaType(weighted=true) did not generate all media types: got %d, want %d", len(counts), len(AllMediaTypes))
		}
	})
}

func TestGenerateUsername(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	t.Run("without number", func(t *testing.T) {
		username := GenerateUsername(r, 0)

		// ユーザー名が空でないことを確認
		if username == "" {
			t.Errorf("GenerateUsername(0) returned empty string")
		}

		// アンダースコアが含まれることを確認（形式: adj_noun）
		if !strings.Contains(username, "_") {
			t.Errorf("GenerateUsername(0) = %s, want format 'adj_noun'", username)
		}

		// 数字が含まれないことを確認
		parts := strings.Split(username, "_")
		if len(parts) != 2 {
			t.Errorf("GenerateUsername(0) = %s, want 2 parts separated by '_'", username)
		}
	})

	t.Run("with number", func(t *testing.T) {
		username := GenerateUsername(r, 123)

		// ユーザー名が空でないことを確認
		if username == "" {
			t.Errorf("GenerateUsername(123) returned empty string")
		}

		// 数字が含まれることを確認
		if !strings.Contains(username, "123") {
			t.Errorf("GenerateUsername(123) = %s, want to contain '123'", username)
		}

		// アンダースコアが2つ含まれることを確認（形式: adj_noun_123）
		parts := strings.Split(username, "_")
		if len(parts) != 3 {
			t.Errorf("GenerateUsername(123) = %s, want 3 parts separated by '_'", username)
		}
	})

	t.Run("uniqueness", func(t *testing.T) {
		// 同じシードでも異なるユーザー名が生成されることを確認
		usernames := make(map[string]bool)
		for i := 1; i <= 100; i++ {
			username := GenerateUsername(r, i)
			usernames[username] = true
		}

		// 少なくとも50件以上の異なるユーザー名が生成されることを確認
		if len(usernames) < 50 {
			t.Errorf("GenerateUsername() generated too few unique usernames: got %d, want at least 50", len(usernames))
		}
	})
}

func TestAllSeasons(t *testing.T) {
	expectedSeasons := []SeasonName{
		SeasonSpring,
		SeasonSummer,
		SeasonAutumn,
		SeasonWinter,
	}

	if len(AllSeasons) != len(expectedSeasons) {
		t.Errorf("AllSeasons length = %d, want %d", len(AllSeasons), len(expectedSeasons))
	}

	for i, season := range AllSeasons {
		if season != expectedSeasons[i] {
			t.Errorf("AllSeasons[%d] = %s, want %s", i, season, expectedSeasons[i])
		}
	}
}

func TestAllMediaTypes(t *testing.T) {
	expectedMediaTypes := []MediaType{
		MediaTV,
		MediaOVA,
		MediaMovie,
		MediaWeb,
	}

	if len(AllMediaTypes) != len(expectedMediaTypes) {
		t.Errorf("AllMediaTypes length = %d, want %d", len(AllMediaTypes), len(expectedMediaTypes))
	}

	for i, mediaType := range AllMediaTypes {
		if mediaType != expectedMediaTypes[i] {
			t.Errorf("AllMediaTypes[%d] = %s, want %s", i, mediaType, expectedMediaTypes[i])
		}
	}
}

func TestGenerateJapaneseEpisodeRecordBody(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	t.Run("basic generation", func(t *testing.T) {
		// 感想文を生成
		body := GenerateJapaneseEpisodeRecordBody(r)

		// 空でないことを確認
		if body == "" {
			t.Errorf("GenerateJapaneseEpisodeRecordBody() returned empty string")
		}

		// プレースホルダーが残っていないことを確認
		if strings.Contains(body, "{character}") {
			t.Errorf("GenerateJapaneseEpisodeRecordBody() contains unprocessed placeholder: {character}")
		}
		if strings.Contains(body, "{scene}") {
			t.Errorf("GenerateJapaneseEpisodeRecordBody() contains unprocessed placeholder: {scene}")
		}
		if strings.Contains(body, "{emotion}") {
			t.Errorf("GenerateJapaneseEpisodeRecordBody() contains unprocessed placeholder: {emotion}")
		}

		// 日本語が含まれていることを確認（ひらがな、カタカナ、漢字のいずれかを含む）
		hasJapanese := false
		for _, r := range body {
			if (r >= 0x3040 && r <= 0x309F) || // ひらがな
				(r >= 0x30A0 && r <= 0x30FF) || // カタカナ
				(r >= 0x4E00 && r <= 0x9FFF) { // 漢字
				hasJapanese = true
				break
			}
		}
		if !hasJapanese {
			t.Errorf("GenerateJapaneseEpisodeRecordBody() does not contain Japanese characters")
		}

		// 文字数が妥当な範囲内であることを確認（20〜300文字程度）
		length := len([]rune(body))
		if length < 20 || length > 300 {
			t.Errorf("GenerateJapaneseEpisodeRecordBody() length = %d, want between 20 and 300", length)
		}
	})

	t.Run("variation", func(t *testing.T) {
		// 100件の感想文を生成して、バリエーションがあることを確認
		bodies := make(map[string]bool)
		for i := 0; i < 100; i++ {
			body := GenerateJapaneseEpisodeRecordBody(r)
			bodies[body] = true
		}

		// 少なくとも50件以上の異なる感想文が生成されることを確認
		if len(bodies) < 50 {
			t.Errorf("GenerateJapaneseEpisodeRecordBody() generated too few unique bodies: got %d, want at least 50", len(bodies))
		}
	})

	t.Run("template coverage", func(t *testing.T) {
		// 多数の感想文を生成して、すべてのテンプレートが使用される可能性を確認
		r := rand.New(rand.NewSource(42))
		bodies := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			body := GenerateJapaneseEpisodeRecordBody(r)
			bodies[body] = true
		}

		// テンプレート数（30）の半分以上の異なるパターンが生成されることを確認
		minExpected := len(episodeBodyTemplates) / 2
		if len(bodies) < minExpected {
			t.Errorf("GenerateJapaneseEpisodeRecordBody() generated too few unique bodies: got %d, want at least %d", len(bodies), minExpected)
		}
	})
}

func TestReplacePlaceholder(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	t.Run("basic replacement", func(t *testing.T) {
		template := "今回の話は{character}が良かったです。"
		words := []string{"主人公", "ヒロイン"}

		result := replacePlaceholder(template, "{character}", words, r)

		// プレースホルダーが置換されていることを確認
		if strings.Contains(result, "{character}") {
			t.Errorf("replacePlaceholder() did not replace placeholder: %s", result)
		}

		// 置換後の文字列に words のいずれかが含まれていることを確認
		containsWord := false
		for _, word := range words {
			if strings.Contains(result, word) {
				containsWord = true
				break
			}
		}
		if !containsWord {
			t.Errorf("replacePlaceholder() did not contain any word from the list: %s", result)
		}
	})

	t.Run("empty words", func(t *testing.T) {
		template := "今回の話は{character}が良かったです。"
		words := []string{}

		result := replacePlaceholder(template, "{character}", words, r)

		// 空のwordsリストの場合、テンプレートがそのまま返されることを確認
		if result != template {
			t.Errorf("replacePlaceholder() with empty words = %s, want %s", result, template)
		}
	})
}

func TestReplaceAll(t *testing.T) {
	tests := []struct {
		name string
		s    string
		old  string
		new  string
		want string
	}{
		{
			name: "single occurrence",
			s:    "今回の話は{character}が良かったです。",
			old:  "{character}",
			new:  "主人公",
			want: "今回の話は主人公が良かったです。",
		},
		{
			name: "multiple occurrences",
			s:    "{character}と{character}が良かったです。",
			old:  "{character}",
			new:  "主人公",
			want: "主人公と主人公が良かったです。",
		},
		{
			name: "no occurrence",
			s:    "今回の話は良かったです。",
			old:  "{character}",
			new:  "主人公",
			want: "今回の話は良かったです。",
		},
		{
			name: "empty string",
			s:    "",
			old:  "{character}",
			new:  "主人公",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceAll(tt.s, tt.old, tt.new)
			if got != tt.want {
				t.Errorf("replaceAll() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   int
	}{
		{
			name:   "found at beginning",
			s:      "{character}が良かったです。",
			substr: "{character}",
			want:   0,
		},
		{
			name:   "found at middle",
			s:      "今回の話は{character}が良かったです。",
			substr: "{character}",
			want:   15, // "今回の話は"（5文字×3バイト）の後
		},
		{
			name:   "not found",
			s:      "今回の話は良かったです。",
			substr: "{character}",
			want:   -1,
		},
		{
			name:   "empty string",
			s:      "",
			substr: "{character}",
			want:   -1,
		},
		{
			name:   "empty substr",
			s:      "今回の話は良かったです。",
			substr: "",
			want:   0, // 空文字列は常に位置0で見つかる
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexOf(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("indexOf() = %d, want %d", got, tt.want)
			}
		})
	}
}
