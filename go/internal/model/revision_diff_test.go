package model_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

func ptr[T any](v T) *T { return &v }

func baseAnimeSnapshot() model.AnimeSnapshot {
	return model.AnimeSnapshot{
		Title:  ptr("作品"),
		Media:  ptr(model.AnimeMediaTV),
		Status: model.AnimeStatusPublished,
		Kind:   model.AnimeClassificationKindWork,
		Links: []model.AnimeSnapshotLink{
			{Kind: model.AnimeLinkKindOfficialSite, Language: model.LanguageJa, URL: "https://example.com/", SortNumber: 1},
		},
		Hashtags: []model.AnimeSnapshotHashtag{
			{Hashtag: "a", SortNumber: 1},
			{Hashtag: "b", SortNumber: 2},
		},
	}
}

func TestDiffAnimeSnapshots(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 同じ内容では差分が無い", func(t *testing.T) {
		t.Parallel()

		before := baseAnimeSnapshot()
		after := baseAnimeSnapshot()
		// 別表の行は並び順を問わない。
		after.Hashtags = []model.AnimeSnapshotHashtag{after.Hashtags[1], after.Hashtags[0]}

		if diff := model.DiffAnimeSnapshots(&before, after); !diff.IsEmpty() {
			t.Errorf("差分 = %+v、差分が無いことを期待", diff)
		}
	})

	t.Run("正常系: 変わった項目だけを変更前と変更後で返す", func(t *testing.T) {
		t.Parallel()

		before := baseAnimeSnapshot()
		after := baseAnimeSnapshot()
		after.Title = ptr("新しい作品")
		after.Media = nil
		after.Standalone = true

		diff := model.DiffAnimeSnapshots(&before, after)
		want := []model.RevisionFieldChange{
			{Field: "title", Before: ptr("作品"), After: ptr("新しい作品")},
			{Field: "media", Before: ptr("tv"), After: nil},
			{Field: "standalone", Before: ptr("false"), After: ptr("true")},
		}
		if !reflect.DeepEqual(diff.Fields, want) {
			t.Errorf("Fields = %s、期待値 = %s", formatFieldChanges(diff.Fields), formatFieldChanges(want))
		}
		if len(diff.Rows) != 0 {
			t.Errorf("Rowsの件数 = %d、期待値 = 0", len(diff.Rows))
		}
	})

	t.Run("正常系: 別表の行の変更を削除と追加で返す", func(t *testing.T) {
		t.Parallel()

		before := baseAnimeSnapshot()
		after := baseAnimeSnapshot()
		after.Links[0].SortNumber = 2
		after.Seasons = []model.AnimeSnapshotSeason{{Year: 2026, Name: ptr(model.SeasonNameFall), IsPrimary: true}}

		diff := model.DiffAnimeSnapshots(&before, after)
		if len(diff.Fields) != 0 {
			t.Errorf("Fieldsの件数 = %d、期待値 = 0", len(diff.Fields))
		}
		got := make([]string, len(diff.Rows))
		for i, r := range diff.Rows {
			got[i] = string(r.Kind) + ":" + string(r.Row.Table)
		}
		want := []string{"removed:links", "added:links", "added:seasons"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Rows = %v、期待値 = %v", got, want)
		}
		season := diff.Rows[2].Row.Values
		if *season[0].Value != "2026" || *season[1].Value != "fall" || *season[2].Value != "true" {
			t.Errorf("追加したシーズンの値 = %s、期待値 = [2026 fall true]", formatValues(season))
		}
	})

	t.Run("正常系: 同じ内容の行が増えたときは増えた分を追加で返す", func(t *testing.T) {
		t.Parallel()

		before := baseAnimeSnapshot()
		after := baseAnimeSnapshot()
		after.Hashtags = append(after.Hashtags, after.Hashtags[0])

		diff := model.DiffAnimeSnapshots(&before, after)
		if len(diff.Rows) != 1 || diff.Rows[0].Kind != model.RevisionRowAdded {
			t.Fatalf("Rows = %+v、追加1件を期待", diff.Rows)
		}
	})

	t.Run("正常系: 作成の版は値のある項目とすべての行を変更後だけで返す", func(t *testing.T) {
		t.Parallel()

		after := baseAnimeSnapshot()
		after.TitleEn = ptr("")
		startedOn := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		after.Events = []model.AnimeSnapshotEvent{{Kind: model.AnimeEventKindBroadcast, StartedOn: startedOn, SortNumber: 1}}

		diff := model.DiffAnimeSnapshots(nil, after)
		want := []model.RevisionFieldChange{
			{Field: "title", After: ptr("作品")},
			{Field: "media", After: ptr("tv")},
			{Field: "status", After: ptr("published")},
			{Field: "kind", After: ptr("work")},
			{Field: "standalone", After: ptr("false")},
		}
		if !reflect.DeepEqual(diff.Fields, want) {
			t.Errorf("Fields = %s、期待値 = %s", formatFieldChanges(diff.Fields), formatFieldChanges(want))
		}
		if len(diff.Rows) != 4 {
			t.Fatalf("Rowsの件数 = %d、期待値 = 4", len(diff.Rows))
		}
		for _, r := range diff.Rows {
			if r.Kind != model.RevisionRowAdded {
				t.Errorf("行の変化 = %q、期待値 = added", r.Kind)
			}
		}
		if got := diff.Rows[3].Row.Values[1].Value; got == nil || *got != "2026-10-01" {
			t.Errorf("イベントの開始日 = %v、期待値 = 2026-10-01", got)
		}
	})

	t.Run("正常系: 未設定と空文字列を区別する", func(t *testing.T) {
		t.Parallel()

		before := baseAnimeSnapshot()
		after := baseAnimeSnapshot()
		after.ArchiveMessage = ptr("")
		after.Links[0].Label = ptr("")

		diff := model.DiffAnimeSnapshots(&before, after)
		if len(diff.Fields) != 1 || diff.Fields[0].Field != "archive_message" {
			t.Errorf("Fields = %s、archive_messageの変更だけを期待", formatFieldChanges(diff.Fields))
		}
		if len(diff.Rows) != 2 {
			t.Errorf("Rowsの件数 = %d、期待値 = 2", len(diff.Rows))
		}
	})
}

