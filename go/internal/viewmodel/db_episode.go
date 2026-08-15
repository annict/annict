package viewmodel

import (
	"context"
	"strconv"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

// DBEpisodeListWorkName returns the parent work's display name for the Annict DB episode
// pages (the list, the bulk-create form and the edit form), trimmed of surrounding whitespace.
// An empty result means the work has no name to show: the page heading falls back to the
// generic page name and the document title omits the work, so the two never disagree about
// which works have one.
//
// [Ja] DBEpisodeListWorkName は Annict DB のエピソード関連ページ (一覧・一括作成フォーム・
// 編集フォーム) が表示する親作品の名前を、前後の空白を落として返す。空文字列は表示できる
// 名前が無いことを表し、ページ見出しは汎用のページ名へフォールバックし、文書タイトルは作品を
// 省く。これにより、どの作品に名前があるかの判断が両者でずれない。
func DBEpisodeListWorkName(workTitle string) string {
	return strings.TrimSpace(workTitle)
}

// DBEpisodeIdentifier returns the label used at the start of a document title that names one
// episode. The immutable id always remains in the label so two episodes of one work have
// different titles; a display number, when present, makes the label recognisable before the id.
//
// [Ja] DBEpisodeIdentifier は 1 件のエピソードを名指しする文書タイトルの先頭で使うラベルを返す。
// 同じ作品の 2 エピソードが異なるタイトルになるよう不変の ID を必ず残し、表示用話数があれば
// ID より先に置いて識別しやすくする。
func DBEpisodeIdentifier(ctx context.Context, episode *model.Episode) string {
	templateData := map[string]any{"EpisodeID": episode.ID.String()}
	if number := strings.TrimSpace(derefString(episode.Number)); number != "" {
		templateData["Number"] = number
		return i18n.T(ctx, "db_episodes_identifier", templateData)
	}

	return i18n.T(ctx, "db_episodes_identifier_without_number", templateData)
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
	Title   string
	TitleEn string
	// PrevNumber names the episode that comes just before this one in sort_number order,
	// formatted for display and empty when there is none. It is empty for the work's first
	// episode and for a preceding episode that carries neither number.
	//
	// [Ja] PrevNumber は sort_number 順でこのエピソードの直前に来るエピソードを表示用に
	// 整形して名指しする。該当が無ければ空文字列。作品の最初のエピソードと、直前の
	// エピソードがどちらの話数も持たない場合に空になる。
	PrevNumber          string
	SortNumber          int32
	EpisodeRecordsCount int32
	Status              PublishingStatus
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
		ID:                  EpisodeID(episode.ID),
		WorkID:              WorkID(episode.WorkID),
		Number:              derefString(episode.Number),
		RawNumber:           formatRawNumber(episode.RawNumber),
		Title:               derefString(episode.Title),
		TitleEn:             episode.TitleEn,
		PrevNumber:          formatPrevNumber(episode),
		SortNumber:          episode.SortNumber,
		EpisodeRecordsCount: episode.EpisodeRecordsCount,
		Status:              PublishingStatus(episode.DerivedStatus()),
	}
}

// formatPrevNumber renders the preceding episode as its display number, falling back to its
// numeric number when it has no display one. The fallback keeps the column from reading as
// "no preceding episode" for an episode that has one but was only ever given a raw number.
//
// [Ja] formatPrevNumber は直前のエピソードを表示用話数で描画し、表示用話数を持たない場合は
// 数値話数にフォールバックする。このフォールバックにより、直前のエピソードが存在するのに
// 数値話数しか付けられていない場合でも、列が「直前のエピソードなし」と読めてしまうのを防ぐ。
func formatPrevNumber(episode *model.Episode) string {
	if number := derefString(episode.PrevNumber); number != "" {
		return number
	}

	return formatRawNumber(episode.PrevRawNumber)
}

