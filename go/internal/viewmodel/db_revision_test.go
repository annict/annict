package viewmodel

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/model"
)

func TestNewDBRevisionsFromAnimeRevisions(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	createdAt := time.Date(2026, 11, 8, 15, 30, 0, 0, time.UTC)
	before := "前"
	after := "後"
	userID := model.UserID(1)
	revisions := []*model.AnimeRevision{
		{
			DBEdit:   &model.DBEdit{Actor: model.DBEditActorUser, UserID: &userID, User: &model.User{ID: userID, Username: "editor"}, Action: model.DBEditActionUpdate, CreatedAt: createdAt},
			Previous: &model.AnimeSnapshot{Title: &before, Status: model.AnimeStatusPublished},
			Snapshot: model.AnimeSnapshot{Title: &after, Status: model.AnimeStatusPublished},
		},
		{
			// 退会したユーザーの作成の版。
			DBEdit:   &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionCreate, CreatedAt: createdAt},
			Snapshot: model.AnimeSnapshot{Title: &before, Status: model.AnimeStatusPublished, Seasons: []model.AnimeSnapshotSeason{{Year: 2026, IsPrimary: true}}},
		},
		{
			DBEdit:   &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate, CreatedAt: createdAt},
			Previous: &model.AnimeSnapshot{Title: &before},
			Snapshot: model.AnimeSnapshot{Title: &after},
		},
	}

	got := NewDBRevisionsFromAnimeRevisions(ctx, revisions, time.UTC, nil, DBRevisionIDNames{})

	update := got[0]
	if update.EditorKind != DBRevisionEditorUser || update.EditorUsername != "editor" || update.ActionLabel != "更新" || update.IsCreate {
		t.Errorf("更新の版 = %+v、期待値 = editorの更新", update)
	}
	if update.CreatedAtDisplay != "2026-11-08 15:30" || update.CreatedAtDateTime != "2026-11-08T15:30:00Z" {
		t.Errorf("日時 = %q・%q、期待値 = 2026-11-08 15:30・2026-11-08T15:30:00Z", update.CreatedAtDisplay, update.CreatedAtDateTime)
	}
	if len(update.Fields) != 1 || update.Fields[0] != (DBRevisionField{Label: "タイトル", Before: "前", After: "後"}) {
		t.Errorf("更新の版の項目 = %+v、タイトルの変更だけを期待", update.Fields)
	}

	create := got[1]
	if create.EditorKind != DBRevisionEditorDeletedUser || !create.IsCreate {
		t.Errorf("作成の版 = %+v、期待値 = 退会したユーザーの作成", create)
	}
	if len(create.Rows) != 1 || !create.Rows[0].Added || create.Rows[0].TableLabel != "リリース時期" || create.Rows[0].Values != "年: 2026 / 主なリリース時期: ○" {
		t.Errorf("作成の版の行 = %+v、リリース時期の追加を期待", create.Rows)
	}

	if got[2].EditorKind != DBRevisionEditorSystem {
		t.Errorf("システムの版の編集者 = %q、期待値 = system", got[2].EditorKind)
	}
}

// 差分の項目名が、作品・エピソードのフォームのラベルと同じ呼び方であることを確かめる。
func TestDBRevisionFieldLabel(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		field model.RevisionField
		ja    string
		en    string
	}{
		{field: "title_alter", ja: "タイトル (別名)", en: "Title (Alternate)"},
		{field: "title_alter_ro", ja: "タイトル (ローマ字別名)", en: "Title (Romaji Alternate)"},
		{field: "title_alter_en", ja: "タイトル (英語別名)", en: "Title (English Alternate)"},
		{field: "title_alter_other", ja: "タイトル (その他の別名)", en: "Title (Other Alternate)"},
		{field: "synopsis_source", ja: "あらすじの引用元", en: "Synopsis Source"},
		{field: "synopsis_source_en", ja: "あらすじの引用元 (英語)", en: "Synopsis Source (English)"},
		{field: "number", ja: "話数", en: "Number"},
		{field: "number_text", ja: "表示用話数", en: "Display number"},
		{field: "sort_number", ja: "ソート番号", en: "Sort number"},
		{field: "number_format_id", ja: "話数フォーマット", en: "Episode Number Format"},
		{field: "episode_start_number", ja: "開始話数", en: "Start Episode Number"},
		{field: "expected_episodes_count", ja: "予定エピソード数", en: "Expected Episodes"},
	} {
		t.Run(string(tt.field), func(t *testing.T) {
			t.Parallel()

			if got := dbRevisionFieldLabel(i18n.SetLocale(context.Background(), "ja"), tt.field); got != tt.ja {
				t.Errorf("日本語の項目名 = %q、期待値 = %q", got, tt.ja)
			}
			if got := dbRevisionFieldLabel(i18n.SetLocale(context.Background(), "en"), tt.field); got != tt.en {
				t.Errorf("英語の項目名 = %q、期待値 = %q", got, tt.en)
			}
		})
	}
}

