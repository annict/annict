package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// LibraryEntryRepositoryはライブラリエントリ関連のデータアクセスを担当します
type LibraryEntryRepository struct {
	queries *query.Queries
}

// NewLibraryEntryRepositoryはLibraryEntryRepositoryを作成します
func NewLibraryEntryRepository(queries *query.Queries) *LibraryEntryRepository {
	return &LibraryEntryRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *LibraryEntryRepository) WithTx(tx *sql.Tx) *LibraryEntryRepository {
	return &LibraryEntryRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのライブラリエントリを最大limit件削除し、削除した件数を返します。
// 見たい・見てる・見たのエントリの分だけ、作品の視聴者数 (works.watchers_count) を同じ文の中で減算します。
func (r *LibraryEntryRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteLibraryEntriesByUserID(ctx, query.DeleteLibraryEntriesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// UpdateNextResourcesは「見てる」のライブラリのうち、updatedFrom以降に更新され、次のエピソードが
// 無いか、番組を選んでいて次の放送枠が無いものを、afterIDより大きいidの順に最大limit件読み、
// 次のエピソード・次の放送枠を1つの文で書き換えます。
// 読んだ行の最後のid (読んだ行が無いときは0) と、書き換えた件数を返します。
// 値の変わらない行はupdated_atも進めません。
func (r *LibraryEntryRepository) UpdateNextResources(ctx context.Context, afterID int64, updatedFrom, now time.Time, limit int32) (int64, int64, error) {
	row, err := r.queries.UpdateLibraryEntryNextResources(ctx, query.UpdateLibraryEntryNextResourcesParams{
		AfterID:     afterID,
		UpdatedFrom: updatedFrom.UTC(),
		BatchSize:   limit,
		Now:         now.UTC(),
	})
	if err != nil {
		return 0, 0, err
	}
	return row.LastID, row.UpdatedCount, nil
}
