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

// TestIsCurrentPathPrefix verifies IsCurrentPathPrefix treats the link path and the pages
// below it as the current page, while a different page that merely shares a textual prefix
// does not match.
//
// [Ja] TestIsCurrentPathPrefix は IsCurrentPathPrefix がリンク先とその配下のページを現在
// ページと判定し、文字列として接頭辞が共通なだけの別ページには一致しないことを検証する。
func TestIsCurrentPathPrefix(t *testing.T) {
	tests := []struct {
		name        string
		currentPath string
		linkPath    string
		want        bool
	}{
		{
			name:        "完全一致",
			currentPath: "/db/works",
			linkPath:    "/db/works",
			want:        true,
		},
		{
			name:        "子パス",
			currentPath: "/db/works/1/edit",
			linkPath:    "/db/works",
			want:        true,
		},
		{
			name:        "新規作成の子パス",
			currentPath: "/db/works/new",
			linkPath:    "/db/works",
			want:        true,
		},
		{
			name:        "セグメント境界をまたぐ別ページは一致しない",
			currentPath: "/db/series_works/1",
			linkPath:    "/db/series",
			want:        false,
		},
		{
			name:        "接頭辞が共通なだけの別ページは一致しない",
			currentPath: "/db/channel_groups",
			linkPath:    "/db/channels",
			want:        false,
		},
		{
			name:        "末尾スラッシュの違いを無視する",
			currentPath: "/db/works/",
			linkPath:    "/db/works",
			want:        true,
		},
		{
			name:        "クエリ文字列を無視する",
			currentPath: "/db/works/1/edit?tab=basic",
			linkPath:    "/db/works",
			want:        true,
		},
		{
			name:        "親パスは子のリンクに一致しない",
			currentPath: "/db/works",
			linkPath:    "/db/works/1/edit",
			want:        false,
		},
		{
			name:        "別ページは一致しない",
			currentPath: "/db/people",
			linkPath:    "/db/works",
			want:        false,
		},
		{
			// Prefix-matching the root would mark every page, so "/" only matches itself.
			//
			// [Ja] ルートを前方一致で扱うと全ページに印が付いてしまうため、"/" は自分自身にだけ
			// 一致する。
			name:        "ルートは配下のページに一致しない",
			currentPath: "/db/works",
			linkPath:    "/",
			want:        false,
		},
		{
			name:        "外部リンクは内部パスと一致しない",
			currentPath: "/db/works",
			linkPath:    "https://developers.annict.com/",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := SetCurrentPath(context.Background(), tt.currentPath)
			if got := IsCurrentPathPrefix(ctx, tt.linkPath); got != tt.want {
				t.Errorf("IsCurrentPathPrefix(%q, %q) = %v, want %v", tt.currentPath, tt.linkPath, got, tt.want)
			}
		})
	}
}

// TestIsCurrentPathPrefix_NoPathSet verifies no link matches when no path has been set.
//
// [Ja] TestIsCurrentPathPrefix_NoPathSet はパス未設定のときどのリンクにも一致しないことを
// 検証する。
func TestIsCurrentPathPrefix_NoPathSet(t *testing.T) {
	ctx := context.Background()
	if IsCurrentPathPrefix(ctx, "/db/works") {
		t.Error("パス未設定のとき /db/works は一致しないはず")
	}
}
