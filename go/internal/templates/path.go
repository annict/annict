package templates

import (
	"fmt"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/viewmodel"
)

// PathはURLのパスを表す型です
type Path string

// Stringはパスを文字列として返します
func (p Path) String() string {
	return string(p)
}

// SafeURLはパスをtempl.SafeURLとして返します
func (p Path) SafeURL() templ.SafeURL {
	return templ.SafeURL(p)
}

// WorkPathは作品の公開ページ (Annict DB管理画面の外) のパスを生成します。
func WorkPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/works/%s", id))
}

// EpisodePathはエピソードの公開ページ (Annict DB管理画面の外) のパスを生成します。
func EpisodePath(workID viewmodel.WorkID, episodeID viewmodel.EpisodeID) Path {
	return Path(fmt.Sprintf("/works/%s/episodes/%s", workID, episodeID))
}

// PersonPathは人物の公開ページ (Annict DB管理画面の外) のパスを生成します。
func PersonPath(id viewmodel.PersonID) Path {
	return Path(fmt.Sprintf("/people/%s", id))
}

// DBPeoplePathはDB管理画面の人物一覧のパスを生成します。
func DBPeoplePath() Path {
	return Path("/db/people")
}

// DBPeopleNewPathはDB管理画面の人物一括登録のパスを生成します。
func DBPeopleNewPath() Path {
	return Path("/db/people/new")
}

// DBPersonPathはDB管理画面の単一人物のパスを生成します。編集フォームのPATCH先です。
func DBPersonPath(id viewmodel.PersonID) Path {
	return Path(fmt.Sprintf("/db/people/%s", id))
}

// DBPersonEditPathはDB管理画面の人物編集のパスを生成します。
func DBPersonEditPath(id viewmodel.PersonID) Path {
	return Path(fmt.Sprintf("/db/people/%s/edit", id))
}

// DBPersonArchivePathはDB管理画面で人物をアーカイブするパスを生成します。アーカイブ確認
// フォームのPOST先、およびアーカイブ解除のDELETE先として使います。
func DBPersonArchivePath(id viewmodel.PersonID) Path {
	return Path(fmt.Sprintf("/db/people/%s/archive", id))
}

// DBPersonArchiveNewPathはDB管理画面の人物のアーカイブ確認画面のパスを生成します。人物一覧の
// アーカイブ操作からリンクします。
func DBPersonArchiveNewPath(id viewmodel.PersonID) Path {
	return Path(fmt.Sprintf("/db/people/%s/archive/new", id))
}

// DBPersonDelistingPathはDB管理画面で人物の掲載を停止するパスを生成します。掲載停止確認
// フォームのPOST先、および掲載停止の解除のDELETE先として使います。
func DBPersonDelistingPath(id viewmodel.PersonID) Path {
	return Path(fmt.Sprintf("/db/people/%s/delisting", id))
}

// DBPersonDelistingNewPathはDB管理画面の人物の掲載停止確認画面のパスを生成します。人物一覧の
// 掲載停止の操作からリンクします。
func DBPersonDelistingNewPath(id viewmodel.PersonID) Path {
	return Path(fmt.Sprintf("/db/people/%s/delisting/new", id))
}

// DBPersonUndelistingNewPathはDB管理画面の人物の掲載停止解除確認画面のパスを生成します。人物一覧の
// 掲載停止の解除の操作からリンクします。
func DBPersonUndelistingNewPath(id viewmodel.PersonID) Path {
	return Path(fmt.Sprintf("/db/people/%s/undelisting/new", id))
}

// DBWorksPathはDB管理画面の作品一覧のパスを生成します
func DBWorksPath() Path {
	return Path("/db/works")
}

// DBWorksNewPathはDB管理画面の作品新規作成のパスを生成します
func DBWorksNewPath() Path {
	return Path("/db/works/new")
}

// DBWorkPathはDB管理画面の作品詳細のパスを生成します
func DBWorkPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s", id))
}

// DBWorkEditPathはDB管理画面の作品編集のパスを生成します。
func DBWorkEditPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/edit", id))
}

// DBWorkArchivePathはDB管理画面で作品を非公開にするパスを生成します。非公開確認
// フォームのPOST先、および再公開のDELETE先として使います。
func DBWorkArchivePath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/archive", id))
}

// DBWorkArchiveNewPathはDB管理画面の非公開確認画面のパスを生成します。作品一覧の
// 非公開操作からリンクします。
func DBWorkArchiveNewPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/archive/new", id))
}

// DBWorkEpisodesPathはDB管理画面の作品のエピソード一覧のパスを生成します。
func DBWorkEpisodesPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/episodes", id))
}

// DBWorkEpisodesNewPathはDB管理画面の作品のエピソード一括作成フォームのパスを生成します。
func DBWorkEpisodesNewPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/episodes/new", id))
}

// DBEpisodePathはDB管理画面の単一エピソードのパスを生成します。編集フォームのPATCH
// 先です。
func DBEpisodePath(id viewmodel.EpisodeID) Path {
	return Path(fmt.Sprintf("/db/episodes/%s", id))
}

// DBEpisodeEditPathはDB管理画面のエピソード編集ページのパスを生成します。エピソード
// 一覧の編集操作からリンクします。
func DBEpisodeEditPath(id viewmodel.EpisodeID) Path {
	return Path(fmt.Sprintf("/db/episodes/%s/edit", id))
}

// DBEpisodeArchivePathはDB管理画面でエピソードを非公開にするパスを生成します。非公開
// 確認フォームのPOST先、および再公開のDELETE先として使います。
func DBEpisodeArchivePath(id viewmodel.EpisodeID) Path {
	return Path(fmt.Sprintf("/db/episodes/%s/archive", id))
}

// DBEpisodeArchiveNewPathはDB管理画面のエピソード非公開確認画面のパスを生成します。
// エピソード一覧の非公開操作からリンクします。
func DBEpisodeArchiveNewPath(id viewmodel.EpisodeID) Path {
	return Path(fmt.Sprintf("/db/episodes/%s/archive/new", id))
}

// DBWorkProgramsPathはDB管理画面の作品の番組情報一覧のパスを生成します。
func DBWorkProgramsPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/programs", id))
}

// DBWorkSlotsPathはDB管理画面の作品の放送予定一覧のパスを生成します。
func DBWorkSlotsPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/slots", id))
}

// DBWorkCastsPathはDB管理画面の作品のキャスト一覧のパスを生成します。
func DBWorkCastsPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/casts", id))
}

// DBWorkStaffsPathはDB管理画面の作品のスタッフ一覧のパスを生成します。
func DBWorkStaffsPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/staffs", id))
}

// DBWorkImagePathはDB管理画面の作品の作品画像ページのパスを生成します。
func DBWorkImagePath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/image", id))
}

// DBWorkTrailersPathはDB管理画面の作品のPV一覧のパスを生成します。
func DBWorkTrailersPath(id viewmodel.WorkID) Path {
	return Path(fmt.Sprintf("/db/works/%s/trailers", id))
}