// エピソードの版の差分の表には分類の項目を出さず、作品の版には種類 (kind) だけを出さない。
func TestDBAnimeRevisionHiddenFields(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	parent := model.AnimeID(1)
	otherParent := model.AnimeID(2)
	before, after := "前", "後"
	edit := &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionUpdate, CreatedAt: time.Now()}
	createEdit := &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionCreate, CreatedAt: time.Now()}

	labels := func(fields []DBRevisionField) []string {
		got := make([]string, len(fields))
		for i, f := range fields {
			got[i] = f.Label
		}
		return got
	}

	for _, tt := range []struct {
		name     string
		revision *model.AnimeRevision
		want     []string
	}{
		{
			name: "エピソードの更新の版は分類の項目を除く",
			revision: &model.AnimeRevision{
				DBEdit:   edit,
				Previous: &model.AnimeSnapshot{Title: &before, Kind: model.AnimeClassificationKindWork, ParentAnimeID: &parent, Standalone: true},
				Snapshot: model.AnimeSnapshot{Title: &after, Kind: model.AnimeClassificationKindEpisode, ParentAnimeID: &otherParent},
			},
			want: []string{"タイトル"},
		},
		{
			name: "エピソードの作成の版は分類の項目を除く",
			revision: &model.AnimeRevision{
				DBEdit:   createEdit,
				Snapshot: model.AnimeSnapshot{Title: &after, Kind: model.AnimeClassificationKindEpisode, ParentAnimeID: &parent},
			},
			want: []string{"タイトル"},
		},
		{
			name: "作品の更新の版は種類だけを除く",
			revision: &model.AnimeRevision{
				DBEdit:   edit,
				Previous: &model.AnimeSnapshot{Title: &before, Kind: model.AnimeClassificationKindEpisode, ParentAnimeID: &parent, Standalone: true},
				Snapshot: model.AnimeSnapshot{Title: &after, Kind: model.AnimeClassificationKindWork, ParentAnimeID: &otherParent},
			},
			want: []string{"タイトル", "親の作品", "エピソードなし"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			revisions := NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{tt.revision}, time.UTC, nil, DBRevisionIDNames{})
			if got := labels(revisions[0].Fields); !slices.Equal(got, tt.want) {
				t.Errorf("各リソースの履歴の項目 = %v、期待値 = %v", got, tt.want)
			}

			// 全体の編集履歴は操作の下に版を持つ。
			revision := *tt.revision
			revision.DBEdit = nil
			dbEdit := *tt.revision.DBEdit
			dbEdit.AnimeRevisions = []*model.AnimeRevision{&revision}
			edits := NewDBEdits(ctx, []*model.DBEdit{&dbEdit}, time.UTC, nil, DBRevisionIDNames{})
			if got := labels(edits[0].Revisions[0].Revision.Fields); !slices.Equal(got, tt.want) {
				t.Errorf("全体の編集履歴の項目 = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}

func TestDBRevisionLocation(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		viewer *model.User
		want   string
	}{
		{name: "未ログイン", viewer: nil, want: "Asia/Tokyo"},
		{name: "タイムゾーンを持つ閲覧者", viewer: &model.User{TimeZone: "America/New_York"}, want: "America/New_York"},
		{name: "読めないタイムゾーン", viewer: &model.User{TimeZone: "Not/AZone"}, want: "Asia/Tokyo"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := DBRevisionLocation(tt.viewer).String(); got != tt.want {
				t.Errorf("DBRevisionLocation() = %q、期待値 = %q", got, tt.want)
			}
		})
	}
}

func TestDBWorkRevisionWork(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, title, locale, heading string }{
		{name: "作品名の空白を除く", title: "  作品  ", locale: "ja", heading: "作品"},
		{name: "空の作品名は日本語のページ名", locale: "ja", heading: "作品の履歴"},
		{name: "空白だけの作品名は英語のページ名", title: " \t", locale: "en", heading: "Work History"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			work := NewDBWorkRevisionWork(&model.Work{ID: 42, Title: tt.title, NoEpisodes: true})
			ctx := i18n.SetLocale(context.Background(), tt.locale)
			if work.ID != WorkID(42) || !work.NoEpisodes {
				t.Errorf("サブナビの作品データ = %+v", work)
			}
			if got := work.Heading(ctx); got != tt.heading {
				t.Errorf("見出し = %q、期待値 = %q", got, tt.heading)
			}
		})
	}
}

