package model

import "time"

// AnimeRevisionは作品・エピソードの版。DBEditIDがnilの版は起点の版 (記録を始めた時点の状態) で、
// 履歴の一覧には並べず、次の版の差分の比べる相手にだけ使う。
type AnimeRevision struct {
	ID        AnimeRevisionID
	DBEditID  *DBEditID
	AnimeID   AnimeID
	Snapshot  AnimeSnapshot
	CreatedAt time.Time

	// DBEditは版が属する操作。値が入るのは履歴のローダー (AnimeRevisionRepository.ListHistoryByAnimeIDと
	// ListHistoryByWorkAnimeID) のみ。
	DBEdit *DBEdit
	// Previousは同じ作品・エピソードの1つ前の版 (起点の版を含む) のスナップショットで、差分の比べる相手。
	// 1つ前の版が無いときはnil。値が入るのは履歴のローダーのみ。
	Previous *AnimeSnapshot
	// Workは版が属する今の作品 (IDとTitleのみ)。作品の版ではその作品、エピソードの版では今の親作品を指す。
	// 値が入るのはAnnict DB全体の編集履歴のローダー (AnimeRevisionRepository.ListForDBEdits) のみ。
	Work *Work
	// Episodeは版がエピソードの版のときの、そのエピソード (IDのみ)。値が入るのは作品の履歴のローダー
	// (AnimeRevisionRepository.ListHistoryByWorkAnimeID) とAnnict DB全体の編集履歴のローダー
	// (ListForDBEdits) のみで、作品の版ではnil。
	Episode *Episode
}

// AnimeSnapshotは作品・エピソードの版が持つ、変更後の状態。animesとanime_classificationsの編集できる列と、
// 6つの別表の行を持つ。別表の行は作品だけが持ち、エピソードでは空になる。集計・時刻の列は持たない。
type AnimeSnapshot struct {
	// animesの列
	Title            *string
	TitleKana        *string
	TitleRo          *string
	TitleEn          *string
	TitleAlter       *string
	TitleAlterRo     *string
	TitleAlterEn     *string
	TitleAlterOther  *string
	Media            *AnimeMedia
	ReleaseStatus    *ReleaseStatus
	Synopsis         *string
	SynopsisEn       *string
	SynopsisSource   *string
	SynopsisSourceEn *string
	Status           AnimeStatus
	ArchiveMessage   *string

	// anime_classificationsの列。Number・EpisodeStartNumberはNUMERICの値を丸めずに文字列で持つ。
	Kind                  AnimeClassificationKind
	ParentAnimeID         *AnimeID
	Number                *string
	NumberText            *string
	SortNumber            *int32
	Standalone            bool
	NumberFormatID        *NumberFormatID
	EpisodeStartNumber    *string
	ExpectedEpisodesCount *int32

	// 別表の行
	ExternalIDs      []AnimeSnapshotExternalID
	Links            []AnimeSnapshotLink
	OfficialAccounts []AnimeSnapshotOfficialAccount
	Hashtags         []AnimeSnapshotHashtag
	Seasons          []AnimeSnapshotSeason
	Events           []AnimeSnapshotEvent
}

// AnimeSnapshotExternalIDはanime_external_idsの1行の内容。
type AnimeSnapshotExternalID struct {
	Service    AnimeExternalService
	ExternalID string
}

// AnimeSnapshotLinkはanime_linksの1行の内容。
type AnimeSnapshotLink struct {
	Kind       AnimeLinkKind
	Language   Language
	URL        string
	Label      *string
	LabelEn    *string
	SortNumber int32
}

// AnimeSnapshotOfficialAccountはanime_official_accountsの1行の内容。
type AnimeSnapshotOfficialAccount struct {
	Service    AnimeAccountService
	Account    string
	Label      *string
	LabelEn    *string
	SortNumber int32
}

// AnimeSnapshotHashtagはanime_hashtagsの1行の内容。
type AnimeSnapshotHashtag struct {
	Hashtag    string
	SortNumber int32
}

// AnimeSnapshotSeasonはanime_seasonsの1行の内容。
type AnimeSnapshotSeason struct {
	Year      int32
	Name      *SeasonName
	IsPrimary bool
}

// AnimeSnapshotEventはanime_eventsの1行の内容。StartedOn・EndedOnは日付だけを持つ。
type AnimeSnapshotEvent struct {
	Kind          AnimeEventKind
	StartedOn     time.Time
	EndedOn       *time.Time
	Title         *string
	TitleEn       *string
	Description   *string
	DescriptionEn *string
	SortNumber    int32
}

// DiffAnimeSnapshotsは作品・エピソードの2つの版の差分を返す。beforeがnilのときは作成の版として、
// 変更後だけを返す。
func DiffAnimeSnapshots(before *AnimeSnapshot, after AnimeSnapshot) RevisionDiff {
	if before == nil {
		return diffRevisions(false, nil, after.revisionValues(), nil, after.revisionRows())
	}
	return diffRevisions(true, before.revisionValues(), after.revisionValues(), before.revisionRows(), after.revisionRows())
}

