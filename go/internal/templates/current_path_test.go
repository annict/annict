package templates

import (
	"context"
	"testing"
)

// ========================================
// 現在パスヘルパーのテスト
// ========================================

// TestIsCurrentPath は IsCurrentPath が現在ページのリンクを正しく判定することを確認する
func TestIsCurrentPath(t *testing.T) {
	tests := []struct {
		name        string
		currentPath string
		linkPath    string
		want        bool
	}{
		{
			name:        "完全一致",
			currentPath: "/track",
			linkPath:    "/track",
			want:        true,
		},
		{
			name:        "ルートパス",
			currentPath: "/",
			linkPath:    "/",
			want:        true,
		},
		{
			name:        "末尾スラッシュの違いを無視する",
			currentPath: "/track/",
			linkPath:    "/track",
			want:        true,
		},
		{
			name:        "クエリ文字列を無視する",
			currentPath: "/works/popular?sort=asc",
			linkPath:    "/works/popular",
			want:        true,
		},
		{
			name:        "別ページは一致しない",
			currentPath: "/works/popular",
			linkPath:    "/works/newest",
			want:        false,
		},
		{
			name:        "前方一致では一致させない",
			currentPath: "/track/123",
			linkPath:    "/track",
			want:        false,
		},
		{
			name:        "外部リンクは内部パスと一致しない",
			currentPath: "/faq",
			linkPath:    "https://developers.annict.com/",
			want:        false,
		},
		{
			name:        "ルートと他ページは一致しない",
			currentPath: "/notifications",
			linkPath:    "/",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := SetCurrentPath(context.Background(), tt.currentPath)
			if got := IsCurrentPath(ctx, tt.linkPath); got != tt.want {
				t.Errorf("IsCurrentPath(%q, %q) = %v, want %v", tt.currentPath, tt.linkPath, got, tt.want)
			}
		})
	}
}

// TestIsCurrentPath_NoPathSet はパス未設定のとき (ルート以外は) 一致しないことを確認する
func TestIsCurrentPath_NoPathSet(t *testing.T) {
	ctx := context.Background()
	if IsCurrentPath(ctx, "/track") {
		t.Error("パス未設定のとき /track は一致しないはず")
	}
}