func TestNewDBRevisionsFromAnimeRevisions_Episode(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	number, title := "第3話", "タイトル"
	edit := &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionCreate, CreatedAt: time.Now()}
	revisions := []*model.AnimeRevision{
		{DBEdit: edit, Snapshot: model.AnimeSnapshot{NumberText: &number, Title: &title}, Episode: &model.Episode{ID: 7}},
		{DBEdit: edit, Snapshot: model.AnimeSnapshot{Title: &title}},
	}

	got := NewDBRevisionsFromAnimeRevisions(ctx, revisions, time.UTC, nil, DBRevisionIDNames{})
	if got[0].Episode == nil || got[0].Episode.ID != 7 || got[0].Episode.Label != "第3話 タイトル" {
		t.Errorf("エピソードの版のEpisode = %+v、期待値 = ID 7・第3話 タイトル", got[0].Episode)
	}
	if got[1].Episode != nil {
		t.Errorf("作品の版のEpisode = %+v、nilを期待", got[1].Episode)
	}
}

// 作品の履歴に並べる作品の版だけが版の時点の作品名を持ち、エピソードの版とエピソードの履歴の版は持たないことを確かめる。
func TestNewDBRevisionsFromAnimeRevisions_WorkName(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	title, blank := " 版の時点の作品 ", " "
	edit := &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate, CreatedAt: time.Now()}
	for _, tt := range []struct {
		name     string
		revision *model.AnimeRevision
		want     string
	}{
		{
			name:     "作品の版",
			revision: &model.AnimeRevision{DBEdit: edit, Snapshot: model.AnimeSnapshot{Kind: model.AnimeClassificationKindWork, Title: &title}},
			want:     "版の時点の作品",
		},
		{
			name:     "タイトルが空の作品の版",
			revision: &model.AnimeRevision{DBEdit: edit, Snapshot: model.AnimeSnapshot{Kind: model.AnimeClassificationKindWork, Title: &blank}},
			want:     "名前の無い作品",
		},
		{
			name:     "作品の履歴のエピソードの版",
			revision: &model.AnimeRevision{DBEdit: edit, Snapshot: model.AnimeSnapshot{Kind: model.AnimeClassificationKindEpisode, Title: &title}, Episode: &model.Episode{ID: 7}},
		},
		{
			name:     "エピソードの履歴の版",
			revision: &model.AnimeRevision{DBEdit: edit, Snapshot: model.AnimeSnapshot{Kind: model.AnimeClassificationKindEpisode, Title: &title}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{tt.revision}, time.UTC, nil, DBRevisionIDNames{})[0]
			if got.WorkName != tt.want {
				t.Errorf("WorkName = %q、期待値 = %q", got.WorkName, tt.want)
			}
		})
	}
}

func TestDBEpisodeRevisionLabel(t *testing.T) {
	t.Parallel()

	str := func(v string) *string { return &v }
	for _, tt := range []struct {
		name   string
		number *string
		title  *string
		want   string
	}{
		{name: "話数とタイトル", number: str("第1話"), title: str(" はじまり "), want: "第1話 はじまり"},
		{name: "タイトルだけ", number: str(""), title: str("総集編"), want: "総集編"},
		{name: "どちらも無い", want: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DBEpisodeRevisionLabel(tt.number, tt.title); got != tt.want {
				t.Errorf("DBEpisodeRevisionLabel() = %q、期待値 = %q", got, tt.want)
			}
		})
	}
}

