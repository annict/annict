package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// animeSnapshotVersionは、anime_revisions.snapshotに書く保存形式の版。形を変えるときは版を上げ、
// 古い形を読むときに新しい形へ読み替える。
const animeSnapshotVersion = 1

// AnimeRevisionRepositoryは作品・エピソードの版 (anime_revisions) のデータアクセスを担当します。
// スナップショットとjsonbの保存形式の変換は、このリポジトリに閉じ込めます。
type AnimeRevisionRepository struct {
	queries *query.Queries
}

// NewAnimeRevisionRepositoryはAnimeRevisionRepositoryを作成します
func NewAnimeRevisionRepository(queries *query.Queries) *AnimeRevisionRepository {
	return &AnimeRevisionRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *AnimeRevisionRepository) WithTx(tx *sql.Tx) *AnimeRevisionRepository {
	return &AnimeRevisionRepository{queries: r.queries.WithTx(tx)}
}

// Createは操作dbEditIDに属する版を作成します。parent_anime_idの列には、スナップショットの親の作品を書きます。
func (r *AnimeRevisionRepository) Create(ctx context.Context, dbEditID model.DBEditID, animeID model.AnimeID, snapshot model.AnimeSnapshot) error {
	data, err := encodeAnimeSnapshot(snapshot)
	if err != nil {
		return err
	}
	return r.queries.CreateAnimeRevision(ctx, query.CreateAnimeRevisionParams{
		DbEditID:        sql.NullInt64{Int64: int64(dbEditID), Valid: true},
		AnimeID:         int64(animeID),
		ParentAnimeID:   nullInt64FromAnimeID(snapshot.ParentAnimeID),
		Snapshot:        data,
		SnapshotVersion: animeSnapshotVersion,
	})
}

// CreateInitialは起点の版を作成します。同じ作品・エピソードの起点の版がすでにあるときは何もしません。
func (r *AnimeRevisionRepository) CreateInitial(ctx context.Context, animeID model.AnimeID, snapshot model.AnimeSnapshot) error {
	data, err := encodeAnimeSnapshot(snapshot)
	if err != nil {
		return err
	}
	return r.queries.CreateInitialAnimeRevision(ctx, query.CreateInitialAnimeRevisionParams{
		AnimeID:         int64(animeID),
		ParentAnimeID:   nullInt64FromAnimeID(snapshot.ParentAnimeID),
		Snapshot:        data,
		SnapshotVersion: animeSnapshotVersion,
	})
}

// ListAnimeIDsWithRevisionsはanimeIDsのうち、版を1つ以上持つ作品・エピソードのIDの集合を返します。
// 空の入力ではクエリせずに空の集合を返します。
func (r *AnimeRevisionRepository) ListAnimeIDsWithRevisions(ctx context.Context, animeIDs []model.AnimeID) (map[model.AnimeID]bool, error) {
	result := make(map[model.AnimeID]bool, len(animeIDs))
	if len(animeIDs) == 0 {
		return result, nil
	}

	ids := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		ids[i] = int64(id)
	}
	rows, err := r.queries.ListAnimeIDsWithRevisions(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, id := range rows {
		result[model.AnimeID(id)] = true
	}
	return result, nil
}

// ListByAnimeIDは作品・エピソードの版を、起点の版を含めてIDの昇順に返します。
func (r *AnimeRevisionRepository) ListByAnimeID(ctx context.Context, animeID model.AnimeID) ([]*model.AnimeRevision, error) {
	rows, err := r.queries.ListAnimeRevisionsByAnimeID(ctx, int64(animeID))
	if err != nil {
		return nil, err
	}

	revisions := make([]*model.AnimeRevision, len(rows))
	for i, row := range rows {
		snapshot, err := decodeAnimeSnapshot(row.SnapshotVersion, row.Snapshot)
		if err != nil {
			return nil, fmt.Errorf("作品・エピソードの版%dのスナップショットの読み込みに失敗: %w", row.ID, err)
		}
		revision := &model.AnimeRevision{
			ID:        model.AnimeRevisionID(row.ID),
			AnimeID:   model.AnimeID(row.AnimeID),
			Snapshot:  snapshot,
			CreatedAt: row.CreatedAt,
		}
		if row.DbEditID.Valid {
			dbEditID := model.DBEditID(row.DbEditID.Int64)
			revision.DBEditID = &dbEditID
		}
		revisions[i] = revision
	}
	return revisions, nil
}

