package model

import (
	"strconv"
	"strings"
	"time"
)

// RevisionFieldは版のスナップショットの項目を表す。別表の行の項目にも使う。
type RevisionField string

// RevisionTableは作品のスナップショットが行を持つ別表を表す。
type RevisionTable string

const (
	RevisionTableExternalIDs      RevisionTable = "external_ids"
	RevisionTableLinks            RevisionTable = "links"
	RevisionTableOfficialAccounts RevisionTable = "official_accounts"
	RevisionTableHashtags         RevisionTable = "hashtags"
	RevisionTableSeasons          RevisionTable = "seasons"
	RevisionTableEvents           RevisionTable = "events"
)

// RevisionFieldValueは版の1つの項目の値を、表示と比較に使える文字列で持つ。Valueがnilの項目は未設定。
type RevisionFieldValue struct {
	Field RevisionField
	Value *string
}

// RevisionRowは別表の1行を、項目の値の並びで持つ。
type RevisionRow struct {
	Table  RevisionTable
	Values []RevisionFieldValue
}

// RevisionFieldChangeは1つの項目の変更前と変更後の値。作成の版ではBeforeが常にnil。
type RevisionFieldChange struct {
	Field  RevisionField
	Before *string
	After  *string
}

// RevisionRowChangeKindは別表の行の変化の種類。
type RevisionRowChangeKind string

const (
	RevisionRowAdded   RevisionRowChangeKind = "added"
	RevisionRowRemoved RevisionRowChangeKind = "removed"
)

// RevisionRowChangeは別表の行の追加・削除。行の項目が変わったときは、変更前の行の削除と変更後の行の
// 追加として表す。
type RevisionRowChange struct {
	Kind RevisionRowChangeKind
	Row  RevisionRow
}

// RevisionDiffは同じリソースの2つの版の差分。
type RevisionDiff struct {
	Fields []RevisionFieldChange
	Rows   []RevisionRowChange
}

// IsEmptyは差分が無いかを返す。
func (d RevisionDiff) IsEmpty() bool {
	return len(d.Fields) == 0 && len(d.Rows) == 0
}

// diffRevisionsは変更前と変更後の項目・別表の行を比べて差分を返す。hasBeforeがfalseのときは作成の版として、
// 値のある項目と、すべての行を変更後だけで返す。項目の並びは変更前と変更後で同じ前提で、別表の行は
// 並び順を問わずに比べる。
func diffRevisions(hasBefore bool, before, after []RevisionFieldValue, beforeRows, afterRows []RevisionRow) RevisionDiff {
	var diff RevisionDiff

	for i, a := range after {
		if !hasBefore {
			if a.Value != nil && *a.Value != "" {
				diff.Fields = append(diff.Fields, RevisionFieldChange{Field: a.Field, After: a.Value})
			}
			continue
		}
		b := before[i]
		if !equalRevisionValue(b.Value, a.Value) {
			diff.Fields = append(diff.Fields, RevisionFieldChange{Field: a.Field, Before: b.Value, After: a.Value})
		}
	}

	// 同じ内容の行が複数あっても数が合うよう、行の内容ごとの数で比べる。
	remaining := make(map[string]int, len(afterRows))
	for _, row := range afterRows {
		remaining[row.key()]++
	}
	for _, row := range beforeRows {
		key := row.key()
		if remaining[key] > 0 {
			remaining[key]--
			continue
		}
		diff.Rows = append(diff.Rows, RevisionRowChange{Kind: RevisionRowRemoved, Row: row})
	}
	unmatched := make(map[string]int, len(beforeRows))
	for _, row := range beforeRows {
		unmatched[row.key()]++
	}
	for _, row := range afterRows {
		key := row.key()
		if unmatched[key] > 0 {
			unmatched[key]--
			continue
		}
		diff.Rows = append(diff.Rows, RevisionRowChange{Kind: RevisionRowAdded, Row: row})
	}

	return diff
}

// keyは行を比べるための、別表と項目の値をつないだ文字列を返す。未設定の値と空文字列を区別する。
func (r RevisionRow) key() string {
	var b strings.Builder
	b.WriteString(string(r.Table))
	for _, v := range r.Values {
		b.WriteByte(0)
		if v.Value == nil {
			b.WriteByte(1)
			continue
		}
		b.WriteByte(2)
		b.WriteString(*v.Value)
	}
	return b.String()
}

func equalRevisionValue(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// 以下は、スナップショットの値を版の項目の文字列にする。

func revisionString(v string) *string { return &v }

func revisionOptionalString[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func revisionInt32(v int32) *string { return revisionString(strconv.FormatInt(int64(v), 10)) }

func revisionOptionalInt32(v *int32) *string {
	if v == nil {
		return nil
	}
	return revisionInt32(*v)
}

func revisionOptionalID[T ~int64](v *T) *string {
	if v == nil {
		return nil
	}
	return revisionString(strconv.FormatInt(int64(*v), 10))
}

func revisionBool(v bool) *string { return revisionString(strconv.FormatBool(v)) }

func revisionDate(v time.Time) *string { return revisionString(v.Format(time.DateOnly)) }

func revisionOptionalDate(v *time.Time) *string {
	if v == nil {
		return nil
	}
	return revisionDate(*v)
}
