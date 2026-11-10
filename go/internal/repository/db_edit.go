package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// DBEditRepositoryはAnnict DBの編集履歴の操作 (db_edits) のデータアクセスを担当します
type DBEditRepository struct {
	queries *query.Queries
}

// NewDBEditRepositoryはDBEditRepositoryを作成します
func NewDBEditRepository(queries *query.Queries) *DBEditRepository {
	return &DBEditRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *DBEditRepository) WithTx(tx *sql.Tx) *DBEditRepository {
	return &DBEditRepository{queries: r.queries.WithTx(tx)}
}

// CreateDBEditParamsは操作の作成に使うパラメータを保持します。
type CreateDBEditParams struct {
	Actor model.DBEditActor
	// UserIDは編集した編集者です。ActorがDBEditActorSystemのときは空にします
	// (db_editsのCHECK制約が、システムの書き込みに編集者を持たせません)。
	UserID *model.UserID
	Action model.DBEditAction
}

// Createは操作を1件作成し、作成した操作を返します。
func (r *DBEditRepository) Create(ctx context.Context, params CreateDBEditParams) (*model.DBEdit, error) {
	var userID sql.NullInt64
	if params.UserID != nil {
		userID = sql.NullInt64{Int64: int64(*params.UserID), Valid: true}
	}

	row, err := r.queries.CreateDBEdit(ctx, query.CreateDBEditParams{
		Actor:  query.DbEditActor(params.Actor),
		UserID: userID,
		Action: query.DbEditAction(params.Action),
	})
	if err != nil {
		return nil, err
	}
	return toDBEdit(row), nil
}

// DBEditListParamsは操作の一覧の取得条件を保持します。IncludeDelistedが偽のときは、人物の掲載停止にかかわる
// 版 (掲載停止中の人物の版と、掲載停止・掲載再開の操作の版と、掲載停止中に記録した版) を見せない版として扱います。
type DBEditListParams struct {
	IncludeDelisted bool
	Page            int32
	PerPage         int32
}

// Listは、見せる版を1つ以上持つ操作を、IDの降順に1ページ分取得します。見せる版は、削除していない作品・
// エピソード・人物の版で、IncludeDelistedが偽のときは人物の掲載停止にかかわる版を除きます。
func (r *DBEditRepository) List(ctx context.Context, params DBEditListParams) ([]*model.DBEdit, error) {
	// 乗算の前に幅を広げる。ページ番号がint32の範囲で大きいとint32同士の積が負に折り返し、
	// PostgreSQLがそのOFFSETを拒否するため (PersonRepository.ListForDBと同じ)。
	offset := int64(params.Page-1) * int64(params.PerPage)

	rows, err := r.queries.ListDBEdits(ctx, query.ListDBEditsParams{
		IncludeDelisted: params.IncludeDelisted,
		PerPage:         params.PerPage,
		PageOffset:      offset,
	})
	if err != nil {
		return nil, err
	}

	edits := make([]*model.DBEdit, len(rows))
	for i, row := range rows {
		edit := toDBEdit(query.DbEdit{ID: row.ID, Actor: row.Actor, UserID: row.UserID, Action: row.Action, CreatedAt: row.CreatedAt})
		if edit.UserID != nil {
			edit.User = &model.User{ID: *edit.UserID, Username: row.Username, ProfileImageData: row.ProfileImageData}
		}
		edits[i] = edit
	}
	return edits, nil
}

// Countは、Listが並べる操作の数を返します。
func (r *DBEditRepository) Count(ctx context.Context, includeDelisted bool) (int64, error) {
	return r.queries.CountDBEdits(ctx, includeDelisted)
}

// DetachFromUserはユーザーが行った操作を最大limit件、編集者との紐づきを外し、外した件数を返します。
func (r *DBEditRepository) DetachFromUser(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DetachDBEditsFromUser(ctx, query.DetachDBEditsFromUserParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

func toDBEdit(row query.DbEdit) *model.DBEdit {
	edit := &model.DBEdit{
		ID:        model.DBEditID(row.ID),
		Actor:     model.DBEditActor(row.Actor),
		Action:    model.DBEditAction(row.Action),
		CreatedAt: row.CreatedAt,
	}
	if row.UserID.Valid {
		userID := model.UserID(row.UserID.Int64)
		edit.UserID = &userID
	}
	return edit
}
