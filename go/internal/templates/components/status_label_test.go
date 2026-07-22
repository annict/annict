package components

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

func TestStatusLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   viewmodel.WorkStatus
		locale   string
		wantText string
		wantAttr string
	}{
		{
			name:     "公開状態（日本語）",
			status:   viewmodel.WorkStatusPublished,
			locale:   "ja",
			wantText: "公開",
			wantAttr: `class="badge" data-variant="success"`,
		},
		{
			name:     "非公開状態（日本語）",
			status:   viewmodel.WorkStatusArchived,
			locale:   "ja",
			wantText: "非公開",
			wantAttr: `class="badge" data-variant="warning"`,
		},
		{
			name:     "削除状態（日本語）",
			status:   viewmodel.WorkStatusDeleted,
			locale:   "ja",
			wantText: "削除",
			wantAttr: `class="badge" data-variant="destructive"`,
		},
		{
			name:     "公開状態（英語）",
			status:   viewmodel.WorkStatusPublished,
			locale:   "en",
			wantText: "Published",
			wantAttr: `class="badge" data-variant="success"`,
		},
		{
			name:     "アーカイブ状態（英語）",
			status:   viewmodel.WorkStatusArchived,
			locale:   "en",
			wantText: "Archived",
			wantAttr: `class="badge" data-variant="warning"`,
		},
		{
			name:     "削除状態（英語）",
			status:   viewmodel.WorkStatusDeleted,
			locale:   "en",
			wantText: "Deleted",
			wantAttr: `class="badge" data-variant="destructive"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ctx = i18n.SetLocale(ctx, tt.locale)

			var buf bytes.Buffer
			err := StatusLabel(tt.status).Render(ctx, &buf)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			html := buf.String()
			if !strings.Contains(html, tt.wantText) {
				t.Errorf("出力に %q が含まれていません: %s", tt.wantText, html)
			}
			if !strings.Contains(html, tt.wantAttr) {
				t.Errorf("出力に %q が含まれていません: %s", tt.wantAttr, html)
			}
		})
	}
}

func TestStatusLabel_UnknownStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = i18n.SetLocale(ctx, "ja")

	var buf bytes.Buffer
	err := StatusLabel(viewmodel.WorkStatus("unknown")).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := buf.String()
	if html != "" {
		t.Errorf("不明なステータスの場合は空出力を期待, got: %s", html)
	}
}
