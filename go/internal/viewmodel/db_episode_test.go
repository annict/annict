package viewmodel

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// TestDBEpisodeListWorkName verifies that the work name is trimmed and that a title holding
// nothing but whitespace collapses to the empty string, which the heading and the document
// title read as "the work has no display name".
//
// [Ja] TestDBEpisodeListWorkName は作品の名前が前後の空白を落として返ること、および空白だけの
// タイトルが空文字列に畳まれることを検証する。見出しと文書タイトルはいずれもこれを「作品に
// 表示名が無い」合図として読む。
func TestDBEpisodeListWorkName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		workTitle string
		want      string
	}{
		{name: "通常のタイトル", workTitle: "テストアニメ", want: "テストアニメ"},
		{name: "前後の空白は落とす", workTitle: "  テストアニメ  ", want: "テストアニメ"},
		{name: "空白のみは空文字列", workTitle: "   ", want: ""},
		{name: "空文字列は空文字列", workTitle: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := DBEpisodeListWorkName(tt.workTitle); got != tt.want {
				t.Errorf("DBEpisodeListWorkName(%q) = %q, want %q", tt.workTitle, got, tt.want)
			}
		})
	}
}

// TestDBEpisodeIdentifier verifies the label that opens the document title of a page naming one
// episode. Both locales are covered because the two formats live in the translation files: a key
// that went missing from one of them would come back as its own message id rather than as a
// label.
//
// [Ja] TestDBEpisodeIdentifier は 1 件のエピソードを名指しするページの文書タイトル先頭に来る
// ラベルを検証する。2 つの書式は翻訳ファイルにあるため両ロケールを対象にする。片方から翻訳キーが
// 抜けた場合、ラベルではなくメッセージ ID がそのまま返るため。
func TestDBEpisodeIdentifier(t *testing.T) {
	t.Parallel()

	number := "  第2話  "
	tests := []struct {
		name    string
		locale  string
		episode *model.Episode
		want    string
	}{
		{
			name:    "日本語: 表示用話数と ID を含める",
			locale:  "ja",
			episode: &model.Episode{ID: 123, Number: &number},
			want:    "第2話 (ID: 123)",
		},
		{
			name:    "日本語: 表示用話数が無ければ ID を使う",
			locale:  "ja",
			episode: &model.Episode{ID: 456},
			want:    "ID: 456",
		},
		{
			name:    "英語: 表示用話数と ID を含める",
			locale:  "en",
			episode: &model.Episode{ID: 123, Number: &number},
			want:    "第2話 (ID: 123)",
		},
		{
			name:    "英語: 表示用話数が無ければ ID を使う",
			locale:  "en",
			episode: &model.Episode{ID: 456},
			want:    "ID: 456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			if got := DBEpisodeIdentifier(ctx, tt.episode); got != tt.want {
				t.Errorf("DBEpisodeIdentifier() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNewDBEpisodeListItem verifies the projection of an episode onto its display row:
// the ids and the parent work id the row's link needs, the two number representations,
// both titles, and the unset attributes left empty for the template to render as gaps.
//
// [Ja] TestNewDBEpisodeListItem はエピソードから表示行への射影を検証する。行のリンクが必要と
// する各 ID と親作品 ID、2 系統の話数、両方のタイトル、およびテンプレートが欠落として描画
// できるよう空のまま残す未設定の属性を確認する。
func TestNewDBEpisodeListItem(t *testing.T) {
	t.Parallel()

	number := "第2話"
	rawNumber := 2.5
	title := "エピソードタイトル"

	tests := []struct {
		name          string
		episode       *model.Episode
		wantNumber    string
		wantRawNumber string
		wantTitle     string
		wantTitleEn   string
	}{
		{
			name: "全項目あり",
			episode: &model.Episode{
				ID:                  10,
				WorkID:              3,
				Number:              &number,
				RawNumber:           &rawNumber,
				Title:               &title,
				TitleEn:             "Episode Title",
				SortNumber:          200,
				EpisodeRecordsCount: 42,
			},
			wantNumber:    "第2話",
			wantRawNumber: "2.5",
			wantTitle:     "エピソードタイトル",
			wantTitleEn:   "Episode Title",
		},
		{
			name: "未設定の属性は空文字列のまま",
			episode: &model.Episode{
				ID:         11,
				WorkID:     3,
				SortNumber: 100,
			},
			wantNumber:    "",
			wantRawNumber: "",
			wantTitle:     "",
			wantTitleEn:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBEpisodeListItem(tt.episode)

			if got.ID != EpisodeID(tt.episode.ID) {
				t.Errorf("ID = %q, want %q", got.ID, EpisodeID(tt.episode.ID))
			}
			if got.WorkID != WorkID(tt.episode.WorkID) {
				t.Errorf("WorkID = %q, want %q", got.WorkID, WorkID(tt.episode.WorkID))
			}
			if got.Number != tt.wantNumber {
				t.Errorf("Number = %q, want %q", got.Number, tt.wantNumber)
			}
			if got.RawNumber != tt.wantRawNumber {
				t.Errorf("RawNumber = %q, want %q", got.RawNumber, tt.wantRawNumber)
			}
			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
			}
			if got.TitleEn != tt.wantTitleEn {
				t.Errorf("TitleEn = %q, want %q", got.TitleEn, tt.wantTitleEn)
			}
			if got.SortNumber != tt.episode.SortNumber {
				t.Errorf("SortNumber = %d, want %d", got.SortNumber, tt.episode.SortNumber)
			}
			if got.EpisodeRecordsCount != tt.episode.EpisodeRecordsCount {
				t.Errorf("EpisodeRecordsCount = %d, want %d", got.EpisodeRecordsCount, tt.episode.EpisodeRecordsCount)
			}
		})
	}
}

// TestNewDBEpisodeListItem_PrevNumber verifies how the preceding episode is named: its
// display number when it has one, its numeric number as a fallback, and the empty string
// when there is no preceding episode at all, which the template renders as a gap.
//
// [Ja] TestNewDBEpisodeListItem_PrevNumber は直前のエピソードの名指し方を検証する。表示用
// 話数があればそれを、無ければ数値話数を使い、直前のエピソード自体が無ければ空文字列とする
// (テンプレートはこれを欠落として描画する)。
func TestNewDBEpisodeListItem_PrevNumber(t *testing.T) {
	t.Parallel()

	prevNumber := "第1話"
	emptyPrevNumber := ""
	prevRawNumber := 1.5

	tests := []struct {
		name          string
		prevNumber    *string
		prevRawNumber *float64
		want          string
	}{
		{
			name:          "表示用話数があればそれを使う",
			prevNumber:    &prevNumber,
			prevRawNumber: &prevRawNumber,
			want:          "第1話",
		},
		{
			name:          "表示用話数が無ければ数値話数にフォールバックする",
			prevRawNumber: &prevRawNumber,
			want:          "1.5",
		},
		{
			name:          "表示用話数が空文字列でも数値話数にフォールバックする",
			prevNumber:    &emptyPrevNumber,
			prevRawNumber: &prevRawNumber,
			want:          "1.5",
		},
		{
			name: "直前のエピソードが無ければ空文字列",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBEpisodeListItem(&model.Episode{
				ID:            1,
				WorkID:        1,
				PrevNumber:    tt.prevNumber,
				PrevRawNumber: tt.prevRawNumber,
			})

			if got.PrevNumber != tt.want {
				t.Errorf("PrevNumber = %q, want %q", got.PrevNumber, tt.want)
			}
		})
	}
}

// TestNewDBEpisodeGenerationSummary verifies that the notice carries the published count
// and maximum generatable episode number as given. It formats the work's expected episode
// count, leaving it empty when the work records
// none so the template can say so in words. A recorded 0 is a count the work states, not a
// gap, so it must survive as "0" rather than collapsing into the unset case.
//
// [Ja] TestNewDBEpisodeGenerationSummary は案内が公開中のエピソード数と生成可能な最大話数を
// そのまま持ち、作品の予定総話数を整形すること、および作品が記録していない場合は空のまま
// 残し、テンプレートが言葉で示せるように
// することを検証する。記録された 0 は作品が述べた件数であって欠落ではないため、未登録と同一視
// されず "0" のまま残ること。
func TestNewDBEpisodeGenerationSummary(t *testing.T) {
	t.Parallel()

	plannedCount := int32(12)
	zeroPlannedCount := int32(0)

	tests := []struct {
		name             string
		plannedCount     *int32
		wantPlannedCount string
	}{
		{name: "予定総話数あり", plannedCount: &plannedCount, wantPlannedCount: "12"},
		{name: "予定総話数 0 は未登録ではなく 0 として扱う", plannedCount: &zeroPlannedCount, wantPlannedCount: "0"},
		{name: "予定総話数なしは空文字列", plannedCount: nil, wantPlannedCount: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBEpisodeGenerationSummary(tt.plannedCount, 5, 9)

			if got.PlannedCount != tt.wantPlannedCount {
				t.Errorf("PlannedCount = %q, want %q", got.PlannedCount, tt.wantPlannedCount)
			}
			if got.PublishedEpisodeCount != 5 {
				t.Errorf("PublishedEpisodeCount = %d, want 5", got.PublishedEpisodeCount)
			}
			if got.MaxGeneratableEpisodeNumber != 9 {
				t.Errorf("MaxGeneratableEpisodeNumber = %d, want 9", got.MaxGeneratableEpisodeNumber)
			}
		})
	}
}

