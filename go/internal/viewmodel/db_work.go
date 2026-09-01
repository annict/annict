package viewmodel

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// DBWorkListItem is the per-row display data for the work list on the Annict DB admin screen.
//
// [Ja] DBWorkListItem は Annict DB 管理画面の作品一覧で 1 行ごとに表示する整形済みデータ。
type DBWorkListItem struct {
	ID    WorkID
	Title string
	// Alternate titles shown below the main title. Empty when unset; the
	// template renders a "-" placeholder in that case.
	//
	// [Ja] メインタイトルの下に並べる別タイトル。未設定なら空文字列で、
	// テンプレート側で "-" のプレースホルダーを表示する。
	TitleKana string
	TitleEn   string
	// Pre-translated media label (e.g. "TV", "OVA") shown in the media column.
	//
	// [Ja] メディア列に表示する翻訳済みのメディア名 (例: "TV", "OVA")。
	Media string
	// Pre-formatted season display string.
	//
	// [Ja] フォーマット済みのシーズン表示文字列。
	Season string
	// External-service links (Syoboi Calendar / MyAnimeList) shown in the external
	// services column. Each is the zero value when the work has no id for that
	// service, and the template renders a "-" placeholder in that case.
	//
	// [Ja] 外部サービス列に表示するしょぼかる / MyAnimeList のリンク。作品にその外部 ID が
	// 無い場合はゼロ値になり、テンプレートは "-" のプレースホルダーを表示する。
	Syobocal      ExternalServiceLink
	MalAnime      ExternalServiceLink
	WatchersCount int32
	Status        PublishingStatus
	// Thumbnail resolver, which falls back to the placeholder for works with no image.
	// The display width is chosen by the template, so no URL is pre-generated here.
	//
	// [Ja] サムネイルの解決子。画像が無い作品ではプレースホルダーにフォールバックする。
	// 表示幅はテンプレートが決めるため、ここでは URL を生成しない。
	Image WorkImage
}

func NewDBWorkListItems(ctx context.Context, works []*model.Work, helper *image.Helper) []DBWorkListItem {
	result := make([]DBWorkListItem, len(works))
	for i, work := range works {
		result[i] = NewDBWorkListItem(ctx, work, helper)
	}
	return result
}

func NewDBWorkListItem(ctx context.Context, work *model.Work, helper *image.Helper) DBWorkListItem {
	return DBWorkListItem{
		ID:            WorkID(work.ID),
		Title:         work.Title,
		TitleKana:     derefString(work.TitleKana),
		TitleEn:       work.TitleEn,
		Media:         formatMedia(ctx, work.Media),
		Season:        formatSeason(ctx, work.SeasonYear, work.SeasonName),
		Syobocal:      newExternalServiceLink(work.ScTid, SyobocalURL),
		MalAnime:      newExternalServiceLink(work.MalAnimeID, MalAnimeURL),
		WatchersCount: work.WatchersCount,
		Status:        PublishingStatus(work.DerivedStatus()),
		Image:         NewWorkImage(work.ImageData, helper),
	}
}

// formatMedia returns the translated media label for a works.media enum value,
// mirroring the media_* option keys used by the work form. It returns "" for
// unknown values so the template can decide how to render the gap.
//
// [Ja] formatMedia は works.media の enum 値に対応する翻訳済みのメディア名を返す。
// 作品フォームで使う media_* のオプションキーと対応させている。未知の値では ""
// を返し、テンプレート側で欠落の描画方法を決められるようにする。
func formatMedia(ctx context.Context, media int32) string {
	key := ""
	switch media {
	case 1:
		key = "media_tv"
	case 2:
		key = "media_ova"
	case 3:
		key = "media_movie"
	case 4:
		key = "media_web"
	case 0:
		key = "media_other"
	}

	if key == "" {
		return ""
	}

	return i18n.T(ctx, key)
}

// formatSeason returns the release-season display for a work's season_year /
// season_name pair. A work may carry a year without a season, so the year alone is
// shown with a note that the season is unregistered, matching the Rails version,
// which falls back to the yearly label when season_name is blank. An unknown
// season_name enum value takes the same path. It returns "" when season_year is
// unset, and the template renders a "-" placeholder in that case.
//
// [Ja] formatSeason は work の season_year / season_name の組に対するリリース時期の
// 表示を返す。年だけが登録された作品があるため、季節が未登録である旨を添えて年のみを
// 表示する。season_name が空のとき年のラベルにフォールバックする Rails 版に合わせて
// いる。未知の season_name の enum 値も同じ経路を通る。season_year が未設定のときは
// "" を返し、テンプレート側で "-" のプレースホルダーを表示する。
func formatSeason(ctx context.Context, year *int32, name *int32) string {
	if year == nil {
		return ""
	}

	seasonKey := ""
	if name != nil {
		seasonKey = seasonLabelKey(*name)
	}
	if seasonKey == "" {
		return i18n.T(ctx, "year_no_season", map[string]any{
			"Year": *year,
		})
	}

	return i18n.T(ctx, "year_season", map[string]any{
		"Year":   *year,
		"Season": i18n.T(ctx, seasonKey),
	})
}

