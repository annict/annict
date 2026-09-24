package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// RecordRepositoryはrecordsテーブルへのデータアクセスを担う。
// 視聴記録ヒートマップ移行では日次集計しか必要としないため、本タスクでは
// Recordモデルを導入せず集計専用APIのみを公開する。Recordモデルは
// 視聴記録のCRUDをGo版に移行する別タスクで追加する。
type RecordRepository struct {
	queries *query.Queries
}

// NewRecordRepositoryはRecordRepositoryを生成する。
func NewRecordRepository(queries *query.Queries) *RecordRepository {
	return &RecordRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRecordRepositoryを返す。
func (r *RecordRepository) WithTx(tx *sql.Tx) *RecordRepository {
	return &RecordRepository{queries: r.queries.WithTx(tx)}
}

// DailyRecordCountは呼び出し元の指定タイムゾーンにおける1日あたりの
// 視聴記録数を表す。
type DailyRecordCount struct {
	// 集計対象の日 (指定タイムゾーンにおける00:00) を表す。
	// PostgreSQLのdate型はlib/pqによってUTC 0時のtime.Timeとして
	// デコードされるため、日付文字列が必要な場合はDay.Format("2006-01-02")
	// を用いてフォーマットすること。
	Day time.Time
	// Dayに記録された論理削除されていないレコードの件数。
	Count int64
}

// AggregateDailyCountsByUserIDは指定ユーザーについてdateFromUTC以降の視聴記録を日次で集計する。
// timeZoneは記録がどの日に属するかを決定する (例: 2026-01-01 23:00 UTCの
// 記録はtimeZone="Asia/Tokyo" の場合に2026-01-02として集計される)。
// 記録のない日は結果に含まれないため、連続した日付配列が必要な場合は
// UseCase側で0埋めを行う。
func (r *RecordRepository) AggregateDailyCountsByUserID(
	ctx context.Context,
	userID model.UserID,
	dateFromUTC time.Time,
	timeZone string,
) ([]*DailyRecordCount, error) {
	rows, err := r.queries.AggregateDailyRecordCountsByUserID(ctx, query.AggregateDailyRecordCountsByUserIDParams{
		UserID:   int64(userID),
		TimeZone: timeZone,
		DateFrom: dateFromUTC,
	})
	if err != nil {
		return nil, fmt.Errorf("日次視聴記録数の集計に失敗: %w", err)
	}

	counts := make([]*DailyRecordCount, len(rows))
	for i, row := range rows {
		counts[i] = &DailyRecordCount{
			Day:   row.Day,
			Count: row.Count,
		}
	}
	return counts, nil
}
