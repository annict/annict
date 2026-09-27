package viewmodel

import (
	"context"
	"strconv"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// DBEpisodeListWorkNameはAnnict DBのエピソード関連ページ (一覧・一括作成フォーム・
// 編集フォーム) が表示する親作品の名前を、前後の空白を落として返す。空文字列は表示できる
// 名前が無いことを表し、ページ見出しは汎用のページ名へフォールバックし、文書タイトルは作品を
// 省く。これにより、どの作品に名前があるかの判断が両者でずれない。
func DBEpisodeListWorkName(workTitle string) string {
	return strings.TrimSpace(workTitle)
}

// DBEpisodeIdentifierは文書タイトルの中でエピソードを名指しするラベルを返す。ラベルは
// 表示用話数で、一覧・行のリンク・編集者自身の語彙がエピソードを指すときの呼び方に揃える。
// 表示用話数が無いエピソードはIDにフォールバックし、ラベルが常に一覧と突き合わせられる
// ものを名指しするようにする。
func DBEpisodeIdentifier(ctx context.Context, episode *model.Episode) string {
	if number := strings.TrimSpace(derefString(episode.Number)); number != "" {
		return number
	}

	return i18n.T(ctx, "db_episodes_identifier_without_number", map[string]any{"EpisodeID": episode.ID.String()})
}

// DBEpisodeNameはページの文章の中でエピソードをどう名指しするかを返す。Railsの
// Episode#number_titleと同じ書式で表示用話数とタイトルを並べ、どちらか一方しか無い場合はその
// 一方に絞る。どちらも無いエピソードはIDにフォールバックし、文が一覧と突き合わせられるものを
// 名指しし続けるようにする。
//
// これは文書タイトルではなく文章の中で使う名前である。文書タイトルに必要な、より短いラベルは
// DBEpisodeIdentifierが担う。
func DBEpisodeName(ctx context.Context, episode *model.Episode) string {
	number := strings.TrimSpace(derefString(episode.Number))
	title := strings.TrimSpace(derefString(episode.Title))

	switch {
	case number != "" && title != "":
		return i18n.T(ctx, "db_episodes_name", map[string]any{"Number": number, "Title": title})
	case number != "":
		return number
	case title != "":
		return i18n.T(ctx, "db_episodes_name_title_only", map[string]any{"Title": title})
	default:
		return i18n.T(ctx, "db_episodes_name_without_number_and_title", map[string]any{"EpisodeID": episode.ID.String()})
	}
}

// DBEpisodeListItemはAnnict DB管理画面の、ある作品のエピソード一覧で1行ごとに
// 表示する整形済みデータ。
type DBEpisodeListItem struct {
	ID EpisodeID
	// PublishedPositionは作品の公開中のエピソードのうち何件目かを表示用に整形したもの。
	// 非公開のエピソードは連番を持たず空文字列で、テンプレート側で "-" を表示する。
	PublishedPosition string
	// WorkIDは親作品。行のIDリンクがエピソードの公開URLを組み立てるのに使う。
	WorkID WorkID
	// Numberは表示用の話数 (episodes.number、例: "第2話")、RawNumberは数値の話数
	// (episodes.raw_number) を表示用に整形したもの。未設定ならいずれも空文字列で、
	// テンプレート側で "-" のプレースホルダーを表示する。
	Number    string
	RawNumber string
	// TitleとTitleEnは日本語・英語のタイトル。未設定なら空文字列。
	Title   string
	TitleEn string
	// PrevNumberはsort_number順でこのエピソードの直前に来るエピソードを表示用に
	// 整形して名指しする。該当が無ければ空文字列。作品の最初のエピソードと、直前の
	// エピソードがどちらの話数も持たない場合に空になる。
	PrevNumber          string
	SortNumber          int32
	EpisodeRecordsCount int32
	Status              PublishingStatus
}

// NewDBEpisodeListItemsは一覧1ページ分のエピソードを表示用の行に変換する。
func NewDBEpisodeListItems(episodes []*model.Episode) []DBEpisodeListItem {
	result := make([]DBEpisodeListItem, len(episodes))
	for i, episode := range episodes {
		result[i] = NewDBEpisodeListItem(episode)
	}
	return result
}

// NewDBEpisodeListItemは1件のエピソードを表示用の行に変換する。
func NewDBEpisodeListItem(episode *model.Episode) DBEpisodeListItem {
	return DBEpisodeListItem{
		ID:                  EpisodeID(episode.ID),
		PublishedPosition:   formatPublishedPosition(episode.PublishedPosition),
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

// formatPublishedPositionは作品内の連番を描画し、連番を持たないエピソードでは空文字列を返す。
func formatPublishedPosition(position *int64) string {
	if position == nil {
		return ""
	}

	return strconv.FormatInt(*position, 10)
}

// formatPrevNumberは直前のエピソードを表示用話数で描画し、表示用話数を持たない場合は
// 話数にフォールバックする。このフォールバックにより、直前のエピソードが存在するのに
// 話数しか付けられていない場合でも、列が「直前のエピソードなし」と読めてしまうのを防ぐ。
func formatPrevNumber(episode *model.Episode) string {
	if number := derefString(episode.PrevNumber); number != "" {
		return number
	}

	return formatRawNumber(episode.PrevRawNumber)
}

// DBEpisodeFormInputはエピソード編集フォームが各欄に描画する値を保持する。値を文字列で
// 持つのはフォームが往復させるものが文字列であるため。未設定の話数と編集者が消した話数は
// どちらも同じ空の入力であり、却下された送信はパースできた値ではなく入力された内容から
// 再描画する。
//
// テンプレートはフィールド名による引き当てではなく各フィールドを直接読む。編集できる欄は
// 5つですべてコンパイル時に分かっているため、名前で持つことでテンプレートの参照が型検査を
// 受けられる。UpdatedAtは編集できる値ではなくhiddenの版として同居する。
type DBEpisodeFormInput struct {
	Number     string
	RawNumber  string
	SortNumber string
	Title      string
	TitleEn    string
	// UpdatedAtはフォームを開いた時点の版で、hiddenで持ち回る。古い読み取りに対する
	// 送信を、間に書いた人の変更を黙って上書きせずに更新側で却下できるようにするため。
	// 入力された値と一緒に持つのは、却下された送信では編集者が送った版をそのまま返す必要が
	// あるからで、サーバーの現在値を読み直すと、次の送信が守るべき変更をかえって上書きして
	// しまう。
	//
	// updated_atを持たないエピソードではFormNullVersionになる。この明示値は、版を指定して
	// いない空の要求とは区別する。
	UpdatedAt string
}

// NewDBEpisodeFormInputFromEpisodeは保存済みのエピソードを、編集フォームが描画する
// フォーム値に射影する。未設定のカラムは "" になるため、表示用話数を持たないエピソードは
// 編集者が消す必要のあるプレースホルダーではなく空の欄で開く。
func NewDBEpisodeFormInputFromEpisode(episode *model.Episode) DBEpisodeFormInput {
	return DBEpisodeFormInput{
		Number:     derefString(episode.Number),
		RawNumber:  formatRawNumber(episode.RawNumber),
		SortNumber: strconv.FormatInt(int64(episode.SortNumber), 10),
		Title:      derefString(episode.Title),
		TitleEn:    episode.TitleEn,
		UpdatedAt:  formatFormVersion(episode.UpdatedAt),
	}
}

// NewDBEpisodeFormInputFromSubmitは却下された送信の値を、入力されたまま再描画する
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

// DBEpisodeGenerationSummaryは作品のエピソード一覧がテーブルの上に出す案内。作品の
// 予定エピソード数、公開中のエピソード数、しょぼいカレンダー由来の自動生成が
// 作品のスロットからどこまで話数を振れるかを表す。
type DBEpisodeGenerationSummary struct {
	// PlannedCountは作品の予定エピソード数を表示用に整形したもので、作品が記録していなければ
	// 空文字列。テンプレートはRailsの案内と同じく、その欠落に「不明」の文言を描画する。
	PlannedCount                string
	PublishedEpisodeCount       int64
	MaxGeneratableEpisodeNumber int64
}

// NewDBEpisodeGenerationSummaryは、作品の予定エピソード数 (記録が無ければnil)、公開中の
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

// DBEpisodeManualCreationRestrictionは、作品のエピソードを手動作成できない理由の
// Presentation層への射影。一括作成ページは理由ごとの警告を描画し、フォームの無効化も同じ値
// から決めるため、複数の条件のどれが優先されるかをページ側で導出せず1つの値を読む。
//
// PublishingStatusと同じく定数はリテラルで書き、テンプレートがmodelへ直接依存しないように
// する。ドメインenumとの対応はNewDBEpisodeManualCreationRestrictionが固定し、両者を比較
// するテストで担保する。
type DBEpisodeManualCreationRestriction string

const (
	DBEpisodeManualCreationAllowed        DBEpisodeManualCreationRestriction = ""
	DBEpisodeManualCreationEpisodesFilled DBEpisodeManualCreationRestriction = "episodes_filled"
	DBEpisodeManualCreationSlotsExist     DBEpisodeManualCreationRestriction = "slots_exist"
)

// NewDBEpisodeManualCreationRestrictionは作品の手動作成状態を、ページが述べる理由へ
// 射影する。
func NewDBEpisodeManualCreationRestriction(state model.ManualEpisodeCreationState) DBEpisodeManualCreationRestriction {
	return DBEpisodeManualCreationRestriction(state.Restriction())
}

// Restrictedはページが警告すべき理由があるかを返す。
func (r DBEpisodeManualCreationRestriction) Restricted() bool {
	return r != DBEpisodeManualCreationAllowed
}

// formatRawNumberはエピソードの話数を末尾の0を付けずに描画する。整数の話数は
// "2"、0.5話は小数のまま "2.5" と読める。未設定では "" を返し、テンプレート側で欠落の
// 描画方法を決められるようにする。
func formatRawNumber(rawNumber *float64) string {
	if rawNumber == nil {
		return ""
	}

	return strconv.FormatFloat(*rawNumber, 'f', -1, 64)
}
