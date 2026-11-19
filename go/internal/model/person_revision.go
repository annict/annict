package model

import "time"

// PersonRevisionは人物の版。DBEditIDがnilの版は起点の版で、AnimeRevisionと同じく次の版の差分の
// 比べる相手にだけ使う。
type PersonRevision struct {
	ID        PersonRevisionID
	DBEditID  *DBEditID
	PersonID  PersonID
	Snapshot  PersonSnapshot
	CreatedAt time.Time

	// DBEditは版が属する操作。値が入るのは履歴のローダー (PersonRevisionRepository.ListHistoryByPersonID) のみ。
	DBEdit *DBEdit
	// Previousは同じ人物の1つ前の版 (起点の版と、管理者以外に隠す版を含む) のスナップショットで、差分の比べる相手。
	// 1つ前の版が無いときはnil。値が入るのは履歴のローダーのみ。
	Previous *PersonSnapshot
}

// PersonSnapshotは人物の版が持つ、変更後の状態。peopleの編集できる列とStatus・ArchiveMessageを持つ。
// 掲載停止の対応記録 (delist_message) は、管理者が直しても古い内容が消せない履歴に残り続けないよう持たない。
// 集計・時刻の列とRails版の状態の列も持たない。
type PersonSnapshot struct {
	Name              string
	NameKana          string
	NameEn            string
	Nickname          *string
	NicknameEn        string
	Gender            *PersonGender
	BloodType         *PersonBloodType
	PrefectureID      *PrefectureID
	Birthday          *time.Time
	Height            *int32
	URL               *string
	URLEn             string
	WikipediaURL      *string
	WikipediaURLEn    string
	TwitterUsername   *string
	TwitterUsernameEn string
	Status            PersonStatus
	ArchiveMessage    *string
}

// DiffPersonSnapshotsは人物の2つの版の差分を返す。beforeがnilのときは作成の版として、変更後だけを返す。
func DiffPersonSnapshots(before *PersonSnapshot, after PersonSnapshot) RevisionDiff {
	if before == nil {
		return diffRevisions(false, nil, after.revisionValues(), nil, nil)
	}
	return diffRevisions(true, before.revisionValues(), after.revisionValues(), nil, nil)
}

func (s PersonSnapshot) revisionValues() []RevisionFieldValue {
	return []RevisionFieldValue{
		{Field: "name", Value: revisionString(s.Name)},
		{Field: "name_kana", Value: revisionString(s.NameKana)},
		{Field: "name_en", Value: revisionString(s.NameEn)},
		{Field: "nickname", Value: s.Nickname},
		{Field: "nickname_en", Value: revisionString(s.NicknameEn)},
		{Field: "gender", Value: revisionOptionalString(s.Gender)},
		{Field: "blood_type", Value: revisionOptionalString(s.BloodType)},
		{Field: "prefecture_id", Value: revisionOptionalID(s.PrefectureID)},
		{Field: "birthday", Value: revisionOptionalDate(s.Birthday)},
		{Field: "height", Value: revisionOptionalInt32(s.Height)},
		{Field: "url", Value: s.URL},
		{Field: "url_en", Value: revisionString(s.URLEn)},
		{Field: "wikipedia_url", Value: s.WikipediaURL},
		{Field: "wikipedia_url_en", Value: revisionString(s.WikipediaURLEn)},
		{Field: "twitter_username", Value: s.TwitterUsername},
		{Field: "twitter_username_en", Value: revisionString(s.TwitterUsernameEn)},
		{Field: "status", Value: revisionString(string(s.Status))},
		{Field: "archive_message", Value: s.ArchiveMessage},
	}
}
