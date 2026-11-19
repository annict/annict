package viewmodel

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/model"
)

// dbRevisionDefaultTimeZoneは、ログインしていない閲覧者とタイムゾーンを読めない閲覧者に、版の日時を
// 示すタイムゾーン。Annictの利用者の大半が日本にいるため。
const dbRevisionDefaultTimeZone = "Asia/Tokyo"

// dbEditEditorAvatarImageSizeは編集者のアバター画像の幅。高解像度の画面に合わせて、表示する幅 (20px) の2倍にする。
const dbEditEditorAvatarImageSize = 40

// DBRevisionEditorKindは版の編集者の種類。
type DBRevisionEditorKind string

const (
	DBRevisionEditorUser        DBRevisionEditorKind = "user"
	DBRevisionEditorSystem      DBRevisionEditorKind = "system"
	DBRevisionEditorDeletedUser DBRevisionEditorKind = "deleted_user"
)

// DBEditHeaderは操作の編集者・操作・日時。各リソースの履歴では版ごとに、Annict DB全体の編集履歴では
// 操作ごとに示す。
type DBEditHeader struct {
	EditorKind DBRevisionEditorKind
	// EditorUsernameは編集者のアットネーム。EditorKindがDBRevisionEditorUserのときだけ値を持つ。
	EditorUsername string
	// EditorAvatarURLは編集者のアバター画像のURL。プロフィール画像の無い編集者と、編集者のいない操作では空。
	EditorAvatarURL string
	// ActionLabelは名詞の操作 (「更新」など) で、表のcaptionに使う。
	ActionLabel string
	// SentenceBeforeEditorとSentenceAfterEditorは、編集者と操作の文 (「@editorが更新しました」など) の、
	// 編集者の前と後ろの文字列。編集者をプロフィールへのリンクにできるよう、文を編集者の位置で分けて持つ。
	SentenceBeforeEditor string
	SentenceAfterEditor  string
	// CreatedAtDisplayは閲覧者のタイムゾーンで整形した日時、CreatedAtDateTimeは <time datetime> に入れる
	// 機械可読の日時 (UTCのRFC 3339)。
	CreatedAtDisplay  string
	CreatedAtDateTime string
}

// EditorInitialは、プロフィール画像の無い編集者のアバターの代わりに示す、アットネームの頭文字を返す。
func (h DBEditHeader) EditorInitial() string {
	for _, r := range h.EditorUsername {
		return string(r)
	}
	return ""
}

// newDBEditHeaderは操作から編集者・操作・日時を作る。
func newDBEditHeader(ctx context.Context, edit *model.DBEdit, loc *time.Location, helper *image.Helper) DBEditHeader {
	h := DBEditHeader{
		EditorKind:        DBRevisionEditorDeletedUser,
		ActionLabel:       i18n.T(ctx, "db_revision_action_"+string(edit.Action)),
		CreatedAtDisplay:  edit.CreatedAt.In(loc).Format("2006-01-02 15:04"),
		CreatedAtDateTime: edit.CreatedAt.UTC().Format(time.RFC3339),
	}
	switch {
	case edit.Actor == model.DBEditActorSystem:
		h.EditorKind = DBRevisionEditorSystem
	case edit.User != nil:
		h.EditorKind = DBRevisionEditorUser
		h.EditorUsername = edit.User.Username
		if edit.User.ProfileImageData.Valid && helper != nil {
			h.EditorAvatarURL = helper.GetAvatarImageURL(edit.User.ProfileImageData.String, dbEditEditorAvatarImageSize, "webp")
		}
	}
	sentence := i18n.T(ctx, "db_revision_sentence_"+string(edit.Action), map[string]any{"Editor": dbEditSentenceEditorMarker})
	h.SentenceBeforeEditor, h.SentenceAfterEditor, _ = strings.Cut(sentence, dbEditSentenceEditorMarker)
	return h
}

// dbEditSentenceEditorMarkerは、編集者と操作の文で編集者が入る位置の印。利用者の入力や翻訳の文面に現れない
// 制御文字で囲む。
const dbEditSentenceEditorMarker = "\x00editor\x00"