// TestDBEpisodeDisplayNameは、作品の履歴とエピソードの履歴が同じエピソードを同じ名前で呼ぶことを検証する。
func TestDBEpisodeDisplayName(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		locale string
		label  string
		want   string
	}{
		{locale: "ja", label: "第1話 はじまり", want: "第1話 はじまり"},
		{locale: "ja", label: "", want: "名前の無いエピソード"},
		{locale: "en", label: "", want: "Untitled episode"},
	} {
		t.Run(tt.locale+"/"+tt.want, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			if got := (DBRevisionEpisode{Label: tt.label}).DisplayName(ctx); got != tt.want {
				t.Errorf("DBRevisionEpisode.DisplayName() = %q、期待値 = %q", got, tt.want)
			}
			if got := (DBEpisodeRevisionEpisode{Label: tt.label}).DisplayName(ctx); got != tt.want {
				t.Errorf("DBEpisodeRevisionEpisode.DisplayName() = %q、期待値 = %q", got, tt.want)
			}
		})
	}
}

func TestNewDBRevisionsFromPersonRevisions(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	edit := &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate, CreatedAt: time.Now()}
	revisions := []*model.PersonRevision{{
		DBEdit:   edit,
		Snapshot: model.PersonSnapshot{Name: "新しい名前", Status: model.PersonStatusPublished},
		Previous: &model.PersonSnapshot{Name: "古い名前", Status: model.PersonStatusPublished},
	}}

	got := NewDBRevisionsFromPersonRevisions(ctx, revisions, time.UTC, nil, DBRevisionIDNames{})
	if len(got) != 1 || got[0].IsCreate || len(got[0].Fields) != 1 {
		t.Fatalf("版 = %+v、期待値 = 名前の変更だけを持つ更新の版", got)
	}
	if f := got[0].Fields[0]; f.Label != "名前" || f.Before != "古い名前" || f.After != "新しい名前" {
		t.Errorf("変わった項目 = %+v、期待値 = 名前: 古い名前 → 新しい名前", f)
	}
}

// 「状態」の値を、保存した値ではなく人物一覧の状態のバッジと同じ言葉で示すことを確かめる。言葉を持たない値は
// 保存した値のまま示す。
func TestNewDBRevisions_StatusValue(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name          string
		locale        string
		before, after string
		wantBefore    string
		wantAfter     string
	}{
		{name: "日本語の非公開", locale: "ja", before: "published", after: "archived", wantBefore: "公開", wantAfter: "非公開"},
		{name: "日本語の掲載停止", locale: "ja", before: "published", after: "delisted", wantBefore: "公開", wantAfter: "掲載停止"},
		{name: "英語の非公開", locale: "en", before: "published", after: "archived", wantBefore: "Published", wantAfter: "Archived"},
		{name: "英語の掲載停止", locale: "en", before: "archived", after: "delisted", wantBefore: "Archived", wantAfter: "Delisted"},
		{name: "言葉を持たない値", locale: "ja", before: "published", after: "merged", wantBefore: "公開", wantAfter: "merged"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := i18n.SetLocale(context.Background(), tt.locale)
			edit := &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate}
			anime := NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{{
				DBEdit:   edit,
				Previous: &model.AnimeSnapshot{Status: model.AnimeStatus(tt.before)},
				Snapshot: model.AnimeSnapshot{Status: model.AnimeStatus(tt.after)},
			}}, time.UTC, nil, DBRevisionIDNames{})
			person := NewDBRevisionsFromPersonRevisions(ctx, []*model.PersonRevision{{
				DBEdit:   edit,
				Previous: &model.PersonSnapshot{Status: model.PersonStatus(tt.before)},
				Snapshot: model.PersonSnapshot{Status: model.PersonStatus(tt.after)},
			}}, time.UTC, nil, DBRevisionIDNames{})
			for _, r := range []struct {
				name     string
				revision DBRevision
			}{{"作品の版", anime[0]}, {"人物の版", person[0]}} {
				if len(r.revision.Fields) != 1 {
					t.Fatalf("%sの変わった項目 = %+v、期待値 = 状態だけ", r.name, r.revision.Fields)
				}
				if f := r.revision.Fields[0]; f.Before != tt.wantBefore || f.After != tt.wantAfter {
					t.Errorf("%sの状態 = %q → %q、期待値 = %q → %q", r.name, f.Before, f.After, tt.wantBefore, tt.wantAfter)
				}
			}
		})
	}
}