// AnimeRevisionHistoryParamsは作品・エピソードの版の履歴の取得条件を保持します。
type AnimeRevisionHistoryParams struct {
	AnimeID model.AnimeID
	Page    int32
	PerPage int32
}

// ListHistoryByAnimeIDは作品・エピソードの操作の版を、新しい順に1ページ分返します。起点の版は並べず、
// 各版のPreviousに、起点の版を含めた1つ前の版のスナップショットを入れます。
func (r *AnimeRevisionRepository) ListHistoryByAnimeID(ctx context.Context, params AnimeRevisionHistoryParams) ([]*model.AnimeRevision, error) {
	// 乗算の前に幅を広げる。ページ番号がint32の範囲で大きいとint32同士の積が負に折り返し、
	// PostgreSQLがそのOFFSETを拒否するため (PersonRepository.ListForDBと同じ)。
	offset := int64(params.Page-1) * int64(params.PerPage)

	rows, err := r.queries.ListAnimeRevisionHistory(ctx, query.ListAnimeRevisionHistoryParams{
		AnimeID:    int64(params.AnimeID),
		PerPage:    params.PerPage,
		PageOffset: offset,
	})
	if err != nil {
		return nil, err
	}

	revisions := make([]*model.AnimeRevision, len(rows))
	for i, row := range rows {
		revision, err := animeRevisionFromHistoryRow(query.ListWorkAnimeRevisionHistoryRow{
			ID:                      row.ID,
			DbEditID:                row.DbEditID,
			AnimeID:                 row.AnimeID,
			Snapshot:                row.Snapshot,
			SnapshotVersion:         row.SnapshotVersion,
			CreatedAt:               row.CreatedAt,
			PreviousSnapshot:        row.PreviousSnapshot,
			PreviousSnapshotVersion: row.PreviousSnapshotVersion,
			Actor:                   row.Actor,
			UserID:                  row.UserID,
			Action:                  row.Action,
			DbEditCreatedAt:         row.DbEditCreatedAt,
			Username:                row.Username,
			ProfileImageData:        row.ProfileImageData,
		})
		if err != nil {
			return nil, err
		}
		revisions[i] = revision
	}
	return revisions, nil
}

// CountHistoryByAnimeIDは作品・エピソードの操作の版の数を返します。起点の版は数えません。
func (r *AnimeRevisionRepository) CountHistoryByAnimeID(ctx context.Context, animeID model.AnimeID) (int64, error) {
	return r.queries.CountAnimeRevisionHistory(ctx, int64(animeID))
}

// ListHistoryByWorkAnimeIDは、作品の操作の版と、版の時点でその作品の配下にあったエピソードの操作の版を、
// 新しい順に1ページ分返します。削除したエピソードの版は並べません。エピソードの版にはEpisodeを入れます。
func (r *AnimeRevisionRepository) ListHistoryByWorkAnimeID(ctx context.Context, params AnimeRevisionHistoryParams) ([]*model.AnimeRevision, error) {
	offset := int64(params.Page-1) * int64(params.PerPage)

	rows, err := r.queries.ListWorkAnimeRevisionHistory(ctx, query.ListWorkAnimeRevisionHistoryParams{
		AnimeID:    int64(params.AnimeID),
		PerPage:    params.PerPage,
		PageOffset: offset,
	})
	if err != nil {
		return nil, err
	}

	revisions := make([]*model.AnimeRevision, len(rows))
	for i, row := range rows {
		revision, err := animeRevisionFromHistoryRow(row)
		if err != nil {
			return nil, err
		}
		if row.EpisodeID != 0 {
			revision.Episode = &model.Episode{ID: model.EpisodeID(row.EpisodeID)}
		}
		revisions[i] = revision
	}
	return revisions, nil
}

// CountHistoryByWorkAnimeIDは、ListHistoryByWorkAnimeIDが並べる版の数を返します。
func (r *AnimeRevisionRepository) CountHistoryByWorkAnimeID(ctx context.Context, animeID model.AnimeID) (int64, error) {
	return r.queries.CountWorkAnimeRevisionHistory(ctx, int64(animeID))
}