type SelectOption struct {
	Value string
	Label string
}

type DBWorkFormOptions struct {
	MediaOptions        []SelectOption
	SeasonYearOptions   []SelectOption
	SeasonNameOptions   []SelectOption
	NumberFormatOptions []SelectOption
}

func NewDBWorkFormOptions(ctx context.Context, numberFormats []model.NumberFormat) DBWorkFormOptions {
	return DBWorkFormOptions{
		MediaOptions:        buildMediaOptions(ctx),
		SeasonYearOptions:   buildSeasonYearOptions(),
		SeasonNameOptions:   buildSeasonNameOptions(ctx),
		NumberFormatOptions: buildNumberFormatOptions(numberFormats),
	}
}

func buildMediaOptions(ctx context.Context) []SelectOption {
	return []SelectOption{
		{Value: "1", Label: i18n.T(ctx, "media_tv")},
		{Value: "2", Label: i18n.T(ctx, "media_ova")},
		{Value: "3", Label: i18n.T(ctx, "media_movie")},
		{Value: "4", Label: i18n.T(ctx, "media_web")},
		{Value: "0", Label: i18n.T(ctx, "media_other")},
	}
}

func buildSeasonYearOptions() []SelectOption {
	maxYear := seasonMaxYear()
	options := make([]SelectOption, 0, maxYear-seasonStartYear+1)
	for y := maxYear; y >= seasonStartYear; y-- {
		options = append(options, SelectOption{
			Value: fmt.Sprintf("%d", y),
			Label: fmt.Sprintf("%d", y),
		})
	}
	return options
}

func buildSeasonNameOptions(ctx context.Context) []SelectOption {
	options := make([]SelectOption, len(seasons))
	for i, s := range seasons {
		options[i] = SelectOption{
			Value: strconv.FormatInt(int64(s.value), 10),
			Label: i18n.T(ctx, s.key),
		}
	}
	return options
}

func buildNumberFormatOptions(formats []model.NumberFormat) []SelectOption {
	options := make([]SelectOption, len(formats))
	for i, f := range formats {
		options[i] = SelectOption{
			Value: fmt.Sprintf("%d", f.ID),
			Label: f.Name,
		}
	}
	return options
}

// DBWorkFormInput holds the submitted form values so the work form can be re-rendered with the user's input after a validation error.
//
// [Ja] DBWorkFormInput はバリデーションエラー時に作品フォームを再描画するために、送信された入力値を保持する。
type DBWorkFormInput struct {
	Title                 string
	TitleKana             string
	TitleAlter            string
	TitleEn               string
	TitleAlterEn          string
	Media                 string
	SeasonYear            string
	SeasonName            string
	StartedOn             string
	EndedOn               string
	OfficialSiteURL       string
	OfficialSiteURLEn     string
	WikipediaURL          string
	WikipediaURLEn        string
	TwitterUsername       string
	TwitterHashtag        string
	ScTid                 string
	MalAnimeID            string
	Synopsis              string
	SynopsisSource        string
	SynopsisEn            string
	SynopsisSourceEn      string
	ManualEpisodesCount   string
	StartEpisodeRawNumber string
	NumberFormatID        string
	NoEpisodes            string
	// UpdatedAt is the version the edit form was opened against, carried in a hidden field so
	// the update can reject a submit made against a stale read instead of silently
	// overwriting whoever wrote in between. It travels with the form values because a
	// non-conflict rejection has to echo back the version the editor submitted. On a conflict,
	// the handler instead presents the current stored state and replaces UpdatedAt with that
	// state's version, so the next submit knowingly overwrites exactly the state that was shown.
	//
	// It is FormNullVersion for a work whose updated_at is unset, and empty on the create
	// form, which has no stored row to state a version for.
	//
	// [Ja] UpdatedAt は編集フォームを開いた時点の版で、hidden で持ち回る。古い読み取りに対する
	// 送信を、間に書いた人の変更を黙って上書きせずに更新側で却下できるようにするため。入力値と
	// 一緒に持つのは、非競合の却下では編集者が送った版をそのまま返す必要があるから。競合時は
	// ハンドラーが現在の保存状態を示し、UpdatedAt をその状態の版へ載せ替える。これにより次の送信は、
	// 示された状態を確認したうえで上書きする意味になる。
	//
	// updated_at を持たない作品では FormNullVersion になり、版を示すべき保存済みの行が無い
	// 作成フォームでは空になる。
	UpdatedAt string
}

