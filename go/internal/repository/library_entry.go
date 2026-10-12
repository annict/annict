package repository

import (
	"context"
	"database/sql"

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