// animeRevisionFromHistoryRowは、履歴のクエリの行から、操作と1つ前の版を持つ版を組み立てます。
func animeRevisionFromHistoryRow(row query.ListWorkAnimeRevisionHistoryRow) (*model.AnimeRevision, error) {
	snapshot, err := decodeAnimeSnapshot(row.SnapshotVersion, row.Snapshot)
	if err != nil {
		return nil, fmt.Errorf("作品・エピソードの版%dのスナップショットの読み込みに失敗: %w", row.ID, err)
	}
	dbEditID := model.DBEditID(row.DbEditID)
	edit := &model.DBEdit{
		ID:        dbEditID,
		Actor:     model.DBEditActor(row.Actor),
		Action:    model.DBEditAction(row.Action),
		CreatedAt: row.DbEditCreatedAt,
	}
	if row.UserID.Valid {
		userID := model.UserID(row.UserID.Int64)
		edit.UserID = &userID
		edit.User = &model.User{ID: userID, Username: row.Username, ProfileImageData: row.ProfileImageData}
	}
	revision := &model.AnimeRevision{
		ID:        model.AnimeRevisionID(row.ID),
		DBEditID:  &dbEditID,
		AnimeID:   model.AnimeID(row.AnimeID),
		Snapshot:  snapshot,
		CreatedAt: row.CreatedAt,
		DBEdit:    edit,
	}
	if row.PreviousSnapshotVersion != 0 {
		previous, err := decodeAnimeSnapshot(row.PreviousSnapshotVersion, row.PreviousSnapshot)
		if err != nil {
			return nil, fmt.Errorf("作品・エピソードの版%dの1つ前の版の読み込みに失敗: %w", row.ID, err)
		}
		revision.Previous = &previous
	}
	return revision, nil
}

// 以下はsnapshot_version = 1の保存形式。キーはテーブルの列名にそろえ、値の無い列もnullで書く。
// 日付は "2006-01-02" の文字列で書く。

type animeSnapshotV1 struct {
	Title                 *string                          `json:"title"`
	TitleKana             *string                          `json:"title_kana"`
	TitleRo               *string                          `json:"title_ro"`
	TitleEn               *string                          `json:"title_en"`
	TitleAlter            *string                          `json:"title_alter"`
	TitleAlterRo          *string                          `json:"title_alter_ro"`
	TitleAlterEn          *string                          `json:"title_alter_en"`
	TitleAlterOther       *string                          `json:"title_alter_other"`
	Media                 *string                          `json:"media"`
	ReleaseStatus         *string                          `json:"release_status"`
	Synopsis              *string                          `json:"synopsis"`
	SynopsisEn            *string                          `json:"synopsis_en"`
	SynopsisSource        *string                          `json:"synopsis_source"`
	SynopsisSourceEn      *string                          `json:"synopsis_source_en"`
	Status                string                           `json:"status"`
	ArchiveMessage        *string                          `json:"archive_message"`
	Kind                  string                           `json:"kind"`
	ParentAnimeID         *int64                           `json:"parent_anime_id"`
	Number                *string                          `json:"number"`
	NumberText            *string                          `json:"number_text"`
	SortNumber            *int32                           `json:"sort_number"`
	Standalone            bool                             `json:"standalone"`
	NumberFormatID        *int64                           `json:"number_format_id"`
	EpisodeStartNumber    *string                          `json:"episode_start_number"`
	ExpectedEpisodesCount *int32                           `json:"expected_episodes_count"`
	ExternalIDs           []animeSnapshotExternalIDV1      `json:"external_ids"`
	Links                 []animeSnapshotLinkV1            `json:"links"`
	OfficialAccounts      []animeSnapshotOfficialAccountV1 `json:"official_accounts"`
	Hashtags              []animeSnapshotHashtagV1         `json:"hashtags"`
	Seasons               []animeSnapshotSeasonV1          `json:"seasons"`
	Events                []animeSnapshotEventV1           `json:"events"`
}