// NewDBWorkFormInput preserves the submitted work form values so the create or edit form
// can be re-rendered with the user's input when validation fails. It takes the shared
// usecase.WorkFormInput, so the create and update handlers feed the same converter.
//
// [Ja] NewDBWorkFormInput は送信された作品フォームの入力値を保持し、バリデーション失敗時に
// 作成・編集フォームをユーザーの入力のまま再描画できるようにする。共有の usecase.WorkFormInput
// を受け取るため、作成・更新ハンドラーが同じ変換を通す。
func NewDBWorkFormInput(input usecase.WorkFormInput) *DBWorkFormInput {
	return &DBWorkFormInput{
		Title:                 input.Title,
		TitleKana:             input.TitleKana,
		TitleAlter:            input.TitleAlter,
		TitleEn:               input.TitleEn,
		TitleAlterEn:          input.TitleAlterEn,
		Media:                 input.Media,
		SeasonYear:            input.SeasonYear,
		SeasonName:            input.SeasonName,
		StartedOn:             input.StartedOn,
		EndedOn:               input.EndedOn,
		OfficialSiteURL:       input.OfficialSiteURL,
		OfficialSiteURLEn:     input.OfficialSiteURLEn,
		WikipediaURL:          input.WikipediaURL,
		WikipediaURLEn:        input.WikipediaURLEn,
		TwitterUsername:       input.TwitterUsername,
		TwitterHashtag:        input.TwitterHashtag,
		ScTid:                 input.ScTid,
		MalAnimeID:            input.MalAnimeID,
		Synopsis:              input.Synopsis,
		SynopsisSource:        input.SynopsisSource,
		SynopsisEn:            input.SynopsisEn,
		SynopsisSourceEn:      input.SynopsisSourceEn,
		ManualEpisodesCount:   input.ManualEpisodesCount,
		StartEpisodeRawNumber: input.StartEpisodeRawNumber,
		NumberFormatID:        input.NumberFormatID,
		NoEpisodes:            input.NoEpisodes,
	}
}

// NewDBWorkFormInputFromSubmit copies a rejected edit submit's values and version into the
// re-rendered form exactly as submitted. Keeping the submitted version is correct for a
// non-conflict rejection. When rendering a conflict, the caller presents the current stored
// state and then replaces UpdatedAt with that state's version.
//
// [Ja] NewDBWorkFormInputFromSubmit は却下された編集の送信の値と版を、送信されたまま再描画する
// フォームへコピーする。送信された版を保つのは非競合の却下では正しい。競合を描画する呼び出し元は、
// 現在の保存状態を示してから UpdatedAt をその状態の版へ載せ替える。
func NewDBWorkFormInputFromSubmit(input usecase.UpdateWorkInput) *DBWorkFormInput {
	formInput := NewDBWorkFormInput(input.WorkFormInput)
	formInput.UpdatedAt = input.UpdatedAt

	return formInput
}

