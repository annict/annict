package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// personSnapshotVersionは、person_revisions.snapshotに書く保存形式の版。
const personSnapshotVersion = 1

// PersonRevisionRepositoryは人物の版 (person_revisions) のデータアクセスを担当します。
// スナップショットとjsonbの保存形式の変換は、このリポジトリに閉じ込めます。
type PersonRevisionRepository struct {
	queries *query.Queries
}

// NewPersonRevisionRepositoryはPersonRevisionRepositoryを作成します
func NewPersonRevisionRepository(queries *query.Queries) *PersonRevisionRepository {
	return &PersonRevisionRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *PersonRevisionRepository) WithTx(tx *sql.Tx) *PersonRevisionRepository {
	return &PersonRevisionRepository{queries: r.queries.WithTx(tx)}
}

// Createは操作dbEditIDに属する版を作成します。
func (r *PersonRevisionRepository) Create(ctx context.Context, dbEditID model.DBEditID, personID model.PersonID, snapshot model.PersonSnapshot) error {
	data, err := encodePersonSnapshot(snapshot)
	if err != nil {
		return err
	}
	return r.queries.CreatePersonRevision(ctx, query.CreatePersonRevisionParams{
		DbEditID:        sql.NullInt64{Int64: int64(dbEditID), Valid: true},
		PersonID:        int64(personID),
		Snapshot:        data,
		SnapshotVersion: personSnapshotVersion,
	})
}

// CreateInitialは起点の版を作成します。同じ人物の起点の版がすでにあるときは何もしません。
func (r *PersonRevisionRepository) CreateInitial(ctx context.Context, personID model.PersonID, snapshot model.PersonSnapshot) error {
	data, err := encodePersonSnapshot(snapshot)
	if err != nil {
		return err
	}
	return r.queries.CreateInitialPersonRevision(ctx, query.CreateInitialPersonRevisionParams{
		PersonID:        int64(personID),
		Snapshot:        data,
		SnapshotVersion: personSnapshotVersion,
	})
}

