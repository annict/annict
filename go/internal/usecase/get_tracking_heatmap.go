package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// trackingHeatmapDays is the size of the heatmap window. The starting day
// is computed as (today - trackingHeatmapDays) snapped back to that week's
// Sunday, mirroring the Rails implementation.
//
// [Ja] trackingHeatmapDays はヒートマップの表示日数。開始日は
// (今日 - trackingHeatmapDays) をその週の日曜日に丸めた日となり、
// Rails 版と同じ挙動になる。
const trackingHeatmapDays = 150

// GetTrackingHeatmapUsecase aggregates a user's daily watch-record counts
// over the trailing trackingHeatmapDays window and prepares them for
// rendering as a heatmap fragment.
//
// [Ja] GetTrackingHeatmapUsecase は直近 trackingHeatmapDays 日分の視聴記録
// 数を日次集計し、ヒートマップフラグメント描画用のセル配列に整形する。
type GetTrackingHeatmapUsecase struct {
	userRepo   *repository.UserRepository
	recordRepo *repository.RecordRepository
}

// NewGetTrackingHeatmapUsecase constructs the use case.
// [Ja] NewGetTrackingHeatmapUsecase は GetTrackingHeatmapUsecase を生成する。
func NewGetTrackingHeatmapUsecase(
	userRepo *repository.UserRepository,
	recordRepo *repository.RecordRepository,
) *GetTrackingHeatmapUsecase {
	return &GetTrackingHeatmapUsecase{
		userRepo:   userRepo,
		recordRepo: recordRepo,
	}
}

// GetTrackingHeatmapInput is the use case input.
// [Ja] GetTrackingHeatmapInput は GetTrackingHeatmapUsecase の入力。
type GetTrackingHeatmapInput struct {
	// Username is the target profile owner.
	// [Ja] 対象プロフィールオーナーの username。
	Username string
	// TimeZone is the IANA time zone name used to bucket records into days.
	// Resolved by the handler ahead of time (user setting > cookie > default).
	//
	// [Ja] 集計のタイムゾーン名 (IANA)。Handler が事前に解決した値を渡す
	// (ログインユーザー設定 > Cookie > デフォルトの優先順)。
	TimeZone string
	// Now is the reference "now" used to compute today. Injected so tests
	// can pin the heatmap window deterministically.
	//
	// [Ja] 「今日」を決める基準時刻。テストで決定的に固定できるよう外部
	// から注入する。
	Now time.Time
}

// TrackingHeatmapCell represents one day in the heatmap.
// [Ja] TrackingHeatmapCell はヒートマップ上の 1 日分のセル。
type TrackingHeatmapCell struct {
	// Date is the formatted day in the requested time zone (YYYY-MM-DD).
	// [Ja] 指定タイムゾーン上の日付 (YYYY-MM-DD)。
	Date string
	// Count is the number of records on Date (0 for days without records).
	// [Ja] Date の視聴記録数 (記録のない日は 0)。
	Count int
	// LeveledCount is the 0-4 density bucket used to pick a CSS class.
	// [Ja] CSS クラス選択用の 0〜4 段階の密度レベル。
	LeveledCount int
}

// GetTrackingHeatmapOutput is the use case output. Cells contains a
// contiguous list of days from date_from through "today".
//
// [Ja] GetTrackingHeatmapOutput は GetTrackingHeatmapUsecase の出力。
// Cells は date_from から「今日」までの連続した日付配列。
type GetTrackingHeatmapOutput struct {
	Cells []TrackingHeatmapCell
}