// NewDBWorkFormInputFromWork projects an existing work onto the string form values
// the work edit form renders. It is the inverse of buildWorkFormParams's
// string->typed conversion: pointers and sql-nullable values become "" when unset,
// dates use the YYYY-MM-DD input format, and the no_episodes checkbox uses "1".
//
// [Ja] NewDBWorkFormInputFromWork は既存の work を、作品編集フォームが描画する
// 文字列のフォーム値に射影する。buildWorkFormParams の文字列→型変換の逆向きで、
// ポインタや NULL 許容値は未設定なら "" に、日付は YYYY-MM-DD 形式に、no_episodes
// チェックボックスは "1" にする。
func NewDBWorkFormInputFromWork(work *model.Work) *DBWorkFormInput {
	return &DBWorkFormInput{
		Title:                 work.Title,
		TitleKana:             derefString(work.TitleKana),
		TitleAlter:            work.TitleAlter,
		TitleEn:               work.TitleEn,
		TitleAlterEn:          work.TitleAlterEn,
		Media:                 strconv.FormatInt(int64(work.Media), 10),
		SeasonYear:            formatNullableInt32(work.SeasonYear),
		SeasonName:            formatNullableInt32(work.SeasonName),
		StartedOn:             formatDateInput(work.StartedOn),
		EndedOn:               formatDateInput(work.EndedOn),
		OfficialSiteURL:       work.OfficialSiteURL,
		OfficialSiteURLEn:     work.OfficialSiteURLEn,
		WikipediaURL:          work.WikipediaURL,
		WikipediaURLEn:        work.WikipediaURLEn,
		TwitterUsername:       derefString(work.TwitterUsername),
		TwitterHashtag:        derefString(work.TwitterHashtag),
		ScTid:                 formatNullableInt32(work.ScTid),
		MalAnimeID:            formatNullableInt32(work.MalAnimeID),
		Synopsis:              work.Synopsis,
		SynopsisSource:        work.SynopsisSource,
		SynopsisEn:            work.SynopsisEn,
		SynopsisSourceEn:      work.SynopsisSourceEn,
		ManualEpisodesCount:   formatNullableInt32(work.ManualEpisodesCount),
		StartEpisodeRawNumber: strconv.FormatFloat(work.StartEpisodeRawNumber, 'f', -1, 64),
		NumberFormatID:        formatNumberFormatID(work.NumberFormatID),
		NoEpisodes:            formatCheckbox(work.NoEpisodes),
		UpdatedAt:             formatFormVersion(work.UpdatedAt),
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatNullableInt32(v *int32) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(int64(*v), 10)
}

func formatNumberFormatID(id *model.NumberFormatID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func formatDateInput(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatCheckbox(checked bool) string {
	if checked {
		return "1"
	}
	return ""
}

// Val returns the form value for the given field, or "" when the receiver is nil. The hidden
// version is reachable under "updated_at" alongside the editable fields, so the template reads
// every input's value the same nil-safe way.
//
// [Ja] Val は指定フィールドのフォーム値を返す。レシーバが nil のときは "" を返す。hidden の版も
// 編集可能なフィールドと並んで "updated_at" で引ける。テンプレートがどの入力欄の値も同じく
// nil 安全な方法で読めるようにするため。
func (d *DBWorkFormInput) Val(field string) string {
	if d == nil {
		return ""
	}
	switch field {
	case "title":
		return d.Title
	case "title_kana":
		return d.TitleKana
	case "title_alter":
		return d.TitleAlter
	case "title_en":
		return d.TitleEn
	case "title_alter_en":
		return d.TitleAlterEn
	case "media":
		return d.Media
	case "season_year":
		return d.SeasonYear
	case "season_name":
		return d.SeasonName
	case "started_on":
		return d.StartedOn
	case "ended_on":
		return d.EndedOn
	case "official_site_url":
		return d.OfficialSiteURL
	case "official_site_url_en":
		return d.OfficialSiteURLEn
	case "wikipedia_url":
		return d.WikipediaURL
	case "wikipedia_url_en":
		return d.WikipediaURLEn
	case "twitter_username":
		return d.TwitterUsername
	case "twitter_hashtag":
		return d.TwitterHashtag
	case "sc_tid":
		return d.ScTid
	case "mal_anime_id":
		return d.MalAnimeID
	case "synopsis":
		return d.Synopsis
	case "synopsis_source":
		return d.SynopsisSource
	case "synopsis_en":
		return d.SynopsisEn
	case "synopsis_source_en":
		return d.SynopsisSourceEn
	case "manual_episodes_count":
		return d.ManualEpisodesCount
	case "start_episode_raw_number":
		return d.StartEpisodeRawNumber
	case "number_format_id":
		return d.NumberFormatID
	case "no_episodes":
		return d.NoEpisodes
	case "updated_at":
		return d.UpdatedAt
	default:
		return ""
	}
}

// LabelLinkURL returns the external link target shown next to a field's label, or ""
// when the field is not linkable or has no value. It mirrors the Rails work form, which
// renders an external-link icon beside the URL / Twitter / Syoboi Calendar / MyAnimeList
// labels once the value is filled in. URL fields link to the submitted value itself,
// while the id/username fields derive their service URL via the shared helpers.
//
// [Ja] LabelLinkURL はフィールドのラベル横に表示する外部リンク先を返す。フィールドがリンク
// 対象でない、または値が無いときは "" を返す。Rails の作品フォーム (URL / Twitter / しょぼかる /
// MyAnimeList のラベル横に、値が入っていれば外部リンクアイコンを出す) に対応させている。URL 系の
// フィールドは送信された値自体をリンク先にし、ID / ユーザー名系は共有ヘルパーでサービス URL を導出する。
func (d *DBWorkFormInput) LabelLinkURL(field string) string {
	if d == nil {
		return ""
	}
	switch field {
	case "official_site_url":
		return d.OfficialSiteURL
	case "official_site_url_en":
		return d.OfficialSiteURLEn
	case "wikipedia_url":
		return d.WikipediaURL
	case "wikipedia_url_en":
		return d.WikipediaURLEn
	case "twitter_username":
		return TwitterUsernameURL(d.TwitterUsername)
	case "twitter_hashtag":
		return TwitterHashtagURL(d.TwitterHashtag)
	case "sc_tid":
		return externalIDURL(d.ScTid, SyobocalURL)
	case "mal_anime_id":
		return externalIDURL(d.MalAnimeID, MalAnimeURL)
	default:
		return ""
	}
}