func TestDiffPersonSnapshots(t *testing.T) {
	t.Parallel()

	before := model.PersonSnapshot{Name: "人物", Status: model.PersonStatusPublished}

	t.Run("正常系: 同じ内容では差分が無い", func(t *testing.T) {
		t.Parallel()

		if diff := model.DiffPersonSnapshots(&before, before); !diff.IsEmpty() {
			t.Errorf("差分 = %+v、差分が無いことを期待", diff)
		}
	})

	t.Run("正常系: 変わった項目だけを返す", func(t *testing.T) {
		t.Parallel()

		after := before
		after.Status = model.PersonStatusArchived
		after.ArchiveMessage = ptr("非公開にした理由")
		after.Birthday = ptr(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC))

		diff := model.DiffPersonSnapshots(&before, after)
		want := []model.RevisionFieldChange{
			{Field: "birthday", After: ptr("1990-01-02")},
			{Field: "status", Before: ptr("published"), After: ptr("archived")},
			{Field: "archive_message", After: ptr("非公開にした理由")},
		}
		if !reflect.DeepEqual(diff.Fields, want) {
			t.Errorf("Fields = %s、期待値 = %s", formatFieldChanges(diff.Fields), formatFieldChanges(want))
		}
	})

	t.Run("正常系: 作成の版は空文字列の項目を返さない", func(t *testing.T) {
		t.Parallel()

		diff := model.DiffPersonSnapshots(nil, model.PersonSnapshot{Name: "人物", Height: ptr(int32(170)), Status: model.PersonStatusPublished})
		want := []model.RevisionFieldChange{
			{Field: "name", After: ptr("人物")},
			{Field: "height", After: ptr("170")},
			{Field: "status", After: ptr("published")},
		}
		if !reflect.DeepEqual(diff.Fields, want) {
			t.Errorf("Fields = %s、期待値 = %s", formatFieldChanges(diff.Fields), formatFieldChanges(want))
		}
	})
}

func formatValue(v *string) string {
	if v == nil {
		return "<nil>"
	}
	return *v
}

func formatFieldChanges(changes []model.RevisionFieldChange) string {
	s := "["
	for i, c := range changes {
		if i > 0 {
			s += " "
		}
		s += string(c.Field) + ":" + formatValue(c.Before) + "->" + formatValue(c.After)
	}
	return s + "]"
}

func formatValues(values []model.RevisionFieldValue) string {
	s := "["
	for i, v := range values {
		if i > 0 {
			s += " "
		}
		s += formatValue(v.Value)
	}
	return s + "]"
}