// ListPersonIDsWithRevisionsはpersonIDsのうち、版を1つ以上持つ人物のIDの集合を返します。
// 空の入力ではクエリせずに空の集合を返します。
func (r *PersonRevisionRepository) ListPersonIDsWithRevisions(ctx context.Context, personIDs []model.PersonID) (map[model.PersonID]bool, error) {
	result := make(map[model.PersonID]bool, len(personIDs))
	if len(personIDs) == 0 {
		return result, nil
	}

	ids := make([]int64, len(personIDs))
	for i, id := range personIDs {
		ids[i] = int64(id)
	}
	rows, err := r.queries.ListPersonIDsWithRevisions(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, id := range rows {
		result[model.PersonID(id)] = true
	}
	return result, nil
}

// ListByPersonIDは人物の版を、起点の版を含めてIDの昇順に返します。
func (r *PersonRevisionRepository) ListByPersonID(ctx context.Context, personID model.PersonID) ([]*model.PersonRevision, error) {
	rows, err := r.queries.ListPersonRevisionsByPersonID(ctx, int64(personID))
	if err != nil {
		return nil, err
	}

	revisions := make([]*model.PersonRevision, len(rows))
	for i, row := range rows {
		snapshot, err := decodePersonSnapshot(row.SnapshotVersion, row.Snapshot)
		if err != nil {
			return nil, fmt.Errorf("人物の版%dのスナップショットの読み込みに失敗: %w", row.ID, err)
		}
		revision := &model.PersonRevision{
			ID:        model.PersonRevisionID(row.ID),
			PersonID:  model.PersonID(row.PersonID),
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

// PersonRevisionHistoryParamsは人物の履歴の1ページの指定。IncludeDelistedが偽のときは、掲載停止・掲載再開の
// 操作の版と、掲載停止中に記録した版を除く。
type PersonRevisionHistoryParams struct {
	PersonID        model.PersonID
	IncludeDelisted bool
	Page            int32
	PerPage         int32
}

// ListHistoryByPersonIDは人物の操作の版を、新しい順に1ページ分返します。起点の版は並べず、各版のPreviousに、
// 起点の版と隠す版を含めた1つ前の版のスナップショットを入れます。
func (r *PersonRevisionRepository) ListHistoryByPersonID(ctx context.Context, params PersonRevisionHistoryParams) ([]*model.PersonRevision, error) {
	// 乗算の前に幅を広げる (AnimeRevisionRepository.ListHistoryByAnimeIDと同じ)。
	offset := int64(params.Page-1) * int64(params.PerPage)

	rows, err := r.queries.ListPersonRevisionHistory(ctx, query.ListPersonRevisionHistoryParams{
		PersonID:        int64(params.PersonID),
		IncludeDelisted: params.IncludeDelisted,
		PerPage:         params.PerPage,
		PageOffset:      offset,
	})
	if err != nil {
		return nil, err
	}

	revisions := make([]*model.PersonRevision, len(rows))
	for i, row := range rows {
		snapshot, err := decodePersonSnapshot(row.SnapshotVersion, row.Snapshot)
		if err != nil {
			return nil, fmt.Errorf("人物の版%dのスナップショットの読み込みに失敗: %w", row.ID, err)
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
		revision := &model.PersonRevision{
			ID:        model.PersonRevisionID(row.ID),
			DBEditID:  &dbEditID,
			PersonID:  model.PersonID(row.PersonID),
			Snapshot:  snapshot,
			CreatedAt: row.CreatedAt,
			DBEdit:    edit,
		}
		if row.PreviousSnapshotVersion != 0 {
			previous, err := decodePersonSnapshot(row.PreviousSnapshotVersion, row.PreviousSnapshot)
			if err != nil {
				return nil, fmt.Errorf("人物の版%dの1つ前の版の読み込みに失敗: %w", row.ID, err)
			}
			revision.Previous = &previous
		}
		revisions[i] = revision
	}
	return revisions, nil
}

// ListForDBEditsは、指定した操作に属する人物の版を、IDの昇順に返します。各版のPreviousに、起点の版と隠す版を
// 含めた1つ前の版を入れます。削除した人物の版は返さず、includeDelistedが偽のときは人物の掲載停止にかかわる版を
// 返しません (ListHistoryByPersonIDと同じ)。dbEditIDsが空のときはクエリを発行しません。
func (r *PersonRevisionRepository) ListForDBEdits(ctx context.Context, dbEditIDs []model.DBEditID, includeDelisted bool) ([]*model.PersonRevision, error) {
	if len(dbEditIDs) == 0 {
		return nil, nil
	}
	ids := make([]int64, len(dbEditIDs))
	for i, id := range dbEditIDs {
		ids[i] = int64(id)
	}
	rows, err := r.queries.ListPersonRevisionsForDBEdits(ctx, query.ListPersonRevisionsForDBEditsParams{
		DbEditIds:       ids,
		IncludeDelisted: includeDelisted,
	})
	if err != nil {
		return nil, err
	}

	revisions := make([]*model.PersonRevision, len(rows))
	for i, row := range rows {
		snapshot, err := decodePersonSnapshot(row.SnapshotVersion, row.Snapshot)
		if err != nil {
			return nil, fmt.Errorf("人物の版%dのスナップショットの読み込みに失敗: %w", row.ID, err)
		}
		dbEditID := model.DBEditID(row.DbEditID)
		revision := &model.PersonRevision{
			ID:        model.PersonRevisionID(row.ID),
			DBEditID:  &dbEditID,
			PersonID:  model.PersonID(row.PersonID),
			Snapshot:  snapshot,
			CreatedAt: row.CreatedAt,
		}
		if row.PreviousSnapshotVersion != 0 {
			previous, err := decodePersonSnapshot(row.PreviousSnapshotVersion, row.PreviousSnapshot)
			if err != nil {
				return nil, fmt.Errorf("人物の版%dの1つ前の版の読み込みに失敗: %w", row.ID, err)
			}
			revision.Previous = &previous
		}
		revisions[i] = revision
	}
	return revisions, nil
}

// CountHistoryByPersonIDは、ListHistoryByPersonIDが並べる版の数を返します。
func (r *PersonRevisionRepository) CountHistoryByPersonID(ctx context.Context, personID model.PersonID, includeDelisted bool) (int64, error) {
	return r.queries.CountPersonRevisionHistory(ctx, query.CountPersonRevisionHistoryParams{
		PersonID:        int64(personID),
		IncludeDelisted: includeDelisted,
	})
}

// personSnapshotV1はsnapshot_version = 1の保存形式。キーはpeopleの列名にそろえる。
type personSnapshotV1 struct {
	Name              string  `json:"name"`
	NameKana          string  `json:"name_kana"`
	NameEn            string  `json:"name_en"`
	Nickname          *string `json:"nickname"`
	NicknameEn        string  `json:"nickname_en"`
	Gender            *string `json:"gender"`
	BloodType         *string `json:"blood_type"`
	PrefectureID      *int64  `json:"prefecture_id"`
	Birthday          *string `json:"birthday"`
	Height            *int32  `json:"height"`
	URL               *string `json:"url"`
	URLEn             string  `json:"url_en"`
	WikipediaURL      *string `json:"wikipedia_url"`
	WikipediaURLEn    string  `json:"wikipedia_url_en"`
	TwitterUsername   *string `json:"twitter_username"`
	TwitterUsernameEn string  `json:"twitter_username_en"`
	Status            string  `json:"status"`
	ArchiveMessage    *string `json:"archive_message"`
}

func encodePersonSnapshot(s model.PersonSnapshot) (json.RawMessage, error) {
	data, err := json.Marshal(personSnapshotV1{
		Name:              s.Name,
		NameKana:          s.NameKana,
		NameEn:            s.NameEn,
		Nickname:          s.Nickname,
		NicknameEn:        s.NicknameEn,
		Gender:            stringPtrFrom(s.Gender),
		BloodType:         stringPtrFrom(s.BloodType),
		PrefectureID:      int64PtrFrom(s.PrefectureID),
		Birthday:          dateStringPtrFrom(s.Birthday),
		Height:            s.Height,
		URL:               s.URL,
		URLEn:             s.URLEn,
		WikipediaURL:      s.WikipediaURL,
		WikipediaURLEn:    s.WikipediaURLEn,
		TwitterUsername:   s.TwitterUsername,
		TwitterUsernameEn: s.TwitterUsernameEn,
		Status:            string(s.Status),
		ArchiveMessage:    s.ArchiveMessage,
	})
	if err != nil {
		return nil, fmt.Errorf("人物のスナップショットの書き出しに失敗: %w", err)
	}
	return data, nil
}

func decodePersonSnapshot(version int16, data json.RawMessage) (model.PersonSnapshot, error) {
	if version != personSnapshotVersion {
		return model.PersonSnapshot{}, fmt.Errorf("人物のスナップショットの保存形式の版%dは読み込めません", version)
	}
	var v personSnapshotV1
	if err := json.Unmarshal(data, &v); err != nil {
		return model.PersonSnapshot{}, err
	}
	birthday, err := dateFromStringPtr(v.Birthday)
	if err != nil {
		return model.PersonSnapshot{}, fmt.Errorf("誕生日の読み込みに失敗: %w", err)
	}

	return model.PersonSnapshot{
		Name:              v.Name,
		NameKana:          v.NameKana,
		NameEn:            v.NameEn,
		Nickname:          v.Nickname,
		NicknameEn:        v.NicknameEn,
		Gender:            typedStringPtrFrom[model.PersonGender](v.Gender),
		BloodType:         typedStringPtrFrom[model.PersonBloodType](v.BloodType),
		PrefectureID:      typedInt64PtrFrom[model.PrefectureID](v.PrefectureID),
		Birthday:          birthday,
		Height:            v.Height,
		URL:               v.URL,
		URLEn:             v.URLEn,
		WikipediaURL:      v.WikipediaURL,
		WikipediaURLEn:    v.WikipediaURLEn,
		TwitterUsername:   v.TwitterUsername,
		TwitterUsernameEn: v.TwitterUsernameEn,
		Status:            model.PersonStatus(v.Status),
		ArchiveMessage:    v.ArchiveMessage,
	}, nil
}