// DBRevisionはAnnict DBの履歴に並べる1つの版。
type DBRevision struct {
	DBEditHeader
	// IsCreateは作成の版で、差分の表に変更前の列を出さない。
	IsCreate bool
	Fields   []DBRevisionField
	Rows     []DBRevisionRow
	// Episodeは作品の履歴に並べるエピソードの版のときの、そのエピソード。作品の版とエピソードの履歴ではnil。
	Episode *DBRevisionEpisode
	// WorkNameは作品の履歴に並べる作品の版のときの、版の時点の作品名。エピソードの版と、エピソード・人物の履歴では空。
	WorkName string
	// TargetNameは、Annict DB全体の編集履歴で表の見出しに入れる対象のリソースの名前。各リソースの履歴では空。
	TargetName string
}

// DBRevisionEpisodeは作品の履歴に並べるエピソードの版が指すエピソード。
type DBRevisionEpisode struct {
	ID EpisodeID
	// Labelは版の時点の話数とタイトルを空白でつないだもの。どちらも無いときは空文字列。
	Label string
}

// DisplayNameは版が指すエピソードの表示名を返す。エピソードの履歴の画面と同じく、話数もタイトルも
// 無いときは名前の無いエピソードと示す。
func (e DBRevisionEpisode) DisplayName(ctx context.Context) string {
	return dbEpisodeDisplayName(ctx, e.Label)
}

// DBRevisionFieldは版で変わった1つの項目。
type DBRevisionField struct {
	Label  string
	Before string
	After  string
}

// DBRevisionRowは版で追加・削除した別表の1行。
type DBRevisionRow struct {
	Added      bool
	TableLabel string
	// Valuesは行の項目を「項目名: 値」の形で並べた文字列。
	Values string
}

// DBRevisionLocationは閲覧者のタイムゾーンを返す。ログインしていないときとタイムゾーンを読めないときは
// 日本時間にする。
func DBRevisionLocation(viewer *model.User) *time.Location {
	if viewer != nil && viewer.TimeZone != "" {
		if loc, err := time.LoadLocation(viewer.TimeZone); err == nil {
			return loc
		}
	}
	loc, err := time.LoadLocation(dbRevisionDefaultTimeZone)
	if err != nil {
		return time.UTC
	}
	return loc
}

// NewDBRevisionsFromAnimeRevisionsは作品・エピソードの版を、履歴に並べる版に変換する。作品の履歴に並べる
// エピソードの版には、そのエピソードを示すEpisodeを入れ、作品の版には版の時点の作品名を入れる。作品の版が並ぶのは
// 作品の履歴だけのため、エピソードの履歴の版には作品名が入らない。差分は各版の
// 1つ前の版 (Previous) と比べて求め、1つ前の版が無い版は作成の版として変更後だけを示す。namesは、IDを値に持つ項目を
// 名前で示すための対応表。
func NewDBRevisionsFromAnimeRevisions(ctx context.Context, revisions []*model.AnimeRevision, loc *time.Location, helper *image.Helper, names DBRevisionIDNames) []DBRevision {
	result := make([]DBRevision, len(revisions))
	for i, r := range revisions {
		result[i] = newDBRevision(ctx, r.DBEdit, newDBEditHeader(ctx, r.DBEdit, loc, helper), r.Previous == nil, dbAnimeRevisionDiff(r.Previous, r.Snapshot), names)
		if r.Episode != nil {
			result[i].Episode = &DBRevisionEpisode{
				ID:    EpisodeID(r.Episode.ID),
				Label: DBEpisodeRevisionLabel(r.Snapshot.NumberText, r.Snapshot.Title),
			}
		} else if r.Snapshot.Kind == model.AnimeClassificationKindWork {
			result[i].WorkName = dbWorkDisplayName(ctx, r.Snapshot.Title)
		}
	}
	return result
}

// dbWorkRevisionHiddenFieldsは、作品の版の差分の表に出さない項目。種類 (作品かエピソードか) は作品のフォームに無く、
// 作品の版では常に作品のため、編集者が読んでも意味を持たない。
var dbWorkRevisionHiddenFields = map[model.RevisionField]bool{
	"kind": true,
}