// TestNewDBEpisodeListItem_RawNumberFormat verifies that a whole numeric number renders
// without a decimal part while a fractional one keeps it, so the list shows "2" rather than
// "2.000000" for a regular episode.
//
// [Ja] TestNewDBEpisodeListItem_RawNumberFormat は整数の数値話数が小数部なしで、小数の話数は
// 小数部を保って描画されることを検証する。通常の話が "2.000000" ではなく "2" と表示される。
func TestNewDBEpisodeListItem_RawNumberFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rawNumber float64
		want      string
	}{
		{name: "整数", rawNumber: 2, want: "2"},
		{name: "0.5 話", rawNumber: 2.5, want: "2.5"},
		{name: "0", rawNumber: 0, want: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBEpisodeListItem(&model.Episode{ID: 1, WorkID: 1, RawNumber: &tt.rawNumber})

			if got.RawNumber != tt.want {
				t.Errorf("RawNumber = %q, want %q", got.RawNumber, tt.want)
			}
		})
	}
}

// TestNewDBEpisodeListItem_StatusFromTimestamps verifies that the display status is derived
// from the episode's unpublished_at / deleted_at timestamps (not the dormant
// episodes.status), with deleted_at taking precedence over unpublished_at.
//
// [Ja] TestNewDBEpisodeListItem_StatusFromTimestamps は表示ステータスがエピソードの
// unpublished_at / deleted_at タイムスタンプ (休眠している episodes.status ではない) から
// 導出され、deleted_at が unpublished_at より優先されることを検証する。
func TestNewDBEpisodeListItem_StatusFromTimestamps(t *testing.T) {
	t.Parallel()

	unpublishedAt := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		unpublishedAt *time.Time
		deletedAt     *time.Time
		want          PublishingStatus
	}{
		{
			name: "両方 nil なら published",
			want: PublishingStatusPublished,
		},
		{
			name:          "unpublished_at のみなら archived",
			unpublishedAt: &unpublishedAt,
			want:          PublishingStatusArchived,
		},
		{
			name:      "deleted_at のみなら deleted",
			deletedAt: &deletedAt,
			want:      PublishingStatusDeleted,
		},
		{
			name:          "両方あれば deleted_at が優先される",
			unpublishedAt: &unpublishedAt,
			deletedAt:     &deletedAt,
			want:          PublishingStatusDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBEpisodeListItem(&model.Episode{
				ID:            1,
				WorkID:        1,
				UnpublishedAt: tt.unpublishedAt,
				DeletedAt:     tt.deletedAt,
			})

			if got.Status != tt.want {
				t.Errorf("Status = %q, want %q", got.Status, tt.want)
			}
		})
	}
}