// 作品・人物の版の選択式の項目と、作品の別表の行の選択式の値を、保存した値ではなく言葉で示すことを確かめる。
// 言葉を持たない値は保存した値のまま示す。
func TestNewDBRevisions_ChoiceValue(t *testing.T) {
	t.Parallel()

	media := func(m model.AnimeMedia) *model.AnimeMedia { return &m }
	releaseStatus := func(s model.ReleaseStatus) *model.ReleaseStatus { return &s }
	gender := func(g model.PersonGender) *model.PersonGender { return &g }
	bloodType := func(b model.PersonBloodType) *model.PersonBloodType { return &b }
	season := func(s model.SeasonName) *model.SeasonName { return &s }
	edit := &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate}
	startedOn := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	animeBefore := &model.AnimeSnapshot{
		Kind:          model.AnimeClassificationKindWork,
		Media:         media(model.AnimeMediaTV),
		ReleaseStatus: releaseStatus(model.ReleaseStatusNotYetReleased),
		Standalone:    false,
	}
	animeAfter := model.AnimeSnapshot{
		Kind:          model.AnimeClassificationKindWork,
		Media:         media(model.AnimeMediaMovie),
		ReleaseStatus: releaseStatus(model.ReleaseStatusReleased),
		Standalone:    true,
		ExternalIDs:   []model.AnimeSnapshotExternalID{{Service: model.AnimeExternalServiceSyobocal, ExternalID: "123"}},
		Links:         []model.AnimeSnapshotLink{{Kind: model.AnimeLinkKindOfficialSite, Language: model.LanguageJa, URL: "https://example.com"}},
		OfficialAccounts: []model.AnimeSnapshotOfficialAccount{
			{Service: model.AnimeAccountServiceX, Account: "annict"},
			{Service: model.AnimeAccountService("unknown"), Account: "annict"},
		},
		Seasons: []model.AnimeSnapshotSeason{{Year: 2026, Name: season(model.SeasonNameFall), IsPrimary: true}},
		Events:  []model.AnimeSnapshotEvent{{Kind: model.AnimeEventKindRevivalScreening, StartedOn: startedOn}},
	}
	personBefore := &model.PersonSnapshot{Name: "人物", Gender: gender(model.PersonGenderMale), BloodType: bloodType(model.PersonBloodTypeA)}
	personAfter := model.PersonSnapshot{Name: "人物", Gender: gender(model.PersonGenderFemale), BloodType: bloodType(model.PersonBloodTypeAB)}

	for _, tt := range []struct {
		locale     string
		wantFields []DBRevisionField
		wantRows   []string
		wantPerson []DBRevisionField
	}{
		{
			locale: "ja",
			wantFields: []DBRevisionField{
				{Label: "メディア", Before: "TV", After: "映画"},
				{Label: "放送・公開の状況", Before: "放送・公開前", After: "放送・公開済み"},
				{Label: "エピソードなし", Before: "×", After: "○"},
			},
			wantRows: []string{
				"サービス: しょぼいカレンダー / 外部ID: 123",
				"種類: 公式サイト / 言語: 日本語 / URL: https://example.com / ソート番号: 0",
				"サービス: X / アカウント: annict / ソート番号: 0",
				"サービス: unknown / アカウント: annict / ソート番号: 0",
				"年: 2026 / 季節: 秋 / 主なリリース時期: ○",
				"種類: リバイバル上映 / 開始日: 2026-04-01 / ソート番号: 0",
			},
			wantPerson: []DBRevisionField{
				{Label: "性別", Before: "男性", After: "女性"},
				{Label: "血液型", Before: "A型", After: "AB型"},
			},
		},
		{
			locale: "en",
			wantFields: []DBRevisionField{
				{Label: "Media", Before: "TV", After: "Movie"},
				{Label: "Release status", Before: "Not yet released", After: "Released"},
				{Label: "No episodes", Before: "No", After: "Yes"},
			},
			wantRows: []string{
				"Service: Syoboi Calendar / External ID: 123",
				"Kind: Official site / Language: Japanese / URL: https://example.com / Sort number: 0",
				"Service: X / Account: annict / Sort number: 0",
				"Service: unknown / Account: annict / Sort number: 0",
				"Year: 2026 / Season: Fall / Primary: Yes",
				"Kind: Revival screening / Start Date: 2026-04-01 / Sort number: 0",
			},
			wantPerson: []DBRevisionField{
				{Label: "Gender", Before: "Male", After: "Female"},
				{Label: "Blood Type", Before: "A", After: "AB"},
			},
		},
	} {
		t.Run(tt.locale, func(t *testing.T) {
			t.Parallel()
			ctx := i18n.SetLocale(context.Background(), tt.locale)

			anime := NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{{DBEdit: edit, Previous: animeBefore, Snapshot: animeAfter}}, time.UTC, nil, DBRevisionIDNames{})[0]
			if !slices.Equal(anime.Fields, tt.wantFields) {
				t.Errorf("作品の版の項目 = %+v、期待値 = %+v", anime.Fields, tt.wantFields)
			}
			rows := make([]string, len(anime.Rows))
			for i, r := range anime.Rows {
				rows[i] = r.Values
			}
			if !slices.Equal(rows, tt.wantRows) {
				t.Errorf("作品の版の行 = %q、期待値 = %q", rows, tt.wantRows)
			}

			person := NewDBRevisionsFromPersonRevisions(ctx, []*model.PersonRevision{{DBEdit: edit, Previous: personBefore, Snapshot: personAfter}}, time.UTC, nil, DBRevisionIDNames{})[0]
			if !slices.Equal(person.Fields, tt.wantPerson) {
				t.Errorf("人物の版の項目 = %+v、期待値 = %+v", person.Fields, tt.wantPerson)
			}
		})
	}
}