type animeSnapshotExternalIDV1 struct {
	Service    string `json:"service"`
	ExternalID string `json:"external_id"`
}

type animeSnapshotLinkV1 struct {
	Kind       string  `json:"kind"`
	Language   string  `json:"language"`
	URL        string  `json:"url"`
	Label      *string `json:"label"`
	LabelEn    *string `json:"label_en"`
	SortNumber int32   `json:"sort_number"`
}

type animeSnapshotOfficialAccountV1 struct {
	Service    string  `json:"service"`
	Account    string  `json:"account"`
	Label      *string `json:"label"`
	LabelEn    *string `json:"label_en"`
	SortNumber int32   `json:"sort_number"`
}

type animeSnapshotHashtagV1 struct {
	Hashtag    string `json:"hashtag"`
	SortNumber int32  `json:"sort_number"`
}

type animeSnapshotSeasonV1 struct {
	Year      int32   `json:"year"`
	Name      *string `json:"name"`
	IsPrimary bool    `json:"is_primary"`
}

type animeSnapshotEventV1 struct {
	Kind          string  `json:"kind"`
	StartedOn     string  `json:"started_on"`
	EndedOn       *string `json:"ended_on"`
	Title         *string `json:"title"`
	TitleEn       *string `json:"title_en"`
	Description   *string `json:"description"`
	DescriptionEn *string `json:"description_en"`
	SortNumber    int32   `json:"sort_number"`
}

func encodeAnimeSnapshot(s model.AnimeSnapshot) (json.RawMessage, error) {
	v := animeSnapshotV1{
		Title:                 s.Title,
		TitleKana:             s.TitleKana,
		TitleRo:               s.TitleRo,
		TitleEn:               s.TitleEn,
		TitleAlter:            s.TitleAlter,
		TitleAlterRo:          s.TitleAlterRo,
		TitleAlterEn:          s.TitleAlterEn,
		TitleAlterOther:       s.TitleAlterOther,
		Media:                 stringPtrFrom(s.Media),
		ReleaseStatus:         stringPtrFrom(s.ReleaseStatus),
		Synopsis:              s.Synopsis,
		SynopsisEn:            s.SynopsisEn,
		SynopsisSource:        s.SynopsisSource,
		SynopsisSourceEn:      s.SynopsisSourceEn,
		Status:                string(s.Status),
		ArchiveMessage:        s.ArchiveMessage,
		Kind:                  string(s.Kind),
		ParentAnimeID:         int64PtrFrom(s.ParentAnimeID),
		Number:                s.Number,
		NumberText:            s.NumberText,
		SortNumber:            s.SortNumber,
		Standalone:            s.Standalone,
		NumberFormatID:        int64PtrFrom(s.NumberFormatID),
		EpisodeStartNumber:    s.EpisodeStartNumber,
		ExpectedEpisodesCount: s.ExpectedEpisodesCount,
		ExternalIDs:           make([]animeSnapshotExternalIDV1, len(s.ExternalIDs)),
		Links:                 make([]animeSnapshotLinkV1, len(s.Links)),
		OfficialAccounts:      make([]animeSnapshotOfficialAccountV1, len(s.OfficialAccounts)),
		Hashtags:              make([]animeSnapshotHashtagV1, len(s.Hashtags)),
		Seasons:               make([]animeSnapshotSeasonV1, len(s.Seasons)),
		Events:                make([]animeSnapshotEventV1, len(s.Events)),
	}
	for i, r := range s.ExternalIDs {
		v.ExternalIDs[i] = animeSnapshotExternalIDV1{Service: string(r.Service), ExternalID: r.ExternalID}
	}
	for i, r := range s.Links {
		v.Links[i] = animeSnapshotLinkV1{Kind: string(r.Kind), Language: string(r.Language), URL: r.URL, Label: r.Label, LabelEn: r.LabelEn, SortNumber: r.SortNumber}
	}
	for i, r := range s.OfficialAccounts {
		v.OfficialAccounts[i] = animeSnapshotOfficialAccountV1{Service: string(r.Service), Account: r.Account, Label: r.Label, LabelEn: r.LabelEn, SortNumber: r.SortNumber}
	}
	for i, r := range s.Hashtags {
		v.Hashtags[i] = animeSnapshotHashtagV1{Hashtag: r.Hashtag, SortNumber: r.SortNumber}
	}
	for i, r := range s.Seasons {
		v.Seasons[i] = animeSnapshotSeasonV1{Year: r.Year, Name: stringPtrFrom(r.Name), IsPrimary: r.IsPrimary}
	}
	for i, r := range s.Events {
		v.Events[i] = animeSnapshotEventV1{
			Kind:          string(r.Kind),
			StartedOn:     r.StartedOn.Format(time.DateOnly),
			EndedOn:       dateStringPtrFrom(r.EndedOn),
			Title:         r.Title,
			TitleEn:       r.TitleEn,
			Description:   r.Description,
			DescriptionEn: r.DescriptionEn,
			SortNumber:    r.SortNumber,
		}
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("作品・エピソードのスナップショットの書き出しに失敗: %w", err)
	}
	return data, nil
}