// Execute returns the heatmap for the trailing trackingHeatmapDays
// window. Returns *model.AppError with AppErrCodeResourceNotFound when the
// user does not exist or is soft-deleted; the handler converts that to 404.
//
// [Ja] Execute は対象ユーザーの直近 trackingHeatmapDays 日分のヒートマップを返す。
// 対象ユーザーが存在しない / 論理削除済みの場合は AppErrCodeResourceNotFound
// の *model.AppError を返し、Handler 側で 404 に変換される想定。
func (uc *GetTrackingHeatmapUsecase) Execute(ctx context.Context, input GetTrackingHeatmapInput) (*GetTrackingHeatmapOutput, error) {
	userID, err := uc.userRepo.FindActiveIDByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("ユーザーの取得に失敗: %w", err)
	}
	if userID == nil {
		return nil, &model.AppError{
			Code:    model.AppErrCodeResourceNotFound,
			UserMsg: i18n.T(ctx, "error_user_not_found"),
			Metadata: map[string]string{
				"username": input.Username,
			},
		}
	}

	loc, err := time.LoadLocation(input.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("タイムゾーンの解決に失敗 (%q): %w", input.TimeZone, err)
	}

	nowInTZ := input.Now.In(loc)
	today := time.Date(nowInTZ.Year(), nowInTZ.Month(), nowInTZ.Day(), 0, 0, 0, 0, loc)
	dateFrom := beginningOfWeekSunday(today.AddDate(0, 0, -trackingHeatmapDays))

	// dateFrom is midnight in the user's time zone; convert to UTC for the
	// SQL range filter so the existing (user_id, watched_at) index can be
	// used directly on the stored UTC values.
	//
	// [Ja] dateFrom はターゲット TZ の 00:00。SQL 側の範囲フィルタには
	// UTC で渡し、保存値 (UTC) に対する (user_id, watched_at) インデックス
	// が直接効くようにする。
	dateFromUTC := dateFrom.UTC()

	dailyCounts, err := uc.recordRepo.AggregateDailyCountsByUserID(ctx, *userID, dateFromUTC, input.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("ヒートマップ集計に失敗: %w", err)
	}

	countByDate := make(map[string]int, len(dailyCounts))
	for _, c := range dailyCounts {
		countByDate[c.Day.Format("2006-01-02")] = int(c.Count)
	}

	// Count days first via AddDate so the slice can be preallocated even
	// across DST transitions (where Hours()/24 would round incorrectly).
	//
	// [Ja] DST 跨ぎでも常に正しい日数を得るため、Hours()/24 ではなく
	// AddDate でカウントしてからスライスを確保する。
	totalDays := 0
	for d := dateFrom; !d.After(today); d = d.AddDate(0, 0, 1) {
		totalDays++
	}

	cells := make([]TrackingHeatmapCell, 0, totalDays)
	for d := dateFrom; !d.After(today); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		count := countByDate[key]
		cells = append(cells, TrackingHeatmapCell{
			Date:         key,
			Count:        count,
			LeveledCount: leveledRecordCount(count),
		})
	}

	return &GetTrackingHeatmapOutput{Cells: cells}, nil
}

// beginningOfWeekSunday snaps `t` back to the previous Sunday (or returns
// `t` unchanged if it is already Sunday). Mirrors Rails's
// `date.beginning_of_week(:sunday)`.
//
// [Ja] beginningOfWeekSunday は `t` をその週の日曜日まで遡らせる
// (`t` が既に日曜日ならそのまま返す)。Rails の
// `date.beginning_of_week(:sunday)` と同じ挙動。
func beginningOfWeekSunday(t time.Time) time.Time {
	offset := int(t.Weekday()) // Sunday = 0, Monday = 1, ...
	return t.AddDate(0, 0, -offset)
}

// leveledRecordCount maps a daily record count to the 0-4 density bucket
// used by the heatmap CSS. The buckets match the Rails implementation.
//
// [Ja] leveledRecordCount は 1 日の視聴記録数を、ヒートマップ CSS が想定
// する 0〜4 段階の密度バケットに割り振る。Rails 版と同じ閾値。
func leveledRecordCount(count int) int {
	switch {
	case count <= 0:
		return 0
	case count <= 3:
		return 1
	case count <= 6:
		return 2
	case count <= 9:
		return 3
	default:
		return 4
	}
}
