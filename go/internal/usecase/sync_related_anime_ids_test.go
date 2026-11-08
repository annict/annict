package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newSyncRelatedAnimeIDsUsecaseInTxは、テストのトランザクションで関連テーブルを照合する
// SyncRelatedAnimeIDsUsecaseを組み立てる。
func newSyncRelatedAnimeIDsUsecaseInTx(db *sql.DB, tx *sql.Tx, batchSize int) *SyncRelatedAnimeIDsUsecase {
	queries := query.New(db).WithTx(tx)
	return NewSyncRelatedAnimeIDsUsecase(
		repository.NewCastRepository(queries),
		repository.NewStaffRepository(queries),
		repository.NewTrailerRepository(queries),
		repository.NewWorkImageRepository(queries),
		repository.NewVodTitleRepository(queries),
		repository.NewSlotRepository(queries),
		repository.NewProgramRepository(queries),
		repository.NewSeriesWorkRepository(queries),
		batchSize,
	)
}

func insertRelatedRow(t *testing.T, tx *sql.Tx, q string, args ...any) int64 {
	t.Helper()
	var id int64
	if err := tx.QueryRow(q, args...).Scan(&id); err != nil {
		t.Fatalf("テストデータの挿入に失敗: %v", err)
	}
	return id
}

// relatedRowは照合で埋まることを確かめる1つの列の1行。
type relatedRow struct {
	table  string
	column string
	id     int64
	want   int64
}

func TestSyncRelatedAnimeIDsUsecase_Execute_FillsAllColumns(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	// 並行テストのコミットで走査対象が変わらないよう、スナップショットを固定する。
	if _, err := tx.Exec("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ"); err != nil {
		t.Fatalf("テスト用トランザクションの分離レベルの設定に失敗: %v", err)
	}

	// animeにマッピング済みの作品とエピソードを作る。
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	workAnimeID := insertRelatedRow(t, tx, `INSERT INTO animes (status) VALUES ('published') RETURNING id`)
	episodeAnimeID := insertRelatedRow(t, tx, `INSERT INTO animes (status) VALUES ('published') RETURNING id`)
	if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, workAnimeID, int64(workID)); err != nil {
		t.Fatalf("works.anime_idの設定に失敗: %v", err)
	}
	if _, err := tx.Exec(`UPDATE episodes SET anime_id = $1 WHERE id = $2`, episodeAnimeID, int64(episodeID)); err != nil {
		t.Fatalf("episodes.anime_idの設定に失敗: %v", err)
	}

	// 作品・エピソードを指し、anime_idが空の行を各テーブルに作る。
	suffix := fmt.Sprintf("%d", workID)
	channelID := testutil.NewChannelBuilder(t, tx).Build()
	seriesID := insertRelatedRow(t, tx, `INSERT INTO series (name, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id`, "照合テストシリーズ"+suffix)
	slotID := int64(testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithEpisodeID(episodeID).WithChannelID(channelID).Build())
	rows := []relatedRow{
		{table: "casts", column: "anime_id", want: workAnimeID, id: int64(testutil.NewCastBuilder(t, tx, workID).
			WithPersonName("照合テスト声優" + suffix).WithCharacterName("照合テストキャラクター" + suffix).Build())},
		{table: "staffs", column: "anime_id", want: workAnimeID, id: int64(testutil.NewStaffBuilder(t, tx, workID).WithName("照合テストスタッフ" + suffix).Build())},
		{table: "trailers", column: "anime_id", want: workAnimeID, id: insertRelatedRow(t, tx,
			`INSERT INTO trailers (work_id, url, title, created_at, updated_at) VALUES ($1, 'https://example.com/pv', 'PV', NOW(), NOW()) RETURNING id`, int64(workID))},
		{table: "work_images", column: "anime_id", want: workAnimeID, id: insertRelatedRow(t, tx,
			`INSERT INTO work_images (work_id, image_data, created_at, updated_at) VALUES ($1, '{}', NOW(), NOW()) RETURNING id`, int64(workID))},
		{table: "vod_titles", column: "anime_id", want: workAnimeID, id: insertRelatedRow(t, tx,
			`INSERT INTO vod_titles (channel_id, work_id, code, name, created_at, updated_at) VALUES ($1, $2, 'code', 'VODタイトル', NOW(), NOW()) RETURNING id`, channelID, int64(workID))},
		{table: "slots", column: "anime_id", want: workAnimeID, id: slotID},
		{table: "slots", column: "episode_anime_id", want: episodeAnimeID, id: slotID},
		{table: "programs", column: "anime_id", want: workAnimeID, id: testutil.NewProgramBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).Build()},
		{table: "series_works", column: "anime_id", want: workAnimeID, id: insertRelatedRow(t, tx,
			`INSERT INTO series_works (series_id, work_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id`, seriesID, int64(workID))},
	}

	// ページをまたいで走査されるよう、ページを小さくする。
	result, err := newSyncRelatedAnimeIDsUsecaseInTx(db, tx, 2).Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	for _, row := range rows {
		var got sql.NullInt64
		// table・columnはテストで固定した識別子のため、文字列で組み立てても安全。
		if err := tx.QueryRow(fmt.Sprintf(`SELECT %s FROM %s WHERE id = $1`, row.column, row.table), row.id).Scan(&got); err != nil {
			t.Fatalf("%sの行の取得に失敗: %v", row.table, err)
		}
		if !got.Valid || got.Int64 != row.want {
			t.Errorf("%s.%s = %+v、期待値 = %d", row.table, row.column, got, row.want)
		}
	}

	// テスト用DBにはほかのテストがコミットした行もあるため、件数はテストで作った行の数以上になる。
	counts := map[string]int{
		"Casts":               result.Casts,
		"Staffs":              result.Staffs,
		"Trailers":            result.Trailers,
		"WorkImages":          result.WorkImages,
		"VodTitles":           result.VodTitles,
		"SlotsAnimeID":        result.SlotsAnimeID,
		"SlotsEpisodeAnimeID": result.SlotsEpisodeAnimeID,
		"Programs":            result.Programs,
		"SeriesWorks":         result.SeriesWorks,
	}
	for name, count := range counts {
		if count < 1 {
			t.Errorf("%s = %d、期待値 = 1以上", name, count)
		}
	}
}

func TestSyncRelatedAnimeIDsUsecase_Execute_PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	uc := newSyncRelatedAnimeIDsUsecaseInTx(db, tx, 1)

	// キャンセル済みのコンテキストで、最初の列の照合を失敗させる。
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := uc.Execute(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %vをラップしたエラー", err, context.Canceled)
	}
	if result != nil {
		t.Errorf("Execute()の結果 = %+v、期待値 = nil", result)
	}
}

func TestNewSyncRelatedAnimeIDsUsecase_DefaultsBatchSize(t *testing.T) {
	t.Parallel()

	// 非正のbatch sizeは既定値にフォールバックし、ページの走査が必ず前進すること。
	for _, batchSize := range []int{0, -1} {
		uc := NewSyncRelatedAnimeIDsUsecase(nil, nil, nil, nil, nil, nil, nil, nil, batchSize)
		if uc.batchSize != DefaultSyncAnimesBatchSize {
			t.Errorf("batchSize %dを渡したときのbatchSize = %d、期待値 = 既定値の%d", batchSize, uc.batchSize, DefaultSyncAnimesBatchSize)
		}
	}
}
