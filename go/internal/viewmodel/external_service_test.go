package viewmodel

import "testing"

func TestSyobocalURL(t *testing.T) {
	t.Parallel()

	if got, want := SyobocalURL(3524), "http://cal.syoboi.jp/tid/3524"; got != want {
		t.Errorf("SyobocalURL(3524) = %q、期待値 = %q", got, want)
	}
}

func TestMalAnimeURL(t *testing.T) {
	t.Parallel()

	if got, want := MalAnimeURL(20), "https://myanimelist.net/anime/20"; got != want {
		t.Errorf("MalAnimeURL(20) = %q、期待値 = %q", got, want)
	}
}

func TestTwitterUsernameURL(t *testing.T) {
	t.Parallel()

	if got, want := TwitterUsernameURL("annict_com"), "https://x.com/annict_com"; got != want {
		t.Errorf("TwitterUsernameURL(%q) = %q、期待値 = %q", "annict_com", got, want)
	}
	if got := TwitterUsernameURL(""); got != "" {
		t.Errorf("TwitterUsernameURL(\"\") = %q、期待値 = \"\"", got)
	}
}

func TestTwitterHashtagURL(t *testing.T) {
	t.Parallel()

	if got, want := TwitterHashtagURL("annict"), "https://x.com/search?q=%23annict"; got != want {
		t.Errorf("TwitterHashtagURL(%q) = %q、期待値 = %q", "annict", got, want)
	}
	if got := TwitterHashtagURL(""); got != "" {
		t.Errorf("TwitterHashtagURL(\"\") = %q、期待値 = \"\"", got)
	}
}

func TestExternalIDURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "有効な整数文字列はURLを生成する", value: "3524", want: "http://cal.syoboi.jp/tid/3524"},
		{name: "空文字列は空を返す", value: "", want: ""},
		{name: "非数値は空を返す", value: "abc", want: ""},
		{name: "int32を超える値は空を返す", value: "9999999999", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := externalIDURL(tt.value, SyobocalURL); got != tt.want {
				t.Errorf("externalIDURL(%q) = %q、期待値 = %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestNewExternalServiceLink(t *testing.T) {
	t.Parallel()

	t.Run("idがあればラベルとURLを生成する", func(t *testing.T) {
		t.Parallel()

		id := int32(3524)
		got := newExternalServiceLink(&id, SyobocalURL)
		if got.Label != "3524" {
			t.Errorf("Label = %q、期待値 = %q", got.Label, "3524")
		}
		if got.URL != "http://cal.syoboi.jp/tid/3524" {
			t.Errorf("URL = %q、期待値 = %q", got.URL, "http://cal.syoboi.jp/tid/3524")
		}
	})

	t.Run("idがnilならゼロ値を返す", func(t *testing.T) {
		t.Parallel()

		got := newExternalServiceLink(nil, SyobocalURL)
		if got.Label != "" || got.URL != "" {
			t.Errorf("newExternalServiceLink(nil) = %+v、期待値 = labelとurlが空", got)
		}
	})
}