// 作品の版の「話数フォーマット」と人物の版の「出身地」を、IDではなく名前で示すことを確かめる。一覧に無いIDと、
// 対応表を渡さないときは、保存した値のまま示す。
func TestNewDBRevisions_IDNames(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	edit := &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate}
	numberFormat := func(id int64) *model.NumberFormatID { v := model.NumberFormatID(id); return &v }
	prefecture := func(id int64) *model.PrefectureID { v := model.PrefectureID(id); return &v }
	names := NewDBRevisionIDNames(
		[]model.NumberFormat{{ID: 1, Name: "第1話"}, {ID: 2, Name: "#1"}},
		[]model.Prefecture{{ID: 1, Name: "東京都"}, {ID: 2, Name: "大阪府"}},
	)

	for _, tt := range []struct {
		name       string
		names      DBRevisionIDNames
		before     int64
		after      int64
		wantWork   DBRevisionField
		wantPerson DBRevisionField
	}{
		{
			name: "一覧にあるIDは名前で示す", names: names, before: 1, after: 2,
			wantWork:   DBRevisionField{Label: "話数フォーマット", Before: "第1話", After: "#1"},
			wantPerson: DBRevisionField{Label: "出身地", Before: "東京都", After: "大阪府"},
		},
		{
			name: "一覧に無いIDは保存した値で示す", names: names, before: 1, after: 99,
			wantWork:   DBRevisionField{Label: "話数フォーマット", Before: "第1話", After: "99"},
			wantPerson: DBRevisionField{Label: "出身地", Before: "東京都", After: "99"},
		},
		{
			name: "対応表を渡さないときは保存した値で示す", names: DBRevisionIDNames{}, before: 1, after: 2,
			wantWork:   DBRevisionField{Label: "話数フォーマット", Before: "1", After: "2"},
			wantPerson: DBRevisionField{Label: "出身地", Before: "1", After: "2"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			animeRevision := &model.AnimeRevision{
				Previous: &model.AnimeSnapshot{Kind: model.AnimeClassificationKindWork, NumberFormatID: numberFormat(tt.before)},
				Snapshot: model.AnimeSnapshot{Kind: model.AnimeClassificationKindWork, NumberFormatID: numberFormat(tt.after)},
			}
			personRevision := &model.PersonRevision{
				Previous: &model.PersonSnapshot{PrefectureID: prefecture(tt.before)},
				Snapshot: model.PersonSnapshot{PrefectureID: prefecture(tt.after)},
			}

			animeRevision.DBEdit = edit
			personRevision.DBEdit = edit
			work := NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{animeRevision}, time.UTC, nil, tt.names)[0]
			person := NewDBRevisionsFromPersonRevisions(ctx, []*model.PersonRevision{personRevision}, time.UTC, nil, tt.names)[0]
			if !slices.Equal(work.Fields, []DBRevisionField{tt.wantWork}) {
				t.Errorf("作品の履歴の項目 = %+v、期待値 = %+v", work.Fields, tt.wantWork)
			}
			if !slices.Equal(person.Fields, []DBRevisionField{tt.wantPerson}) {
				t.Errorf("人物の履歴の項目 = %+v、期待値 = %+v", person.Fields, tt.wantPerson)
			}

			// 全体の編集履歴は操作の下に版を持つ。
			animeRevision.DBEdit, personRevision.DBEdit = nil, nil
			dbEdit := *edit
			dbEdit.AnimeRevisions = []*model.AnimeRevision{animeRevision}
			dbEdit.PersonRevisions = []*model.PersonRevision{personRevision}
			edits := NewDBEdits(ctx, []*model.DBEdit{&dbEdit}, time.UTC, nil, tt.names)
			if got := edits[0].Revisions[0].Revision.Fields; !slices.Equal(got, []DBRevisionField{tt.wantWork}) {
				t.Errorf("全体の編集履歴の作品の項目 = %+v、期待値 = %+v", got, tt.wantWork)
			}
			if got := edits[0].Revisions[1].Revision.Fields; !slices.Equal(got, []DBRevisionField{tt.wantPerson}) {
				t.Errorf("全体の編集履歴の人物の項目 = %+v、期待値 = %+v", got, tt.wantPerson)
			}
		})
	}
}