// dbEpisodeRevisionHiddenFieldsは、エピソードの版の差分の表に出さない項目。分類の項目はエピソードのフォームに無く、
// 編集者が読んでも意味を持たない。
var dbEpisodeRevisionHiddenFields = map[model.RevisionField]bool{
	"kind":            true,
	"parent_anime_id": true,
	"standalone":      true,
}

// dbAnimeRevisionDiffは、作品・エピソードの版の差分を表に出す項目に絞って返す。
func dbAnimeRevisionDiff(before *model.AnimeSnapshot, after model.AnimeSnapshot) model.RevisionDiff {
	diff := model.DiffAnimeSnapshots(before, after)
	hidden := dbWorkRevisionHiddenFields
	if after.Kind == model.AnimeClassificationKindEpisode {
		hidden = dbEpisodeRevisionHiddenFields
	}
	diff.Fields = slices.DeleteFunc(diff.Fields, func(f model.RevisionFieldChange) bool {
		return hidden[f.Field]
	})
	return diff
}

// newDBRevisionは版の差分から、履歴に並べる版を作る。headerは版が属する操作の編集者・操作・日時で、
// Annict DB全体の編集履歴では1つの操作に属する版のあいだで共有する。
func newDBRevision(ctx context.Context, edit *model.DBEdit, header DBEditHeader, noPrevious bool, diff model.RevisionDiff, names DBRevisionIDNames) DBRevision {
	v := DBRevision{
		DBEditHeader: header,
		IsCreate:     noPrevious || edit.Action == model.DBEditActionCreate,
	}

	for _, f := range diff.Fields {
		v.Fields = append(v.Fields, DBRevisionField{
			Label:  dbRevisionFieldLabel(ctx, f.Field),
			Before: dbRevisionFieldValue(ctx, names, f.Field, dbRevisionValue(f.Before)),
			After:  dbRevisionFieldValue(ctx, names, f.Field, dbRevisionValue(f.After)),
		})
	}
	for _, r := range diff.Rows {
		values := make([]string, 0, len(r.Row.Values))
		for _, value := range r.Row.Values {
			if value.Value == nil || *value.Value == "" {
				continue
			}
			// 別表の行の項目は、同じ列名でもリソースの項目と意味が違うことがある (シーズンのnameは季節) ため、
			// 別のキーで見出しを引く。
			label := dbRevisionValueLabel(ctx, dbRevisionRowValueKeys[r.Row.Table][value.Field], *value.Value)
			values = append(values, i18n.T(ctx, "db_revision_row_field_"+string(value.Field))+": "+label)
		}
		v.Rows = append(v.Rows, DBRevisionRow{
			Added:      r.Kind == model.RevisionRowAdded,
			TableLabel: i18n.T(ctx, "db_revision_table_"+string(r.Row.Table)),
			Values:     strings.Join(values, " / "),
		})
	}
	return v
}

func dbRevisionFieldLabel(ctx context.Context, field model.RevisionField) string {
	return i18n.T(ctx, "db_revision_field_"+string(field))
}

// 以下は、選択式の項目の保存した値ごとの言葉の翻訳キー。保存した値 (archived・syobocalなど) はデータベースの中の
// 言葉で編集者には読みにくいため、フォームや一覧と同じ言葉で示す。翻訳キーを使用箇所からgrepで追える状態に保つため、
// 値から組み立てずに直書きのキーで引く。

// dbRevisionBooleanValuesは、真偽値の項目の言葉の翻訳キー。
var dbRevisionBooleanValues = map[string]string{
	"true":  "db_revision_boolean_true",
	"false": "db_revision_boolean_false",
}