func decodeAnimeSnapshot(version int16, data json.RawMessage) (model.AnimeSnapshot, error) {
	if version != animeSnapshotVersion {
		return model.AnimeSnapshot{}, fmt.Errorf("作品・エピソードのスナップショットの保存形式の版%dは読み込めません", version)
	}
	var v animeSnapshotV1
	if err := json.Unmarshal(data, &v); err != nil {
		return model.AnimeSnapshot{}, err
	}

	s := model.AnimeSnapshot{
		Title:                 v.Title,
		TitleKana:             v.TitleKana,
		TitleRo:               v.TitleRo,
		TitleEn:               v.TitleEn,
		TitleAlter:            v.TitleAlter,
		TitleAlterRo:          v.TitleAlterRo,
		TitleAlterEn:          v.TitleAlterEn,
		TitleAlterOther:       v.TitleAlterOther,
		Media:                 typedStringPtrFrom[model.AnimeMedia](v.Media),
		ReleaseStatus:         typedStringPtrFrom[model.ReleaseStatus](v.ReleaseStatus),
		Synopsis:              v.Synopsis,
		SynopsisEn:            v.SynopsisEn,
		SynopsisSource:        v.SynopsisSource,
		SynopsisSourceEn:      v.SynopsisSourceEn,
		Status:                model.AnimeStatus(v.Status),
		ArchiveMessage:        v.ArchiveMessage,
		Kind:                  model.AnimeClassificationKind(v.Kind),
		ParentAnimeID:         typedInt64PtrFrom[model.AnimeID](v.ParentAnimeID),
		Number:                v.Number,
		NumberText:            v.NumberText,
		SortNumber:            v.SortNumber,
		Standalone:            v.Standalone,
		NumberFormatID:        typedInt64PtrFrom[model.NumberFormatID](v.NumberFormatID),
		EpisodeStartNumber:    v.EpisodeStartNumber,
		ExpectedEpisodesCount: v.ExpectedEpisodesCount,
	}
	for _, r := range v.ExternalIDs {
		s.ExternalIDs = append(s.ExternalIDs, model.AnimeSnapshotExternalID{Service: model.AnimeExternalService(r.Service), ExternalID: r.ExternalID})
	}
	for _, r := range v.Links {
		s.Links = append(s.Links, model.AnimeSnapshotLink{Kind: model.AnimeLinkKind(r.Kind), Language: model.Language(r.Language), URL: r.URL, Label: r.Label, LabelEn: r.LabelEn, SortNumber: r.SortNumber})
	}
	for _, r := range v.OfficialAccounts {
		s.OfficialAccounts = append(s.OfficialAccounts, model.AnimeSnapshotOfficialAccount{Service: model.AnimeAccountService(r.Service), Account: r.Account, Label: r.Label, LabelEn: r.LabelEn, SortNumber: r.SortNumber})
	}
	for _, r := range v.Hashtags {
		s.Hashtags = append(s.Hashtags, model.AnimeSnapshotHashtag{Hashtag: r.Hashtag, SortNumber: r.SortNumber})
	}
	for _, r := range v.Seasons {
		s.Seasons = append(s.Seasons, model.AnimeSnapshotSeason{Year: r.Year, Name: typedStringPtrFrom[model.SeasonName](r.Name), IsPrimary: r.IsPrimary})
	}
	for _, r := range v.Events {
		startedOn, err := time.Parse(time.DateOnly, r.StartedOn)
		if err != nil {
			return model.AnimeSnapshot{}, fmt.Errorf("イベントの開始日の読み込みに失敗: %w", err)
		}
		endedOn, err := dateFromStringPtr(r.EndedOn)
		if err != nil {
			return model.AnimeSnapshot{}, fmt.Errorf("イベントの終了日の読み込みに失敗: %w", err)
		}
		s.Events = append(s.Events, model.AnimeSnapshotEvent{
			Kind:          model.AnimeEventKind(r.Kind),
			StartedOn:     startedOn,
			EndedOn:       endedOn,
			Title:         r.Title,
			TitleEn:       r.TitleEn,
			Description:   r.Description,
			DescriptionEn: r.DescriptionEn,
			SortNumber:    r.SortNumber,
		})
	}
	return s, nil
}

