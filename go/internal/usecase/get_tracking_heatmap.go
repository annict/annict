package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// trackingHeatmapDaysはヒートマップの表示日数。開始日は
// (今日 - trackingHeatmapDays) をその週の日曜日に丸めた日となり、
// Rails版と同じ挙動になる。
const trackingHeatmapDays = 150

// GetTrackingHeatmapUsecaseは直近trackingHeatmapDays日分の視聴記録
// 数を日次集計し、ヒートマップフラグメント描画用のセル配列に整形する。
type GetTrackingHeatmapUsecase struct {
	userRepo   *repository.UserRepository
	recordRepo *repository.RecordRepository
}

// NewGetTrackingHeatmapUsecaseはGetTrackingHeatmapUsecaseを生成する。
func NewGetTrackingHeatmapUsecase(
	userRepo *repository.UserRepository,
	recordRepo *repository.RecordRepository,
) *GetTrackingHeatmapUsecase {
	return &GetTrackingHeatmapUsecase{
		userRepo:   userRepo,
		recordRepo: recordRepo,
	}
}

// GetTrackingHeatmapInputはGetTrackingHeatmapUsecaseの入力。
type GetTrackingHeatmapInput struct {
	// 対象プロフィールオーナーのusername。
	Username string
	// 集計のタイムゾーン名 (IANA)。Handlerが事前に解決した値を渡す
	// (ログインユーザー設定 > Cookie > デフォルトの優先順)。
	TimeZone string
	// 「今日」を決める基準時刻。テストで決定的に固定できるよう外部
	// から注入する。
	Now time.Time
}

// TrackingHeatmapCellはヒートマップ上の1日分のセル。
type TrackingHeatmapCell struct {
	// 指定タイムゾーン上の日付 (YYYY-MM-DD)。
	Date string
	// Dateの視聴記録数 (記録のない日は0)。
	Count int
	// CSSクラス選択用の0〜4段階の密度レベル。
	LeveledCount int
}

// GetTrackingHeatmapOutputはGetTrackingHeatmapUsecaseの出力。
// Cellsはdate_fromから「今日」までの連続した日付配列。
type GetTrackingHeatmapOutput struct {
	Cells []TrackingHeatmapCell
}

// Executeは対象ユーザーの直近trackingHeatmapDays日分のヒートマップを返す。
// 対象ユーザーが存在しない / 論理削除済みの場合はAppErrCodeResourceNotFound
// の *model.AppErrorを返し、Handler側で404に変換される想定。
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

	// dateFromはターゲットTZの00:00。SQL側の範囲フィルタには
	// UTCで渡し、保存値 (UTC) に対する (user_id, watched_at) インデックス
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

	// DST跨ぎでも常に正しい日数を得るため、Hours()/24ではなく
	// AddDateでカウントしてからスライスを確保する。
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

// beginningOfWeekSundayは `t` をその週の日曜日まで遡らせる
// (`t` が既に日曜日ならそのまま返す)。Railsの
// `date.beginning_of_week(:sunday)` と同じ挙動。
func beginningOfWeekSunday(t time.Time) time.Time {
	offset := int(t.Weekday()) // Sunday = 0, Monday = 1, ...
	return t.AddDate(0, 0, -offset)
}

// leveledRecordCountは1日の視聴記録数を、ヒートマップCSSが想定
// する0〜4段階の密度バケットに割り振る。Rails版と同じ閾値。
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
