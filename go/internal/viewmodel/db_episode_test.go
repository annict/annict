package viewmodel

import (
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// TestDBEpisodeListWorkName verifies that the work name is trimmed and that a title holding
// nothing but whitespace collapses to the empty string, which the heading and the document
// title both read as "fall back to the generic page name".
//
// [Ja] TestDBEpisodeListWorkName は作品の名前が前後の空白を落として返ること、および空白だけの
// タイトルが空文字列に畳まれることを検証する。見出しと文書タイトルはいずれもこれを「汎用の
// ページ名へフォールバックする」合図として読む。
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
				ID:         10,
				WorkID:     3,
				Number:     &number,
				RawNumber:  &rawNumber,
				Title:      &title,
				TitleEn:    "Episode Title",
				SortNumber: 200,
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