// TestNewDBEpisodeListItems verifies that a page of episodes keeps its order and length
// through the conversion, since the list renders the rows in the order the repository
// returned them.
//
// [Ja] TestNewDBEpisodeListItems は 1 ページ分のエピソードが変換を通しても順序と件数を保つ
// ことを検証する。一覧はリポジトリが返した順で行を描画するため。
func TestNewDBEpisodeListItems(t *testing.T) {
	t.Parallel()

	got := NewDBEpisodeListItems([]*model.Episode{
		{ID: 3, WorkID: 1},
		{ID: 1, WorkID: 1},
		{ID: 2, WorkID: 1},
	})

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}

	wantIDs := []EpisodeID{3, 1, 2}
	for i, want := range wantIDs {
		if got[i].ID != want {
			t.Errorf("[%d].ID = %q, want %q", i, got[i].ID, want)
		}
	}
}

// TestNewDBEpisodeManualCreationRestriction verifies that the domain state projects onto the
// reason the page states, including the order the two conditions are resolved in: a work that
// satisfies both reports the filled count. The page renders one warning per value, so a
// projection that drifted would state the wrong reason or none at all.
//
// [Ja] TestNewDBEpisodeManualCreationRestriction は、ドメインの状態がページの述べる理由へ
// 射影されること、および 2 つの条件を解決する順序 (両方に当てはまる作品は予定話数到達を報告
// する) を検証する。ページは値ごとに 1 つの警告を描画するため、射影がずれると誤った理由を
// 述べるか、何も述べなくなる。
func TestNewDBEpisodeManualCreationRestriction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state model.ManualEpisodeCreationState
		want  DBEpisodeManualCreationRestriction
	}{
		{
			name:  "制限なし",
			state: model.ManualEpisodeCreationState{},
			want:  DBEpisodeManualCreationAllowed,
		},
		{
			name:  "予定話数到達",
			state: model.ManualEpisodeCreationState{EpisodesFilled: true},
			want:  DBEpisodeManualCreationEpisodesFilled,
		},
		{
			name:  "放送枠あり",
			state: model.ManualEpisodeCreationState{SlotsExist: true},
			want:  DBEpisodeManualCreationSlotsExist,
		},
		{
			name:  "両方に当てはまるときは予定話数到達",
			state: model.ManualEpisodeCreationState{EpisodesFilled: true, SlotsExist: true},
			want:  DBEpisodeManualCreationEpisodesFilled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBEpisodeManualCreationRestriction(tt.state)
			if got != tt.want {
				t.Errorf("NewDBEpisodeManualCreationRestriction() = %q, want %q", got, tt.want)
			}
			if got.Restricted() != (tt.want != DBEpisodeManualCreationAllowed) {
				t.Errorf("Restricted() = %v, want %v", got.Restricted(), tt.want != DBEpisodeManualCreationAllowed)
			}
		})
	}
}

