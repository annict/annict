package templates

import (
	"fmt"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/viewmodel"
)

// Path はURLのパスを表す型です
type Path string

// String はパスを文字列として返します
func (p Path) String() string {
	return string(p)
}

// SafeURL はパスをtempl.SafeURLとして返します
func (p Path) SafeURL() templ.SafeURL {
	return templ.SafeURL(p)
}

// DBWorksPath はDB管理画面の作品一覧のパスを生成します
func DBWorksPath() Path {
	return Path("/db/works")
}

// DBWorksNewPath はDB管理画面の作品新規作成のパスを生成します
func DBWorksNewPath() Path {
	return Path("/db/works/new")
}

// DBWorkPath はDB管理画面の作品詳細のパスを生成します
func DBWorkPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s", id))
}
