package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/viewmodel"
)

// TestIndex_Empty は作品が存在しない場合に表が表示されず空メッセージが表示されることをテスト
func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := IndexPageData{
		Works:      []viewmodel.DBWorkListItem{},
		Pagination: viewmodel.NewPagination(1, 0, 30, "/db/works"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// テーブルが表示されないことを確認
	if strings.Contains(html, "<table") {
		t.Error("作品が空の場合は <table> が含まれてはいけません")
	}

	// 空メッセージが表示されることを確認
	if !strings.Contains(html, "作品が見つかりませんでした") {
		t.Error("作品が空の場合は空メッセージが表示されるべきです")
	}
}

// TestIndex_WithWorks は作品が存在する場合に表が表示されることをテスト
func TestIndex_WithWorks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := IndexPageData{
		Works: []viewmodel.DBWorkListItem{
			{
				ID:            1,
				Title:         "テストアニメ1",
				Season:        "2024 春",
				WatchersCount: 100,
				Status:        "currently",
				HasImage:      true,
			},
			{
				ID:            2,
				Title:         "テストアニメ2",
				Season:        "2024 夏",
				WatchersCount: 50,
				Status:        "currently",
				HasImage:      false,
			},
		},
		Pagination: viewmodel.NewPagination(1, 2, 30, "/db/works"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		"<table",
		"<thead",
		"<tbody",
		"テストアニメ1",
		"テストアニメ2",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}