// dbRevisionFieldValueKeysは、作品・エピソード・人物の版の項目ごとの、値の言葉の翻訳キー。作品と人物で同じ名前の
// 項目は「状態」(status) だけで、どちらも同じ値を持つため、項目名だけで引く。
var dbRevisionFieldValueKeys = map[model.RevisionField]map[string]string{
	"status": {
		string(model.AnimeStatusPublished): "db_revision_status_published",
		string(model.AnimeStatusArchived):  "db_revision_status_archived",
		string(model.PersonStatusDelisted): "db_revision_status_delisted",
	},
	"media": {
		string(model.AnimeMediaTV):    "db_revision_media_tv",
		string(model.AnimeMediaOVA):   "db_revision_media_ova",
		string(model.AnimeMediaMovie): "db_revision_media_movie",
		string(model.AnimeMediaONA):   "db_revision_media_ona",
		string(model.AnimeMediaOther): "db_revision_media_other",
	},
	"release_status": {
		string(model.ReleaseStatusNotYetReleased): "db_revision_release_status_not_yet_released",
		string(model.ReleaseStatusReleased):       "db_revision_release_status_released",
		string(model.ReleaseStatusCancelled):      "db_revision_release_status_cancelled",
	},
	"standalone": dbRevisionBooleanValues,
	"gender": {
		string(model.PersonGenderMale):   "db_revision_gender_male",
		string(model.PersonGenderFemale): "db_revision_gender_female",
	},
	"blood_type": {
		string(model.PersonBloodTypeA):  "db_revision_blood_type_a",
		string(model.PersonBloodTypeB):  "db_revision_blood_type_b",
		string(model.PersonBloodTypeAB): "db_revision_blood_type_ab",
		string(model.PersonBloodTypeO):  "db_revision_blood_type_o",
	},
}

// dbRevisionRowValueKeysは、作品の別表の行の項目ごとの、値の言葉の翻訳キー。同じ項目名でも別表によって値の意味が
// 違う (リンクの種類とイベントの種類など) ため、別表と項目名の組で引く。
var dbRevisionRowValueKeys = map[model.RevisionTable]map[model.RevisionField]map[string]string{
	model.RevisionTableExternalIDs: {
		"service": {
			string(model.AnimeExternalServiceSyobocal): "db_revision_external_service_syobocal",
			string(model.AnimeExternalServiceMal):      "db_revision_external_service_mal",
		},
	},
	model.RevisionTableOfficialAccounts: {
		"service": {
			string(model.AnimeAccountServiceBluesky):   "db_revision_account_service_bluesky",
			string(model.AnimeAccountServiceInstagram): "db_revision_account_service_instagram",
			string(model.AnimeAccountServiceLine):      "db_revision_account_service_line",
			string(model.AnimeAccountServiceMastodon):  "db_revision_account_service_mastodon",
			string(model.AnimeAccountServiceMixi2):     "db_revision_account_service_mixi2",
			string(model.AnimeAccountServiceThreads):   "db_revision_account_service_threads",
			string(model.AnimeAccountServiceTiktok):    "db_revision_account_service_tiktok",
			string(model.AnimeAccountServiceX):         "db_revision_account_service_x",
			string(model.AnimeAccountServiceYoutube):   "db_revision_account_service_youtube",
		},
	},
	model.RevisionTableLinks: {
		"kind": {
			string(model.AnimeLinkKindOfficialSite): "db_revision_link_kind_official_site",
			string(model.AnimeLinkKindWikipedia):    "db_revision_link_kind_wikipedia",
			string(model.AnimeLinkKindOther):        "db_revision_link_kind_other",
		},
		"language": {
			string(model.LanguageJa):    "db_revision_language_ja",
			string(model.LanguageEn):    "db_revision_language_en",
			string(model.LanguageOther): "db_revision_language_other",
		},
	},
	model.RevisionTableSeasons: {
		"name": {
			string(model.SeasonNameWinter): "db_revision_season_winter",
			string(model.SeasonNameSpring): "db_revision_season_spring",
			string(model.SeasonNameSummer): "db_revision_season_summer",
			string(model.SeasonNameFall):   "db_revision_season_fall",
		},
		"is_primary": dbRevisionBooleanValues,
	},
	model.RevisionTableEvents: {
		"kind": {
			string(model.AnimeEventKindBroadcast):        "db_revision_event_kind_broadcast",
			string(model.AnimeEventKindRevivalScreening): "db_revision_event_kind_revival_screening",
			string(model.AnimeEventKindOther):            "db_revision_event_kind_other",
		},
	},
}