// DBEpisodeFormInput holds the values the episode edit form renders in its fields. They are
// strings because that is what the form round-trips: an unset number and a number the editor
// cleared are the same empty input, and a rejected submit is re-rendered from what was typed
// rather than from what could be parsed.
//
// The fields are read directly by the template instead of through a lookup by field name.
// The form has five editable ones, all known at compile time, so naming them keeps the
// template's references type-checked. UpdatedAt rides along as the hidden version rather than
// as an editable value.
//
// [Ja] DBEpisodeFormInput はエピソード編集フォームが各欄に描画する値を保持する。値を文字列で
// 持つのはフォームが往復させるものが文字列であるため。未設定の話数と編集者が消した話数は
// どちらも同じ空の入力であり、却下された送信はパースできた値ではなく入力された内容から
// 再描画する。
//
// テンプレートはフィールド名による引き当てではなく各フィールドを直接読む。編集できる欄は
// 5 つですべてコンパイル時に分かっているため、名前で持つことでテンプレートの参照が型検査を
// 受けられる。UpdatedAt は編集できる値ではなく hidden の版として同居する。
type DBEpisodeFormInput struct {
	Number     string
	RawNumber  string
	SortNumber string
	Title      string
	TitleEn    string
	// UpdatedAt is the version the form was opened against, carried in a hidden field so
	// the update can reject a submit made against a stale read instead of silently
	// overwriting whoever wrote in between. It travels with the typed values rather than
	// beside them because a rejected submit has to echo back the version the editor
	// submitted, not the one the server holds now: re-reading it would make the next
	// submit overwrite the very change that is being guarded against.
	//
	// It is DBEpisodeFormNullVersion for an episode whose updated_at is unset. That explicit
	// value is distinct from an empty request, which means no version was stated.
	//
	// [Ja] UpdatedAt はフォームを開いた時点の版で、hidden で持ち回る。古い読み取りに対する
	// 送信を、間に書いた人の変更を黙って上書きせずに更新側で却下できるようにするため。
	// 入力された値と一緒に持つのは、却下された送信では編集者が送った版をそのまま返す必要が
	// あるからで、サーバーの現在値を読み直すと、次の送信が守るべき変更をかえって上書きして
	// しまう。
	//
	// updated_at を持たないエピソードでは DBEpisodeFormNullVersion になる。この明示値は、
	// 版を指定していない空の要求とは区別する。
	UpdatedAt string
}

// DBEpisodeFormNullVersion is the explicit version the edit form carries when the stored
// updated_at is NULL. The update side matches it with updated_at IS NULL; the first successful
// write advances the column to a timestamp, so another submit from the same NULL version
// conflicts. An empty value remains reserved for a request that stated no version.
//
// Both this and the layout below are the validator's constants: it reads the submitted version
// back and decides whether to accept it, so the round-trip format is single-sourced there and
// named here for the form that writes it.
//
// [Ja] DBEpisodeFormNullVersion は保存済み updated_at が NULL のとき、編集フォームが運ぶ
// 明示的な版。更新側は updated_at IS NULL で照合し、最初に成功した書き込みがカラムを
// timestamp へ進めるため、同じ NULL 版からの次の送信は競合する。空文字列は版を指定して
// いない要求のために残す。
//
// 本定数と下の書式はいずれも validator 側の定数である。送信された版を読み戻して受理するかを
// 判断するのは validator のため、往復の書式の正本をそちらに 1 つ置き、ここでは書き出す
// フォームのために名前を与える。
const DBEpisodeFormNullVersion = validator.DBEpisodeNullVersion

const dbEpisodeFormVersionLayout = validator.DBEpisodeVersionLayout

// NewDBEpisodeFormInputFromEpisode projects a stored episode onto the form values the edit
// form renders. Columns the episode leaves unset become "", so an episode with no display
// number opens with an empty field rather than with a placeholder the editor would have to
// clear.
//
// [Ja] NewDBEpisodeFormInputFromEpisode は保存済みのエピソードを、編集フォームが描画する
// フォーム値に射影する。未設定のカラムは "" になるため、表示用話数を持たないエピソードは
// 編集者が消す必要のあるプレースホルダーではなく空の欄で開く。
func NewDBEpisodeFormInputFromEpisode(episode *model.Episode) DBEpisodeFormInput {
	input := DBEpisodeFormInput{
		Number:     derefString(episode.Number),
		RawNumber:  formatRawNumber(episode.RawNumber),
		SortNumber: strconv.FormatInt(int64(episode.SortNumber), 10),
		Title:      derefString(episode.Title),
		TitleEn:    episode.TitleEn,
		UpdatedAt:  DBEpisodeFormNullVersion,
	}
	if episode.UpdatedAt != nil {
		input.UpdatedAt = episode.UpdatedAt.UTC().Format(dbEpisodeFormVersionLayout)
	}

	return input
}

// NewDBEpisodeFormInputFromSubmit keeps a rejected submit's values in the re-rendered form,
// exactly as they were typed. The version rides back unchanged too: echoing the server's
// current one instead would make the corrected submit overwrite the write the rejection was
// guarding against.
//
// [Ja] NewDBEpisodeFormInputFromSubmit は却下された送信の値を、入力されたまま再描画する
// フォームに残す。版もそのまま返す。代わりにサーバーの現在値を返すと、手直し後の送信が、却下に
// よって守られたはずの書き込みを上書きしてしまう。
func NewDBEpisodeFormInputFromSubmit(input usecase.UpdateEpisodeInput) DBEpisodeFormInput {
	return DBEpisodeFormInput{
		Number:     input.Number,
		RawNumber:  input.RawNumber,
		SortNumber: input.SortNumber,
		Title:      input.Title,
		TitleEn:    input.TitleEn,
		UpdatedAt:  input.UpdatedAt,
	}
}

