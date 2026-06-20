package viewmodel

import (
	"testing"
)

func TestNewPagination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		currentPage     int
		totalCount      int
		perPage         int
		basePath        string
		wantCurrentPage int
		wantTotalPages  int
	}{
		{
			name:            "通常のページネーション",
			currentPage:     2,
			totalCount:      50,
			perPage:         10,
			basePath:        "/db/works",
			wantCurrentPage: 2,
			wantTotalPages:  5,
		},
		{
			name:            "端数がある場合は切り上げ",
			currentPage:     1,
			totalCount:      51,
			perPage:         10,
			basePath:        "/db/works",
			wantCurrentPage: 1,
			wantTotalPages:  6,
		},
		{
			name:            "ページ番号が0以下の場合は1に補正",
			currentPage:     0,
			totalCount:      50,
			perPage:         10,
			basePath:        "/db/works",
			wantCurrentPage: 1,
			wantTotalPages:  5,
		},
		{
			name:            "ページ番号が最大を超える場合は最大に補正",
			currentPage:     10,
			totalCount:      50,
			perPage:         10,
			basePath:        "/db/works",
			wantCurrentPage: 5,
			wantTotalPages:  5,
		},
		{
			name:            "データがない場合",
			currentPage:     1,
			totalCount:      0,
			perPage:         10,
			basePath:        "/db/works",
			wantCurrentPage: 1,
			wantTotalPages:  0,
		},
		{
			name:            "perPageが0の場合",
			currentPage:     1,
			totalCount:      50,
			perPage:         0,
			basePath:        "/db/works",
			wantCurrentPage: 1,
			wantTotalPages:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPagination(tt.currentPage, tt.totalCount, tt.perPage, tt.basePath)

			if p.CurrentPage != tt.wantCurrentPage {
				t.Errorf("CurrentPage = %d, want %d", p.CurrentPage, tt.wantCurrentPage)
			}
			if p.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", p.TotalPages, tt.wantTotalPages)
			}
		})
	}
}

func TestPagination_ShouldShow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		totalCount int
		perPage    int
		want       bool
	}{
		{"1ページに収まる場合は非表示", 10, 10, false},
		{"2ページ以上の場合は表示", 11, 10, true},
		{"データがない場合は非表示", 0, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPagination(1, tt.totalCount, tt.perPage, "/test")
			if got := p.ShouldShow(); got != tt.want {
				t.Errorf("ShouldShow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPagination_HasPrevAndHasNext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		currentPage int
		totalCount  int
		perPage     int
		wantPrev    bool
		wantNext    bool
	}{
		{"最初のページ", 1, 50, 10, false, true},
		{"中間のページ", 3, 50, 10, true, true},
		{"最後のページ", 5, 50, 10, true, false},
		{"1ページのみ", 1, 5, 10, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPagination(tt.currentPage, tt.totalCount, tt.perPage, "/test")
			if got := p.HasPrev(); got != tt.wantPrev {
				t.Errorf("HasPrev() = %v, want %v", got, tt.wantPrev)
			}
			if got := p.HasNext(); got != tt.wantNext {
				t.Errorf("HasNext() = %v, want %v", got, tt.wantNext)
			}
		})
	}
}

func TestPagination_PageURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		basePath string
		page     int
		want     string
	}{
		{"1ページ目はpageパラメータなし", "/db/works", 1, "/db/works"},
		{"2ページ目以降はpageパラメータあり", "/db/works", 3, "/db/works?page=3"},
		{"既存クエリパラメータを保持", "/db/works?q=test", 2, "/db/works?page=2&q=test"},
		{"既存クエリパラメータを保持（1ページ目）", "/db/works?q=test", 1, "/db/works?q=test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPagination(1, 100, 10, tt.basePath)
			if got := p.PageURL(tt.page); got != tt.want {
				t.Errorf("PageURL(%d) = %q, want %q", tt.page, got, tt.want)
			}
		})
	}
}

func TestPagination_Pages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		currentPage int
		totalPages  int
		want        []int
	}{
		{
			name:        "7ページ以下の場合はすべて表示",
			currentPage: 1,
			totalPages:  5,
			want:        []int{1, 2, 3, 4, 5},
		},
		{
			name:        "先頭付近のページ",
			currentPage: 2,
			totalPages:  10,
			want:        []int{1, 2, 3, 4, 0, 10},
		},
		{
			name:        "末尾付近のページ",
			currentPage: 9,
			totalPages:  10,
			want:        []int{1, 0, 7, 8, 9, 10},
		},
		{
			name:        "中間のページ",
			currentPage: 5,
			totalPages:  10,
			want:        []int{1, 0, 4, 5, 6, 0, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPagination(tt.currentPage, tt.totalPages*10, 10, "/test")
			got := p.Pages()

			if len(got) != len(tt.want) {
				t.Fatalf("Pages() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Pages()[%d] = %d, want %d (full: %v)", i, got[i], tt.want[i], got)
					break
				}
			}
		})
	}
}