func (s AnimeSnapshot) revisionValues() []RevisionFieldValue {
	return []RevisionFieldValue{
		{Field: "title", Value: s.Title},
		{Field: "title_kana", Value: s.TitleKana},
		{Field: "title_ro", Value: s.TitleRo},
		{Field: "title_en", Value: s.TitleEn},
		{Field: "title_alter", Value: s.TitleAlter},
		{Field: "title_alter_ro", Value: s.TitleAlterRo},
		{Field: "title_alter_en", Value: s.TitleAlterEn},
		{Field: "title_alter_other", Value: s.TitleAlterOther},
		{Field: "media", Value: revisionOptionalString(s.Media)},
		{Field: "release_status", Value: revisionOptionalString(s.ReleaseStatus)},
		{Field: "synopsis", Value: s.Synopsis},
		{Field: "synopsis_en", Value: s.SynopsisEn},
		{Field: "synopsis_source", Value: s.SynopsisSource},
		{Field: "synopsis_source_en", Value: s.SynopsisSourceEn},
		{Field: "status", Value: revisionString(string(s.Status))},
		{Field: "archive_message", Value: s.ArchiveMessage},
		{Field: "kind", Value: revisionString(string(s.Kind))},
		{Field: "parent_anime_id", Value: revisionOptionalID(s.ParentAnimeID)},
		{Field: "number", Value: s.Number},
		{Field: "number_text", Value: s.NumberText},
		{Field: "sort_number", Value: revisionOptionalInt32(s.SortNumber)},
		{Field: "standalone", Value: revisionBool(s.Standalone)},
		{Field: "number_format_id", Value: revisionOptionalID(s.NumberFormatID)},
		{Field: "episode_start_number", Value: s.EpisodeStartNumber},
		{Field: "expected_episodes_count", Value: revisionOptionalInt32(s.ExpectedEpisodesCount)},
	}
}

func (s AnimeSnapshot) revisionRows() []RevisionRow {
	var rows []RevisionRow
	for _, r := range s.ExternalIDs {
		rows = append(rows, RevisionRow{Table: RevisionTableExternalIDs, Values: []RevisionFieldValue{
			{Field: "service", Value: revisionString(string(r.Service))},
			{Field: "external_id", Value: revisionString(r.ExternalID)},
		}})
	}
	for _, r := range s.Links {
		rows = append(rows, RevisionRow{Table: RevisionTableLinks, Values: []RevisionFieldValue{
			{Field: "kind", Value: revisionString(string(r.Kind))},
			{Field: "language", Value: revisionString(string(r.Language))},
			{Field: "url", Value: revisionString(r.URL)},
			{Field: "label", Value: r.Label},
			{Field: "label_en", Value: r.LabelEn},
			{Field: "sort_number", Value: revisionInt32(r.SortNumber)},
		}})
	}
	for _, r := range s.OfficialAccounts {
		rows = append(rows, RevisionRow{Table: RevisionTableOfficialAccounts, Values: []RevisionFieldValue{
			{Field: "service", Value: revisionString(string(r.Service))},
			{Field: "account", Value: revisionString(r.Account)},
			{Field: "label", Value: r.Label},
			{Field: "label_en", Value: r.LabelEn},
			{Field: "sort_number", Value: revisionInt32(r.SortNumber)},
		}})
	}
	for _, r := range s.Hashtags {
		rows = append(rows, RevisionRow{Table: RevisionTableHashtags, Values: []RevisionFieldValue{
			{Field: "hashtag", Value: revisionString(r.Hashtag)},
			{Field: "sort_number", Value: revisionInt32(r.SortNumber)},
		}})
	}
	for _, r := range s.Seasons {
		rows = append(rows, RevisionRow{Table: RevisionTableSeasons, Values: []RevisionFieldValue{
			{Field: "year", Value: revisionInt32(r.Year)},
			{Field: "name", Value: revisionOptionalString(r.Name)},
			{Field: "is_primary", Value: revisionBool(r.IsPrimary)},
		}})
	}
	for _, r := range s.Events {
		rows = append(rows, RevisionRow{Table: RevisionTableEvents, Values: []RevisionFieldValue{
			{Field: "kind", Value: revisionString(string(r.Kind))},
			{Field: "started_on", Value: revisionDate(r.StartedOn)},
			{Field: "ended_on", Value: revisionOptionalDate(r.EndedOn)},
			{Field: "title", Value: r.Title},
			{Field: "title_en", Value: r.TitleEn},
			{Field: "description", Value: r.Description},
			{Field: "description_en", Value: r.DescriptionEn},
			{Field: "sort_number", Value: revisionInt32(r.SortNumber)},
		}})
	}
	return rows
}