// DBEpisodeGenerationSummary is the notice a work's episode list shows above the table: how
// many episodes the work is expected to have in total, how many are published now, and how
// far the Syobocal auto-generation could number them from the work's slots.
//
// [Ja] DBEpisodeGenerationSummary は作品のエピソード一覧がテーブルの上に出す案内。作品が
// 最終的に何話になる予定か、現在何話が公開されているか、しょぼいカレンダー由来の自動生成が
// 作品のスロットからどこまで話数を振れるかを表す。
type DBEpisodeGenerationSummary struct {
	// PlannedCount is the work's expected total episode count formatted for display, empty
	// when the work records none. The template renders the "unknown" wording for that gap,
	// as the Rails notice does.
	//
	// [Ja] PlannedCount は作品の予定総話数を表示用に整形したもので、作品が記録していなければ
	// 空文字列。テンプレートは Rails の案内と同じく、その欠落に「不明」の文言を描画する。
	PlannedCount                string
	PublishedEpisodeCount       int64
	MaxGeneratableEpisodeNumber int64
}

// NewDBEpisodeGenerationSummary builds the notice from the work's expected episode count
// (nil when the work records none), its published episode count, and the highest episode
// number its kept slots let auto-generation reach.
//
// [Ja] NewDBEpisodeGenerationSummary は、作品の予定総話数 (記録が無ければ nil)、公開中の
// エピソード数、有効なスロットから自動生成が到達できる最大話数から案内を組み立てる。
func NewDBEpisodeGenerationSummary(
	plannedCount *int32,
	publishedEpisodeCount int64,
	maxGeneratableEpisodeNumber int64,
) DBEpisodeGenerationSummary {
	summary := DBEpisodeGenerationSummary{
		PublishedEpisodeCount:       publishedEpisodeCount,
		MaxGeneratableEpisodeNumber: maxGeneratableEpisodeNumber,
	}
	if plannedCount != nil {
		summary.PlannedCount = strconv.FormatInt(int64(*plannedCount), 10)
	}

	return summary
}

// DBEpisodeManualCreationRestriction is the Presentation-layer projection of the reason a
// work's episodes may not be created by hand. The bulk-create page renders one warning per
// reason and disables its form from the same value, so the page reads a single value rather
// than re-deriving which of several conditions wins.
//
// Like PublishingStatus the constants are written as literals, keeping the templates free of
// a direct model dependency; NewDBEpisodeManualCreationRestriction pins the projection to the
// domain enum, and a test compares the two.
//
// [Ja] DBEpisodeManualCreationRestriction は、作品のエピソードを手動作成できない理由の
// Presentation 層への射影。一括作成ページは理由ごとの警告を描画し、フォームの無効化も同じ値
// から決めるため、複数の条件のどれが優先されるかをページ側で導出せず 1 つの値を読む。
//
// PublishingStatus と同じく定数はリテラルで書き、テンプレートが model へ直接依存しないように
// する。ドメイン enum との対応は NewDBEpisodeManualCreationRestriction が固定し、両者を比較
// するテストで担保する。
type DBEpisodeManualCreationRestriction string

const (
	DBEpisodeManualCreationAllowed        DBEpisodeManualCreationRestriction = ""
	DBEpisodeManualCreationEpisodesFilled DBEpisodeManualCreationRestriction = "episodes_filled"
	DBEpisodeManualCreationSlotsExist     DBEpisodeManualCreationRestriction = "slots_exist"
)

// NewDBEpisodeManualCreationRestriction projects the work's manual-creation state onto the
// reason the page states.
//
// [Ja] NewDBEpisodeManualCreationRestriction は作品の手動作成状態を、ページが述べる理由へ
// 射影する。
func NewDBEpisodeManualCreationRestriction(state model.ManualEpisodeCreationState) DBEpisodeManualCreationRestriction {
	return DBEpisodeManualCreationRestriction(state.Restriction())
}

// Restricted reports whether the page has a reason to warn about.
//
// [Ja] Restricted はページが警告すべき理由があるかを返す。
func (r DBEpisodeManualCreationRestriction) Restricted() bool {
	return r != DBEpisodeManualCreationAllowed
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