// DBRevisionIDNamesは、差分の表で、IDを値に持つ項目 (話数フォーマット・出身地) を名前で示すための対応表。
// 版ごとに読まずに済むよう、履歴を読むユースケースが1回ずつ読んだ一覧から作る。ゼロ値は対応を持たず、
// IDのまま示す。
type DBRevisionIDNames struct {
	numberFormats map[string]string
	prefectures   map[string]string
}

// NewDBRevisionIDNamesは、話数フォーマットと都道府県の一覧から、版の項目の値 (IDの文字列) で名前を引く対応表を作る。
func NewDBRevisionIDNames(numberFormats []model.NumberFormat, prefectures []model.Prefecture) DBRevisionIDNames {
	names := DBRevisionIDNames{
		numberFormats: make(map[string]string, len(numberFormats)),
		prefectures:   make(map[string]string, len(prefectures)),
	}
	for _, f := range numberFormats {
		names.numberFormats[f.ID.String()] = f.Name
	}
	for _, p := range prefectures {
		names.prefectures[p.ID.String()] = p.Name
	}
	return names
}

// dbRevisionFieldValueは、作品・エピソード・人物の版の項目の値を、表に示す文字列にして返す。IDを値に持つ項目は
// 名前で、選択式の項目は言葉で示し、どちらも持たない値 (一覧に無いIDを含む) は保存した値のまま返す。
func dbRevisionFieldValue(ctx context.Context, names DBRevisionIDNames, field model.RevisionField, v string) string {
	var idNames map[string]string
	switch field {
	case "number_format_id":
		idNames = names.numberFormats
	case "prefecture_id":
		idNames = names.prefectures
	}
	if name, ok := idNames[v]; ok {
		return name
	}
	return dbRevisionValueLabel(ctx, dbRevisionFieldValueKeys[field], v)
}

// dbRevisionValueLabelは、選択式の項目の値を言葉にして返す。言葉を持たない値 (選択式でない項目の値と、選択肢に
// 無い値) は、値が空にならないよう保存した値のまま返す。
func dbRevisionValueLabel(ctx context.Context, keys map[string]string, v string) string {
	if key, ok := keys[v]; ok {
		return i18n.T(ctx, key)
	}
	return v
}

func dbRevisionValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// DBWorkRevisionWorkは作品の履歴の見出しとサブナビに使う作品の表示用データ。
type DBWorkRevisionWork struct {
	ID         WorkID
	Title      string
	NoEpisodes bool
}

// NewDBWorkRevisionWorkは作品名の前後の空白を落とし、文書タイトルと見出しで同じ表示名を使う。
func NewDBWorkRevisionWork(work *model.Work) DBWorkRevisionWork {
	return DBWorkRevisionWork{ID: WorkID(work.ID), Title: strings.TrimSpace(work.Title), NoEpisodes: work.NoEpisodes}
}

// Headingは作品名を返す。表示名が無い作品では履歴のページ名へフォールバックする。
func (w DBWorkRevisionWork) Heading(ctx context.Context) string {
	if w.Title != "" {
		return w.Title
	}
	return i18n.T(ctx, "db_work_revisions_index_title")
}

// DBEpisodeRevisionLabelは、エピソードの話数とタイトルを空白でつなぎ、エピソードを指す表示名にする。
// 空の値は除き、どちらも無いときは空文字列を返す。
func DBEpisodeRevisionLabel(number, title *string) string {
	parts := make([]string, 0, 2)
	for _, v := range []*string{number, title} {
		if v != nil && strings.TrimSpace(*v) != "" {
			parts = append(parts, strings.TrimSpace(*v))
		}
	}
	return strings.Join(parts, " ")
}

// DBEpisodeRevisionEpisodeはエピソードの履歴の見出しとサブナビに使うエピソードと親作品の表示用データ。
type DBEpisodeRevisionEpisode struct {
	ID    EpisodeID
	Label string
	Work  DBWorkRevisionWork
}