// 以下は、スナップショットの型と保存形式の基底型を相互に変換する。

func stringPtrFrom[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func int64PtrFrom[T ~int64](v *T) *int64 {
	if v == nil {
		return nil
	}
	i := int64(*v)
	return &i
}

func typedStringPtrFrom[T ~string](v *string) *T {
	if v == nil {
		return nil
	}
	t := T(*v)
	return &t
}

func typedInt64PtrFrom[T ~int64](v *int64) *T {
	if v == nil {
		return nil
	}
	t := T(*v)
	return &t
}

func dateStringPtrFrom(v *time.Time) *string {
	if v == nil {
		return nil
	}
	s := v.Format(time.DateOnly)
	return &s
}

func dateFromStringPtr(v *string) (*time.Time, error) {
	if v == nil {
		return nil, nil
	}
	t, err := time.Parse(time.DateOnly, *v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListForDBEditsは、指定した操作に属する作品・エピソードの版を、IDの昇順に返します。各版のPreviousに起点の版を
// 含めた1つ前の版を、Workに今の作品 (エピソードの版では今の親作品) のIDとタイトルを、Episodeに今のエピソードの
// IDを入れます。削除した作品・エピソードの版は返しません。
// dbEditIDsが空のときはクエリを発行しません。
func (r *AnimeRevisionRepository) ListForDBEdits(ctx context.Context, dbEditIDs []model.DBEditID) ([]*model.AnimeRevision, error) {
	if len(dbEditIDs) == 0 {
		return nil, nil
	}
	ids := make([]int64, len(dbEditIDs))
	for i, id := range dbEditIDs {
		ids[i] = int64(id)
	}
	rows, err := r.queries.ListAnimeRevisionsForDBEdits(ctx, ids)
	if err != nil {
		return nil, err
	}

	revisions := make([]*model.AnimeRevision, len(rows))
	for i, row := range rows {
		snapshot, err := decodeAnimeSnapshot(row.SnapshotVersion, row.Snapshot)
		if err != nil {
			return nil, fmt.Errorf("作品・エピソードの版%dのスナップショットの読み込みに失敗: %w", row.ID, err)
		}
		dbEditID := model.DBEditID(row.DbEditID)
		revision := &model.AnimeRevision{
			ID:        model.AnimeRevisionID(row.ID),
			DBEditID:  &dbEditID,
			AnimeID:   model.AnimeID(row.AnimeID),
			Snapshot:  snapshot,
			CreatedAt: row.CreatedAt,
		}
		if row.PreviousSnapshotVersion != 0 {
			previous, err := decodeAnimeSnapshot(row.PreviousSnapshotVersion, row.PreviousSnapshot)
			if err != nil {
				return nil, fmt.Errorf("作品・エピソードの版%dの1つ前の版の読み込みに失敗: %w", row.ID, err)
			}
			revision.Previous = &previous
		}
		if row.WorkID != 0 {
			revision.Work = &model.Work{ID: model.WorkID(row.WorkID), Title: row.WorkTitle}
		}
		if row.EpisodeID != 0 {
			revision.Episode = &model.Episode{ID: model.EpisodeID(row.EpisodeID), WorkID: model.WorkID(row.WorkID)}
		}
		revisions[i] = revision
	}
	return revisions, nil
}
