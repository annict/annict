package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeOfficialAccountRepositoryはanime_official_accountsテーブル (animeの公式
// ソーシャルアカウント。例: Xアカウント) へのデータアクセスを担う。
type AnimeOfficialAccountRepository struct {
	queries *query.Queries
}

// NewAnimeOfficialAccountRepositoryはAnimeOfficialAccountRepositoryを生成する。
func NewAnimeOfficialAccountRepository(queries *query.Queries) *AnimeOfficialAccountRepository {
	return &AnimeOfficialAccountRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいAnimeOfficialAccountRepositoryを返す。
func (r *AnimeOfficialAccountRepository) WithTx(tx *sql.Tx) *AnimeOfficialAccountRepository {
	return &AnimeOfficialAccountRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeOfficialAccountParamsは公式アカウント作成時の属性を保持する。id /
// label / label_en / sort_number / タイムスタンプはデータベースが採番する (label /
// label_enはNULL、sort_numberは0が既定値)。worksはaccountのみをsourceするため、
// 自然キー (anime_id, service) とハンドルだけを持つ。(anime_id, service) がUNIQUE
// インデックスで守られる自然キー。
type CreateAnimeOfficialAccountParams struct {
	AnimeID model.AnimeID
	Service model.AnimeAccountService
	Account string
}

// Createは新しい公式アカウントを挿入し、作成された行を返す。
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

// UpdateAnimeOfficialAccountParamsは主キーで特定した公式アカウントの更新時の属性を
// 保持する。可変なのはaccountのみで、自然キー (anime_id, service) は固定、sourceしない
// label / label_en / sort_numberは保全するため、ハンドルが変わったサービスは削除と再作成
// ではなくその場で更新する。
type UpdateAnimeOfficialAccountParams struct {
	ID      model.AnimeOfficialAccountID
	Account string
}

// Updateは指定行のaccountを上書きする。
func (r *AnimeOfficialAccountRepository) Update(ctx context.Context, params UpdateAnimeOfficialAccountParams) error {
	return r.queries.UpdateAnimeOfficialAccount(ctx, query.UpdateAnimeOfficialAccountParams{
		ID:      int64(params.ID),
		Account: params.Account,
	})
}

// Deleteは指定主キーの公式アカウントを削除する。
func (r *AnimeOfficialAccountRepository) Delete(ctx context.Context, id model.AnimeOfficialAccountID) error {
	return r.queries.DeleteAnimeOfficialAccount(ctx, int64(id))
}

// ListByAnimeIDsは指定anime ID群の公式アカウントを (anime_id, service) 順で
// ロードする。フェーズ2の別表リコンシリエーションが、anime解決済みworksの1ページぶんの
// 既存行をN回のanime単位ルックアップではなく1クエリで一括取得するために使う。空入力では
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

// toAnimeOfficialAccountModelはqueryの行をドメインモデルに変換する。
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