// NewDBEpisodeRevisionEpisodeは、現在の話数とタイトルからエピソードの表示名を作る。
func NewDBEpisodeRevisionEpisode(episode *model.Episode, work *model.Work) DBEpisodeRevisionEpisode {
	return DBEpisodeRevisionEpisode{
		ID:    EpisodeID(episode.ID),
		Label: DBEpisodeRevisionLabel(episode.Number, episode.Title),
		Work:  NewDBWorkRevisionWork(work),
	}
}

// DisplayNameは文書タイトルで使う表示名を返す。話数もタイトルも無いときは、名前の無いエピソードと示す。
func (e DBEpisodeRevisionEpisode) DisplayName(ctx context.Context) string {
	return dbEpisodeDisplayName(ctx, e.Label)
}

// dbEpisodeDisplayNameは、DBEpisodeRevisionLabelで作った表示名を返し、空なら名前の無いエピソードと示す。
// 作品の履歴とエピソードの履歴で、同じエピソードを同じ名前で呼ぶため、両方のViewModelがこの規則を使う。
func dbEpisodeDisplayName(ctx context.Context, label string) string {
	if label != "" {
		return label
	}
	return i18n.T(ctx, "db_revision_episode_untitled")
}

// Headingは版の一覧の見出しを返す。話数もタイトルも無いエピソードでは、ページ名にする。
func (e DBEpisodeRevisionEpisode) Heading(ctx context.Context) string {
	if e.Label == "" {
		return i18n.T(ctx, "db_episode_revisions_index_title")
	}
	return i18n.T(ctx, "db_episode_revisions_index_heading", map[string]any{"Episode": e.Label})
}

// NewDBRevisionsFromPersonRevisionsは人物の版を、履歴に並べる版に変換する。差分は各版の1つ前の版 (Previous) と
// 比べて求める。1つ前の版は管理者以外に隠す版も含むため、隠した版の変更が次の版の変更として出ることはない。
func NewDBRevisionsFromPersonRevisions(ctx context.Context, revisions []*model.PersonRevision, loc *time.Location, helper *image.Helper, names DBRevisionIDNames) []DBRevision {
	result := make([]DBRevision, len(revisions))
	for i, r := range revisions {
		result[i] = newDBRevision(ctx, r.DBEdit, newDBEditHeader(ctx, r.DBEdit, loc, helper), r.Previous == nil, model.DiffPersonSnapshots(r.Previous, r.Snapshot), names)
	}
	return result
}

// DBPersonRevisionPersonは人物の履歴の見出しに使う人物の表示用データ。
type DBPersonRevisionPerson struct {
	ID   PersonID
	Name string
	// Delistedは掲載停止中の人物か。掲載停止中の人物は編集できないため、編集画面へのリンクを出さない。
	Delisted bool
}

// NewDBPersonRevisionPersonは人物名の前後の空白を落とし、文書タイトルと見出しで同じ表示名を使う。
func NewDBPersonRevisionPerson(person *model.Person) DBPersonRevisionPerson {
	return DBPersonRevisionPerson{
		ID:       PersonID(person.ID),
		Name:     strings.TrimSpace(person.Name),
		Delisted: person.Status == model.PersonStatusDelisted,
	}
}

// Headingは人物名を返す。表示名が無い人物では履歴のページ名へフォールバックする。
func (p DBPersonRevisionPerson) Heading(ctx context.Context) string {
	if p.Name != "" {
		return p.Name
	}
	return i18n.T(ctx, "db_person_revisions_index_title")
}

// DBEditTargetKindはAnnict DB全体の編集履歴で、版が指すリソースの種類。
type DBEditTargetKind string

const (
	DBEditTargetWork    DBEditTargetKind = "work"
	DBEditTargetEpisode DBEditTargetKind = "episode"
	DBEditTargetPerson  DBEditTargetKind = "person"
)