// TestNewDBEpisodeFormInputFromEpisode verifies that a stored episode reaches the edit form
// as the strings its fields render: unset optional columns become empty inputs, the numeric
// number keeps its fraction without gaining trailing zeros, the version the form carries
// keeps the sub-second digits that separate two writes made within the same second, and a NULL
// updated_at is represented by an explicit version rather than by an absent precondition.
//
// [Ja] TestNewDBEpisodeFormInputFromEpisode は、保存済みのエピソードが編集フォームの各欄が
// 描画する文字列として届くことを検証する。未設定の任意カラムは空の入力欄になり、数値話数は
// 末尾の 0 を増やさずに小数を保ち、フォームが運ぶ版は同一秒内の 2 つの書き込みを区別する
// 秒未満の桁を保つ。NULL の updated_at は、前提条件の欠落ではなく明示的な版で表す。
func TestNewDBEpisodeFormInputFromEpisode(t *testing.T) {
	t.Parallel()

	number := "第2話"
	rawNumber := 2.5
	title := "もう、お婿にいけません"
	updatedAt := time.Date(2026, 8, 12, 9, 30, 15, 123456000, time.UTC)

	tests := []struct {
		name    string
		episode *model.Episode
		want    DBEpisodeFormInput
	}{
		{
			name: "全項目が設定されている",
			episode: &model.Episode{
				Number:     &number,
				RawNumber:  &rawNumber,
				SortNumber: 200,
				Title:      &title,
				TitleEn:    "No Longer Marriageable",
				UpdatedAt:  &updatedAt,
			},
			want: DBEpisodeFormInput{
				Number:     "第2話",
				RawNumber:  "2.5",
				SortNumber: "200",
				Title:      "もう、お婿にいけません",
				TitleEn:    "No Longer Marriageable",
				UpdatedAt:  "2026-08-12T09:30:15.123456Z",
			},
		},
		{
			name:    "任意カラムが未設定",
			episode: &model.Episode{SortNumber: 100},
			want: DBEpisodeFormInput{
				SortNumber: "100",
				UpdatedAt:  DBEpisodeFormNullVersion,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBEpisodeFormInputFromEpisode(tt.episode)
			if got != tt.want {
				t.Errorf("NewDBEpisodeFormInputFromEpisode() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestNewDBEpisodeFormInputFromEpisode_VersionRoundTrips verifies the property the update side
// relies on: the version the form carries parses back to the instant it was read from, whatever
// offset the stored time arrives in. Without it the comparison against the stored column could
// never match and every submit would be reported as a conflict.
//
// [Ja] TestNewDBEpisodeFormInputFromEpisode_VersionRoundTrips は更新側が依存する性質を検証
// する。フォームが運ぶ版は、保存済みの時刻がどのオフセットで届いても、読み取った時刻へ
// パースし直せる。これが崩れると保存済みカラムとの比較が永久に一致せず、すべての送信が
// 競合として報告される。
func TestNewDBEpisodeFormInputFromEpisode_VersionRoundTrips(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		updatedAt time.Time
	}{
		{name: "UTC", updatedAt: time.Date(2026, 8, 12, 9, 30, 15, 123456000, time.UTC)},
		{name: "UTC 以外のオフセット", updatedAt: time.Date(2026, 8, 12, 18, 30, 15, 123456000, time.FixedZone("JST", 9*60*60))},
		{name: "秒未満が 0", updatedAt: time.Date(2026, 8, 12, 9, 30, 15, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDBEpisodeFormInputFromEpisode(&model.Episode{UpdatedAt: &tt.updatedAt})

			parsed, err := time.Parse(dbEpisodeFormVersionLayout, got.UpdatedAt)
			if err != nil {
				t.Fatalf("版 %q をパースできません: %v", got.UpdatedAt, err)
			}
			if !parsed.Equal(tt.updatedAt) {
				t.Errorf("版 %q は %v へ戻りました, want %v", got.UpdatedAt, parsed, tt.updatedAt)
			}
		})
	}
}
