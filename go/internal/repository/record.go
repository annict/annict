package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// RecordRepository handles data access for the records table.
//
// Aggregate-only access is intentionally exposed without a full Record model:
// the tracking-heatmap migration only needs daily counts, and introducing a
// Record model here would couple this work to a broader CRUD migration.
// A model.Record will be added when records-CRUD endpoints are migrated.
//
// [Ja] RecordRepository は records テーブルへのデータアクセスを担う。
// 視聴記録ヒートマップ移行では日次集計しか必要としないため、本タスクでは
// Record モデルを導入せず集計専用 API のみを公開する。Record モデルは
// 視聴記録の CRUD を Go 版に移行する別タスクで追加する。
type RecordRepository struct {
	queries *query.Queries
}

// NewRecordRepository constructs a RecordRepository.
//
// [Ja] NewRecordRepository は RecordRepository を生成する。
func NewRecordRepository(queries *query.Queries) *RecordRepository {
	return &RecordRepository{queries: queries}
}

// WithTx returns a new RecordRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい RecordRepository を返す。
func (r *RecordRepository) WithTx(tx *sql.Tx) *RecordRepository {
	return &RecordRepository{queries: r.queries.WithTx(tx)}
}

// DailyRecordCount represents the number of records logged on a given day,
// where the day is expressed in the caller-supplied time zone.
//
// [Ja] DailyRecordCount は呼び出し元の指定タイムゾーンにおける 1 日あたりの
// 視聴記録数を表す。
type DailyRecordCount struct {
	// Day is midnight of the bucketed day in the requested time zone.
	// PostgreSQL's `date` type is decoded by lib/pq as time.Time at 00:00
	// UTC; callers should format the value via Day.Format("2006-01-02")
	// instead of relying on the location.
	//
	// [Ja] 集計対象の日 (指定タイムゾーンにおける 00:00) を表す。
	// PostgreSQL の date 型は lib/pq によって UTC 0 時の time.Time として
	// デコードされるため、日付文字列が必要な場合は Day.Format("2006-01-02")
	// を用いてフォーマットすること。
	Day time.Time
	// Count is the number of non-deleted records logged on Day.
	//
	// [Ja] Day に記録された論理削除されていないレコードの件数。
	Count int64
}

// AggregateDailyCountsByUserID returns a per-day count of records for the
// given user from dateFromUTC onward. timeZone determines which UTC day a
// record falls into (e.g. a record at 2026-01-01 23:00 UTC is counted in
// 2026-01-02 when timeZone is "Asia/Tokyo"). Days without any records are
// omitted from the result; UseCase code is responsible for zero-filling.
//
// [Ja] 指定ユーザーについて dateFromUTC 以降の視聴記録を日次で集計する。
// timeZone は記録がどの日に属するかを決定する (例: 2026-01-01 23:00 UTC の
// 記録は timeZone="Asia/Tokyo" の場合に 2026-01-02 として集計される)。
// 記録のない日は結果に含まれないため、連続した日付配列が必要な場合は
// UseCase 側で 0 埋めを行う。
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
