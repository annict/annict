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

// WorkPath builds the path for a work's public page (outside the Annict DB admin UI).
//
// [Ja] WorkPath は作品の公開ページ (Annict DB 管理画面の外) のパスを生成します。
func WorkPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/works/%s", id))
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

// DBWorkEditPath builds the path for the work edit page in the Annict DB admin UI.
//
// [Ja] DBWorkEditPath はDB管理画面の作品編集のパスを生成します。
func DBWorkEditPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/edit", id))
}

// DBWorkEpisodesPath builds the path for a work's episode list in the Annict DB admin UI.
//
// [Ja] DBWorkEpisodesPath はDB管理画面の作品のエピソード一覧のパスを生成します。
func DBWorkEpisodesPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/episodes", id))
}

// DBWorkProgramsPath builds the path for a work's program list in the Annict DB admin UI.
//
// [Ja] DBWorkProgramsPath はDB管理画面の作品の番組情報一覧のパスを生成します。
func DBWorkProgramsPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/programs", id))
}

// DBWorkSlotsPath builds the path for a work's broadcast slot list in the Annict DB admin UI.
//
// [Ja] DBWorkSlotsPath はDB管理画面の作品の放送予定一覧のパスを生成します。
func DBWorkSlotsPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/slots", id))
}

// DBWorkCastsPath builds the path for a work's cast list in the Annict DB admin UI.
//
// [Ja] DBWorkCastsPath はDB管理画面の作品のキャスト一覧のパスを生成します。
func DBWorkCastsPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/casts", id))
}

// DBWorkStaffsPath builds the path for a work's staff list in the Annict DB admin UI.
//
// [Ja] DBWorkStaffsPath はDB管理画面の作品のスタッフ一覧のパスを生成します。
func DBWorkStaffsPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/staffs", id))
}

// DBWorkImagePath builds the path for a work's image page in the Annict DB admin UI.
//
// [Ja] DBWorkImagePath はDB管理画面の作品の作品画像ページのパスを生成します。
func DBWorkImagePath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/image", id))
}

// DBWorkTrailersPath builds the path for a work's PV (trailer) list in the Annict DB admin UI.
//
// [Ja] DBWorkTrailersPath はDB管理画面の作品のPV一覧のパスを生成します。
func DBWorkTrailersPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/trailers", id))
}