func TestDBPersonRevisionPerson_Heading(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	if got := NewDBPersonRevisionPerson(&model.Person{ID: 1, Name: " 人物 "}).Heading(ctx); got != "人物" {
		t.Errorf("Heading() = %q、期待値 = 人物", got)
	}
	if got := NewDBPersonRevisionPerson(&model.Person{ID: 1}).Heading(ctx); got != "人物の履歴" {
		t.Errorf("名前の無い人物のHeading() = %q、期待値 = 人物の履歴", got)
	}
}

func TestNewDBEdits(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	title, number := "作品名", "第1話"
	edit := &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionCreate, CreatedAt: time.Date(2026, 11, 8, 3, 4, 0, 0, time.UTC)}
	edit.AnimeRevisions = []*model.AnimeRevision{
		{Snapshot: model.AnimeSnapshot{Title: &title}, Work: &model.Work{ID: 1}},
		// エピソードの版は、どの作品のエピソードかが分かるよう今の親作品の名前を添える。
		{Snapshot: model.AnimeSnapshot{NumberText: &number}, Episode: &model.Episode{ID: 2}, Work: &model.Work{ID: 1, Title: " 今の親作品 "}},
		{Snapshot: model.AnimeSnapshot{}, Work: &model.Work{ID: 3}},
		{Snapshot: model.AnimeSnapshot{}, Episode: &model.Episode{ID: 5}, Work: &model.Work{ID: 3}},
	}
	edit.PersonRevisions = []*model.PersonRevision{{PersonID: 4, Snapshot: model.PersonSnapshot{Name: " 人物 "}}}

	got := NewDBEdits(ctx, []*model.DBEdit{edit}, time.UTC, nil, DBRevisionIDNames{})
	if len(got) != 1 || got[0].EditorKind != DBRevisionEditorSystem || got[0].ActionLabel != "作成" || got[0].CreatedAtDisplay != "2026-11-08 03:04" || got[0].CreatedAtDateTime != "2026-11-08T03:04:00Z" {
		t.Fatalf("操作 = %+v、期待値 = システムの作成の操作", got)
	}
	want := []struct {
		kind DBEditTargetKind
		name string
	}{
		{DBEditTargetWork, "作品名"},
		{DBEditTargetEpisode, "第1話 (今の親作品)"},
		{DBEditTargetWork, "名前の無い作品"},
		{DBEditTargetEpisode, "名前の無いエピソード (名前の無い作品)"},
		{DBEditTargetPerson, "人物"},
	}
	if len(got[0].Revisions) != len(want) {
		t.Fatalf("版 = %d件、期待値 = %d件", len(got[0].Revisions), len(want))
	}
	for i, w := range want {
		r := got[0].Revisions[i]
		if r.Kind != w.kind || r.Name != w.name || r.Revision.TargetName != w.name || !r.Revision.IsCreate {
			t.Errorf("%d件目の版 = %+v、期待値 = %s・%q", i+1, r, w.kind, w.name)
		}
	}
	if got[0].Revisions[1].EpisodeID != 2 || got[0].Revisions[3].EpisodeID != 5 || got[0].Revisions[4].PersonID != 4 {
		t.Errorf("版が指すリソース = %+v", got[0].Revisions)
	}
}

