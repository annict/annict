package viewmodel

import (
	"strconv"
	"strings"

	"github.com/annict/annict/go/internal/model"
)

// DBEpisodeListWorkName returns the parent work's name as the episode list shows it, trimmed
// of surrounding whitespace. An empty result means the work has no name to show: both the page
// heading and the document title read it as the signal to fall back to the generic page name,
// so the two never disagree about which works have one.
//
// [Ja] DBEpisodeListWorkName はエピソード一覧が表示する親作品の名前を、前後の空白を落として
// 返す。空文字列は表示できる名前が無いことを表し、ページ見出しと文書タイトルはいずれもそれを
// 汎用のページ名へフォールバックする合図として読む。これにより、どの作品に名前があるかの判断
// が両者でずれない。
func DBEpisodeListWorkName(workTitle string) string {
	return strings.TrimSpace(workTitle)
}

// DBEpisodeListItem is the per-row display data for a work's episode list on the Annict DB
// admin screen.
//
// [Ja] DBEpisodeListItem は Annict DB 管理画面の、ある作品のエピソード一覧で 1 行ごとに
// 表示する整形済みデータ。
type DBEpisodeListItem struct {
	ID EpisodeID
	// WorkID is the parent work, which the row's id link needs to build the episode's
	// public URL.
	//
	// [Ja] WorkID は親作品。行の ID リンクがエピソードの公開 URL を組み立てるのに使う。
	WorkID WorkID
	// Number is the display number (episodes.number, e.g. "第2話") and RawNumber the
	// numeric one (episodes.raw_number) formatted for display. Both are empty when unset;
	// the template renders a "-" placeholder in that case.
	//
	// [Ja] Number は表示用の話数 (episodes.number、例: "第2話")、RawNumber は数値の話数
	// (episodes.raw_number) を表示用に整形したもの。未設定ならいずれも空文字列で、
	// テンプレート側で "-" のプレースホルダーを表示する。
	Number    string
	RawNumber string
	// Title and TitleEn are the Japanese and English titles, empty when unset.
	//
	// [Ja] Title と TitleEn は日本語・英語のタイトル。未設定なら空文字列。
	Title      string
	TitleEn    string
	SortNumber int32
	Status     PublishingStatus
}

// NewDBEpisodeListItems converts the episodes of one list page into their display rows.
//
// [Ja] NewDBEpisodeListItems は一覧 1 ページ分のエピソードを表示用の行に変換する。
func NewDBEpisodeListItems(episodes []*model.Episode) []DBEpisodeListItem {
	result := make([]DBEpisodeListItem, len(episodes))
	for i, episode := range episodes {
		result[i] = NewDBEpisodeListItem(episode)
	}
	return result
}

// NewDBEpisodeListItem converts one episode into its display row.
//
// [Ja] NewDBEpisodeListItem は 1 件のエピソードを表示用の行に変換する。
func NewDBEpisodeListItem(episode *model.Episode) DBEpisodeListItem {
	return DBEpisodeListItem{
		ID:         EpisodeID(episode.ID),
		WorkID:     WorkID(episode.WorkID),
		Number:     derefString(episode.Number),
		RawNumber:  formatRawNumber(episode.RawNumber),
		Title:      derefString(episode.Title),
		TitleEn:    episode.TitleEn,
		SortNumber: episode.SortNumber,
		Status:     PublishingStatus(episode.DerivedStatus()),
	}
}

// formatRawNumber renders an episode's numeric number without trailing zeros, so a whole
// number reads as "2" while a half episode keeps its fraction ("2.5"). It returns "" when
// unset so the template can decide how to render the gap.
//
// [Ja] formatRawNumber はエピソードの数値話数を末尾の 0 を付けずに描画する。整数の話数は
// "2"、0.5 話は小数のまま "2.5" と読める。未設定では "" を返し、テンプレート側で欠落の
// 描画方法を決められるようにする。
func formatRawNumber(rawNumber *float64) string {
	if rawNumber == nil {
		return ""
	}

	return strconv.FormatFloat(*rawNumber, 'f', -1, 64)
}
