package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeOfficialAccountRepository handles data access for the anime_official_accounts
// table (an anime's official social accounts, e.g. its X account).
//
// [Ja] AnimeOfficialAccountRepository は anime_official_accounts テーブル (anime の公式
// ソーシャルアカウント。例: X アカウント) へのデータアクセスを担う。
type AnimeOfficialAccountRepository struct {
	queries *query.Queries
}

// NewAnimeOfficialAccountRepository constructs an AnimeOfficialAccountRepository.
//
// [Ja] NewAnimeOfficialAccountRepository は AnimeOfficialAccountRepository を生成する。
func NewAnimeOfficialAccountRepository(queries *query.Queries) *AnimeOfficialAccountRepository {
	return &AnimeOfficialAccountRepository{queries: queries}
}

// WithTx returns a new AnimeOfficialAccountRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい AnimeOfficialAccountRepository を返す。
func (r *AnimeOfficialAccountRepository) WithTx(tx *sql.Tx) *AnimeOfficialAccountRepository {
	return &AnimeOfficialAccountRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeOfficialAccountParams holds the attributes for creating an official
// account. The id, label / label_en, sort_number and timestamps are assigned by the
// database (label / label_en default to NULL, sort_number to 0); works source only
// account, so this carries just the natural key (anime_id, service) and the handle.
// (anime_id, service) is the natural key enforced by a UNIQUE index.
//
// [Ja] CreateAnimeOfficialAccountParams は公式アカウント作成時の属性を保持する。id /
// label / label_en / sort_number / タイムスタンプはデータベースが採番する (label /
// label_en は NULL、sort_number は 0 が既定値)。works は account のみを source するため、
// 自然キー (anime_id, service) とハンドルだけを持つ。(anime_id, service) が UNIQUE
// インデックスで守られる自然キー。
type CreateAnimeOfficialAccountParams struct {
	AnimeID model.AnimeID
	Service model.AnimeAccountService
	Account string
}

// Create inserts a new official account and returns the created row.
//
// [Ja] Create は新しい公式アカウントを挿入し、作成された行を返す。
func (r *AnimeOfficialAccountRepository) Create(ctx context.Context, params CreateAnimeOfficialAccountParams) (*model.AnimeOfficialAccount, error) {
	row, err := r.queries.CreateAnimeOfficialAccount(ctx, query.CreateAnimeOfficialAccountParams{
		AnimeID: int64(params.AnimeID),
		Service: query.AnimeAccountService(params.Service),
		Account: params.Account,
	})
	if err != nil {
		return nil, err
	}
	account := toAnimeOfficialAccountModel(row)
	return &account, nil
}

// UpdateAnimeOfficialAccountParams holds the attributes for updating an official
// account, identified by its primary key. Only account is mutable; the natural key
// (anime_id, service) is fixed and the non-sourced label / label_en / sort_number are
// preserved, so a service whose handle changed is updated in place rather than deleted
// and re-created.
//
// [Ja] UpdateAnimeOfficialAccountParams は主キーで特定した公式アカウントの更新時の属性を
// 保持する。可変なのは account のみで、自然キー (anime_id, service) は固定、source しない
// label / label_en / sort_number は保全するため、ハンドルが変わったサービスは削除と再作成
// ではなくその場で更新する。
type UpdateAnimeOfficialAccountParams struct {
	ID      model.AnimeOfficialAccountID
	Account string
}

// Update overwrites the account of the identified row.
//
// [Ja] Update は指定行の account を上書きする。
func (r *AnimeOfficialAccountRepository) Update(ctx context.Context, params UpdateAnimeOfficialAccountParams) error {
	return r.queries.UpdateAnimeOfficialAccount(ctx, query.UpdateAnimeOfficialAccountParams{
		ID:      int64(params.ID),
		Account: params.Account,
	})
}

// Delete removes the official account with the given primary key.
//
// [Ja] Delete は指定主キーの公式アカウントを削除する。
func (r *AnimeOfficialAccountRepository) Delete(ctx context.Context, id model.AnimeOfficialAccountID) error {
	return r.queries.DeleteAnimeOfficialAccount(ctx, int64(id))
}

// ListByAnimeIDs loads the official accounts for the given anime IDs, ordered by
// (anime_id, service). It is used by the phase 2 satellite reconciliation to
// batch-fetch the existing rows for a page of anime-resolved works in one query
// instead of N per-anime lookups. An empty input returns an empty slice without
// querying.
//
// [Ja] ListByAnimeIDs は指定 anime ID 群の公式アカウントを (anime_id, service) 順で
// ロードする。フェーズ 2 の別表リコンシリエーションが、anime 解決済み works の 1 ページぶんの
// 既存行を N 回の anime 単位ルックアップではなく 1 クエリで一括取得するために使う。空入力では
// クエリせず空スライスを返す。
func (r *AnimeOfficialAccountRepository) ListByAnimeIDs(ctx context.Context, animeIDs []model.AnimeID) ([]*model.AnimeOfficialAccount, error) {
	if len(animeIDs) == 0 {
		return []*model.AnimeOfficialAccount{}, nil
	}

	rawIDs := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		rawIDs[i] = int64(id)
	}

	rows, err := r.queries.ListAnimeOfficialAccountsByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	accounts := make([]*model.AnimeOfficialAccount, len(rows))
	for i, row := range rows {
		account := toAnimeOfficialAccountModel(row)
		accounts[i] = &account
	}
	return accounts, nil
}

// toAnimeOfficialAccountModel converts a query row into the domain model.
//
// [Ja] toAnimeOfficialAccountModel は query の行をドメインモデルに変換する。
func toAnimeOfficialAccountModel(row query.AnimeOfficialAccount) model.AnimeOfficialAccount {
	account := model.AnimeOfficialAccount{
		ID:         model.AnimeOfficialAccountID(row.ID),
		AnimeID:    model.AnimeID(row.AnimeID),
		Service:    model.AnimeAccountService(row.Service),
		Account:    row.Account,
		SortNumber: row.SortNumber,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
	if row.Label.Valid {
		label := row.Label.String
		account.Label = &label
	}
	if row.LabelEn.Valid {
		labelEn := row.LabelEn.String
		account.LabelEn = &labelEn
	}
	return account
}
