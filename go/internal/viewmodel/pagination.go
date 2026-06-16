package viewmodel

import (
	"fmt"
	"net/url"
)

type Pagination struct {
	CurrentPage int
	TotalPages  int
	TotalCount  int
	PerPage     int
	BasePath    string
}

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

func (p Pagination) ShouldShow() bool {
	return p.TotalPages > 1
}

func (p Pagination) HasPrev() bool {
	return p.CurrentPage > 1
}

func (p Pagination) HasNext() bool {
	return p.CurrentPage < p.TotalPages
}

func (p Pagination) PrevPage() int {
	if p.HasPrev() {
		return p.CurrentPage - 1
	}
	return p.CurrentPage
}

func (p Pagination) NextPage() int {
	if p.HasNext() {
		return p.CurrentPage + 1
	}
	return p.CurrentPage
}

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

// Pages returns the slice of page numbers to render. A value of 0 marks an ellipsis (…).
//
// [Ja] Pages はレンダリングするページ番号のスライスを返す。値 0 は省略記号 (…) を表す。
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