// DBEditはAnnict DB全体の編集履歴に並べる1つの操作。
type DBEdit struct {
	DBEditHeader
	// Revisionsは操作に属する、閲覧者に見せる版。作品・エピソードの版、人物の版の順に並ぶ。
	Revisions []DBEditRevision
}

// DBEditRevisionは操作に属する1つの版と、版が指すリソース。
type DBEditRevision struct {
	Kind DBEditTargetKind
	// WorkID・EpisodeID・PersonIDは、Kindに対応する1つだけが値を持つ。
	WorkID    WorkID
	EpisodeID EpisodeID
	PersonID  PersonID
	// Nameは版の時点のリソースの名前。エピソードでは、今の親作品の名前を添える。
	Name     string
	Revision DBRevision
}

// NewDBEditsはAnnict DB全体の編集履歴の操作を、並べる操作に変換する。版の差分は、各版の1つ前の版 (Previous) と
// 比べて求める。リソースの名前は版の時点のスナップショットから取る。
func NewDBEdits(ctx context.Context, edits []*model.DBEdit, loc *time.Location, helper *image.Helper, names DBRevisionIDNames) []DBEdit {
	result := make([]DBEdit, len(edits))
	for i, edit := range edits {
		header := newDBEditHeader(ctx, edit, loc, helper)
		v := DBEdit{DBEditHeader: header}
		for _, r := range edit.AnimeRevisions {
			revision := DBEditRevision{
				Revision: newDBRevision(ctx, edit, header, r.Previous == nil, dbAnimeRevisionDiff(r.Previous, r.Snapshot), names),
			}
			if r.Episode != nil {
				revision.Kind = DBEditTargetEpisode
				revision.EpisodeID = EpisodeID(r.Episode.ID)
				revision.Name = dbEditEpisodeName(ctx, DBEpisodeRevisionLabel(r.Snapshot.NumberText, r.Snapshot.Title), r.Work)
			} else {
				revision.Kind = DBEditTargetWork
				if r.Work != nil {
					revision.WorkID = WorkID(r.Work.ID)
				}
				revision.Name = dbWorkDisplayName(ctx, r.Snapshot.Title)
			}
			revision.Revision.TargetName = revision.Name
			v.Revisions = append(v.Revisions, revision)
		}
		for _, r := range edit.PersonRevisions {
			revision := DBEditRevision{
				Kind:     DBEditTargetPerson,
				PersonID: PersonID(r.PersonID),
				Name:     dbPersonDisplayName(ctx, r.Snapshot.Name),
				Revision: newDBRevision(ctx, edit, header, r.Previous == nil, model.DiffPersonSnapshots(r.Previous, r.Snapshot), names),
			}
			revision.Revision.TargetName = revision.Name
			v.Revisions = append(v.Revisions, revision)
		}
		result[i] = v
	}
	return result
}

// dbEditEpisodeNameは、版の時点のエピソードの名前に、今の親作品の名前を添える。全体の編集履歴は作品をまたいで
// 操作を並べるため、話数とタイトルだけではどの作品のエピソードかが分からないため。版の時点の親作品の名前は
// 親作品の版を別に引かないと分からないため、リンク先のエピソードの履歴と同じく今の親作品の名前を使う。
func dbEditEpisodeName(ctx context.Context, label string, work *model.Work) string {
	name := dbEpisodeDisplayName(ctx, label)
	if work == nil {
		return name
	}
	return i18n.T(ctx, "db_edits_episode_name", map[string]any{"Episode": name, "Work": dbWorkDisplayName(ctx, &work.Title)})
}

// dbWorkDisplayNameは作品名を返し、空なら名前の無い作品と示す。
func dbWorkDisplayName(ctx context.Context, title *string) string {
	if title != nil && strings.TrimSpace(*title) != "" {
		return strings.TrimSpace(*title)
	}
	return i18n.T(ctx, "db_edits_index_untitled_work")
}

// dbPersonDisplayNameは版の時点の人物名を返し、空なら名前の無い人物と示す。
func dbPersonDisplayName(ctx context.Context, name string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}
	return i18n.T(ctx, "db_edits_index_untitled_person")
}
