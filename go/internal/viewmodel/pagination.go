package viewmodel

import (
	"fmt"
	"net/url"
)

// Pagination はページネーション情報を保持します
type Pagination struct {
	CurrentPage int
	TotalPages  int
	TotalCount  int
	PerPage     int
	BasePath    string
}

// NewPagination はPaginationを作成します
func NewPagination(currentPage, totalCount, perPage int, basePath string) Pagination {
	totalPages := 0
	if perPage > 0 {
		totalPages = (totalCount + perPage - 1) / perPage
	}
	if currentPage < 1 {
		currentPage = 1
	}
	if totalPages > 0 && currentPage > totalPages {
		currentPage = totalPages
	}
	return Pagination{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		PerPage:     perPage,
		BasePath:    basePath,
	}
}

// ShouldShow はページネーションを表示すべきかどうかを返します
func (p Pagination) ShouldShow() bool {
	return p.TotalPages > 1
}

// HasPrev は前のページが存在するかどうかを返します
func (p Pagination) HasPrev() bool {
	return p.CurrentPage > 1
}

// HasNext は次のページが存在するかどうかを返します
func (p Pagination) HasNext() bool {
	return p.CurrentPage < p.TotalPages
}

// PrevPage は前のページ番号を返します
func (p Pagination) PrevPage() int {
	if p.HasPrev() {
		return p.CurrentPage - 1
	}
	return p.CurrentPage
}

// NextPage は次のページ番号を返します
func (p Pagination) NextPage() int {
	if p.HasNext() {
		return p.CurrentPage + 1
	}
	return p.CurrentPage
}

// PageURL は指定されたページ番号のURLを返します
func (p Pagination) PageURL(page int) string {
	u, err := url.Parse(p.BasePath)
	if err != nil {
		return p.BasePath
	}
	q := u.Query()
	if page <= 1 {
		q.Del("page")
	} else {
		q.Set("page", fmt.Sprintf("%d", page))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// Pages はページ番号のスライスを返します
// 0は省略記号（…）を表します
func (p Pagination) Pages() []int {
	if p.TotalPages <= 7 {
		pages := make([]int, p.TotalPages)
		for i := range pages {
			pages[i] = i + 1
		}
		return pages
	}

	pages := []int{1}

	if p.CurrentPage <= 3 {
		pages = append(pages, 2, 3, 4, 0)
	} else if p.CurrentPage >= p.TotalPages-2 {
		pages = append(pages, 0, p.TotalPages-3, p.TotalPages-2, p.TotalPages-1)
	} else {
		pages = append(pages, 0, p.CurrentPage-1, p.CurrentPage, p.CurrentPage+1, 0)
	}

	pages = append(pages, p.TotalPages)
	return pages
}