// プロフィール画像を持つユーザーの編集者にだけアバターのURLを作り、画像の無い編集者には頭文字を返すことを確かめる。
// 全体の編集履歴では、操作に属する版も操作と同じ編集者を持つ。
func TestNewDBEditHeader_EditorAvatar(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	helper := image.NewHelper(&config.Config{
		ImgproxyEndpoint: "http://localhost:18080",
		ImgproxyKey:      "test-key",
		ImgproxySalt:     "test-salt",
		S3BucketName:     "test-bucket",
	})
	imageData := sql.NullString{String: `{"original":{"id":"avatar.jpg"}}`, Valid: true}

	for _, tt := range []struct {
		name        string
		edit        *model.DBEdit
		wantURL     bool
		wantInitial string
	}{
		{
			name:        "プロフィール画像を持つユーザー",
			edit:        &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionUpdate, User: &model.User{ID: 1, Username: "editor", ProfileImageData: imageData}},
			wantURL:     true,
			wantInitial: "e",
		},
		{
			name:        "プロフィール画像の無いユーザー",
			edit:        &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionUpdate, User: &model.User{ID: 1, Username: "editor"}},
			wantInitial: "e",
		},
		{
			name: "システム",
			edit: &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate},
		},
		{
			name: "退会したユーザー",
			edit: &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionUpdate},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.edit.PersonRevisions = []*model.PersonRevision{{PersonID: 1}}
			revisions := NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{{DBEdit: tt.edit}}, time.UTC, helper, DBRevisionIDNames{})
			edits := NewDBEdits(ctx, []*model.DBEdit{tt.edit}, time.UTC, helper, DBRevisionIDNames{})
			for _, h := range []struct {
				name   string
				header DBEditHeader
			}{
				{"各リソースの履歴の版", revisions[0].DBEditHeader},
				{"全体の編集履歴の操作", edits[0].DBEditHeader},
				{"全体の編集履歴の操作に属する版", edits[0].Revisions[0].Revision.DBEditHeader},
			} {
				if got := h.header.EditorAvatarURL != ""; got != tt.wantURL {
					t.Errorf("%sのEditorAvatarURL = %q、URLの有無の期待値 = %v", h.name, h.header.EditorAvatarURL, tt.wantURL)
				}
				if tt.wantURL && !strings.HasPrefix(h.header.EditorAvatarURL, "http://localhost:18080/") {
					t.Errorf("%sのEditorAvatarURL = %q、imgproxyのURLを期待", h.name, h.header.EditorAvatarURL)
				}
				if got := h.header.EditorInitial(); got != tt.wantInitial {
					t.Errorf("%sのEditorInitial() = %q、期待値 = %q", h.name, got, tt.wantInitial)
				}
			}
		})
	}
}

// 7つの操作のすべてに、日本語と英語で編集者を含む文があり、編集者の位置で前後に分けられることを確かめる。
func TestNewDBEditHeader_Sentence(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		locale        string
		action        model.DBEditAction
		before, after string
	}{
		{"ja", model.DBEditActionCreate, "", "が作成しました"},
		{"ja", model.DBEditActionUpdate, "", "が更新しました"},
		{"ja", model.DBEditActionArchive, "", "が非公開にしました"},
		{"ja", model.DBEditActionUnarchive, "", "が再公開しました"},
		{"ja", model.DBEditActionDelete, "", "が削除しました"},
		{"ja", model.DBEditActionDelist, "", "が掲載を停止しました"},
		{"ja", model.DBEditActionUndelist, "", "が掲載を再開しました"},
		{"en", model.DBEditActionCreate, "", " created"},
		{"en", model.DBEditActionUpdate, "", " updated"},
		{"en", model.DBEditActionArchive, "", " archived"},
		{"en", model.DBEditActionUnarchive, "", " unarchived"},
		{"en", model.DBEditActionDelete, "", " deleted"},
		{"en", model.DBEditActionDelist, "", " delisted"},
		{"en", model.DBEditActionUndelist, "", " undelisted"},
	} {
		t.Run(tt.locale+"/"+string(tt.action), func(t *testing.T) {
			t.Parallel()
			ctx := i18n.SetLocale(context.Background(), tt.locale)
			h := newDBEditHeader(ctx, &model.DBEdit{Actor: model.DBEditActorSystem, Action: tt.action}, time.UTC, nil)
			if h.SentenceBeforeEditor != tt.before || h.SentenceAfterEditor != tt.after {
				t.Errorf("文の前後 = (%q, %q)、期待値 = (%q, %q)", h.SentenceBeforeEditor, h.SentenceAfterEditor, tt.before, tt.after)
			}
		})
	}
}
