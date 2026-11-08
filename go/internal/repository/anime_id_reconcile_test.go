package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// reconcileSeqは、名前に一意制約のある行 (人物・シリーズなど) をテストごとに作り分ける連番。
var reconcileSeq atomic.Int64

func uniqueReconcileName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, reconcileSeq.Add(1))
}

// reconcileSourceは照合の写し元 (作品かエピソード) の1行。
type reconcileSource struct {
	workID    model.WorkID
	episodeID model.EpisodeID
}

// animeIDReconcileTargetは照合する1つの列と、その列を持つ行の作り方。
type animeIDReconcileTarget struct {
	name   string
	table  string
	column string
	// fkColumnは写し元を指す列。nullableFKがtrueなら、写し元を指さない行も作れる。
	fkColumn   string
	nullableFK bool
	// episodeSourceがtrueなら写し元はエピソード、falseなら作品。
	episodeSource bool
	// createは写し元を指す行を作り、そのIDを返す。srcがnilなら写し元を指さない行を作る。
	create    func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64
	reconcile func(queries *query.Queries) func(ctx context.Context, afterID int64, batchSize int) (repository.AnimeIDReconcilePage, error)
}

func animeIDReconcileTargets() []animeIDReconcileTarget {
	return []animeIDReconcileTarget{
		{
			name: "casts.anime_id", table: "casts", column: "anime_id", fkColumn: "work_id",
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				return int64(testutil.NewCastBuilder(t, tx, src.workID).
					WithPersonName(uniqueReconcileName("照合テスト声優")).
					WithCharacterName(uniqueReconcileName("照合テストキャラクター")).
					Build())
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewCastRepository(q).ReconcileAnimeIDsAfter
			},
		},
		{
			name: "staffs.anime_id", table: "staffs", column: "anime_id", fkColumn: "work_id",
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				return int64(testutil.NewStaffBuilder(t, tx, src.workID).WithName(uniqueReconcileName("照合テストスタッフ")).Build())
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewStaffRepository(q).ReconcileAnimeIDsAfter
			},
		},
		{
			name: "trailers.anime_id", table: "trailers", column: "anime_id", fkColumn: "work_id",
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				return insertReconcileRow(t, tx,
					`INSERT INTO trailers (work_id, url, title, created_at, updated_at) VALUES ($1, 'https://example.com/pv', 'PV', NOW(), NOW()) RETURNING id`,
					int64(src.workID))
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewTrailerRepository(q).ReconcileAnimeIDsAfter
			},
		},
		{
			name: "work_images.anime_id", table: "work_images", column: "anime_id", fkColumn: "work_id",
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				return insertReconcileRow(t, tx,
					`INSERT INTO work_images (work_id, image_data, created_at, updated_at) VALUES ($1, '{}', NOW(), NOW()) RETURNING id`,
					int64(src.workID))
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewWorkImageRepository(q).ReconcileAnimeIDsAfter
			},
		},
		{
			name: "vod_titles.anime_id", table: "vod_titles", column: "anime_id", fkColumn: "work_id", nullableFK: true,
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				var workID any
				if src != nil {
					workID = int64(src.workID)
				}
				channelID := testutil.NewChannelBuilder(t, tx).Build()
				return insertReconcileRow(t, tx,
					`INSERT INTO vod_titles (channel_id, work_id, code, name, created_at, updated_at) VALUES ($1, $2, 'code', 'VODタイトル', NOW(), NOW()) RETURNING id`,
					channelID, workID)
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewVodTitleRepository(q).ReconcileAnimeIDsAfter
			},
		},
		{
			name: "slots.anime_id", table: "slots", column: "anime_id", fkColumn: "work_id",
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				channelID := testutil.NewChannelBuilder(t, tx).Build()
				return int64(testutil.NewSlotBuilder(t, tx).WithWorkID(src.workID).WithChannelID(channelID).Build())
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewSlotRepository(q).ReconcileAnimeIDsAfter
			},
		},
		{
			name: "slots.episode_anime_id", table: "slots", column: "episode_anime_id", fkColumn: "episode_id", nullableFK: true, episodeSource: true,
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				channelID := testutil.NewChannelBuilder(t, tx).Build()
				b := testutil.NewSlotBuilder(t, tx).WithChannelID(channelID)
				if src != nil {
					b = b.WithWorkID(src.workID).WithEpisodeID(src.episodeID)
				} else {
					b = b.WithWorkID(testutil.NewWorkBuilder(t, tx).Build())
				}
				return int64(b.Build())
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewSlotRepository(q).ReconcileEpisodeAnimeIDsAfter
			},
		},
		{
			name: "programs.anime_id", table: "programs", column: "anime_id", fkColumn: "work_id",
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				channelID := testutil.NewChannelBuilder(t, tx).Build()
				return testutil.NewProgramBuilder(t, tx).WithWorkID(src.workID).WithChannelID(channelID).Build()
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewProgramRepository(q).ReconcileAnimeIDsAfter
			},
		},
		{
			name: "series_works.anime_id", table: "series_works", column: "anime_id", fkColumn: "work_id",
			create: func(t *testing.T, tx *sql.Tx, src *reconcileSource) int64 {
				seriesID := insertReconcileRow(t, tx, `INSERT INTO series (name, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id`,
					uniqueReconcileName("照合テストシリーズ"))
				return insertReconcileRow(t, tx,
					`INSERT INTO series_works (series_id, work_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id`,
					seriesID, int64(src.workID))
			},
			reconcile: func(q *query.Queries) func(context.Context, int64, int) (repository.AnimeIDReconcilePage, error) {
				return repository.NewSeriesWorkRepository(q).ReconcileAnimeIDsAfter
			},
		},
	}
}

func insertReconcileRow(t *testing.T, tx *sql.Tx, q string, args ...any) int64 {
	t.Helper()
	var id int64
	if err := tx.QueryRow(q, args...).Scan(&id); err != nil {
		t.Fatalf("テストデータの挿入に失敗: %v", err)
	}
	return id
}

// newReconcileSourceは写し元の作品 (episodeSourceならその作品のエピソードも) を作る。
// withAnimeがtrueなら、写し元にanimeを割り当て、そのIDを返す。
func newReconcileSource(t *testing.T, tx *sql.Tx, episodeSource, withAnime bool) (*reconcileSource, sql.NullInt64) {
	t.Helper()

	src := &reconcileSource{workID: testutil.NewWorkBuilder(t, tx).Build()}
	if episodeSource {
		src.episodeID = testutil.NewEpisodeBuilder(t, tx, src.workID).Build()
	}
	if !withAnime {
		return src, sql.NullInt64{}
	}

	animeID := insertReconcileRow(t, tx, `INSERT INTO animes (status) VALUES ('published') RETURNING id`)
	stmt := `UPDATE works SET anime_id = $1 WHERE id = $2`
	id := int64(src.workID)
	if episodeSource {
		stmt = `UPDATE episodes SET anime_id = $1 WHERE id = $2`
		id = int64(src.episodeID)
	}
	if _, err := tx.Exec(stmt, animeID, id); err != nil {
		t.Fatalf("写し元のanime_idの設定に失敗: %v", err)
	}
	return src, sql.NullInt64{Int64: animeID, Valid: true}
}

// reconcileRowStateは照合の前後で確かめる、行の列の値とupdated_at。
type reconcileRowState struct {
	value     sql.NullInt64
	updatedAt time.Time
}

func setReconcileRowState(t *testing.T, tx *sql.Tx, target animeIDReconcileTarget, id int64, value sql.NullInt64, updatedAt time.Time) {
	t.Helper()
	// table・columnはテストで固定した識別子のため、文字列で組み立てても安全。
	stmt := fmt.Sprintf(`UPDATE %s SET %s = $1, updated_at = $2 WHERE id = $3`, target.table, target.column)
	if _, err := tx.Exec(stmt, value, updatedAt, id); err != nil {
		t.Fatalf("%sの行の準備に失敗: %v", target.table, err)
	}
}

func getReconcileRowState(t *testing.T, tx *sql.Tx, target animeIDReconcileTarget, id int64) reconcileRowState {
	t.Helper()
	var st reconcileRowState
	stmt := fmt.Sprintf(`SELECT %s, updated_at FROM %s WHERE id = $1`, target.column, target.table)
	if err := tx.QueryRow(stmt, id).Scan(&st.value, &st.updatedAt); err != nil {
		t.Fatalf("%sの行の取得に失敗: %v", target.table, err)
	}
	return st
}

func TestAnimeIDReconcile(t *testing.T) {
	t.Parallel()

	for _, target := range animeIDReconcileTargets() {
		t.Run(target.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			// 並行テストのコミットでページの中身が変わらないよう、スナップショットを固定する。
			if _, err := tx.Exec("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ"); err != nil {
				t.Fatalf("テスト用トランザクションの分離レベルの設定に失敗: %v", err)
			}
			ctx := context.Background()
			reconcile := target.reconcile(query.New(db).WithTx(tx))

			srcA, animeA := newReconcileSource(t, tx, target.episodeSource, true)
			srcB, animeB := newReconcileSource(t, tx, target.episodeSource, true)
			srcNoAnime, _ := newReconcileSource(t, tx, target.episodeSource, false)

			type rowCase struct {
				name    string
				src     *reconcileSource
				initial sql.NullInt64
				want    sql.NullInt64
				changed bool
			}
			cases := []rowCase{
				{name: "空の行を写し元のanime_idで埋める", src: srcA, initial: sql.NullInt64{}, want: animeA, changed: true},
				{name: "写し元と食い違う行を直す", src: srcA, initial: animeB, want: animeA, changed: true},
				{name: "写し元のanime_idがまだ無い行をNULLに戻す", src: srcNoAnime, initial: animeA, want: sql.NullInt64{}, changed: true},
				{name: "写し元と一致する行は書き換えない", src: srcB, initial: animeB, want: animeB, changed: false},
			}
			if target.nullableFK {
				cases = append(cases, rowCase{name: "写し元を指さない行をNULLに戻す", src: nil, initial: animeA, want: sql.NullInt64{}, changed: true})
			}

			// updated_atを過去の時刻にしておき、照合で変わらないことを確かめる。
			past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			ids := make([]int64, len(cases))
			wantUpdated := 0
			for i, c := range cases {
				ids[i] = target.create(t, tx, c.src)
				setReconcileRowState(t, tx, target, ids[i], c.initial, past)
				if c.changed {
					wantUpdated++
				}
			}
			firstID, lastID := ids[0], ids[len(ids)-1]

			// 最初のページは先頭の2行だけを照合する。
			page, err := reconcile(ctx, firstID-1, 2)
			if err != nil {
				t.Fatalf("照合のエラー = %v", err)
			}
			if page.LastID != ids[1] {
				t.Errorf("最初のページのLastID = %d、期待値 = %d", page.LastID, ids[1])
			}
			if page.Updated != 2 {
				t.Errorf("最初のページのUpdated = %d、期待値 = 2", page.Updated)
			}
			if got := getReconcileRowState(t, tx, target, ids[2]); got.value != cases[2].initial {
				t.Errorf("ページの外の行の%s = %+v、期待値 = %+v (書き換えない)", target.column, got.value, cases[2].initial)
			}

			// 残りの行を照合する。
			page, err = reconcile(ctx, page.LastID, len(ids))
			if err != nil {
				t.Fatalf("照合のエラー = %v", err)
			}
			if page.LastID != lastID {
				t.Errorf("次のページのLastID = %d、期待値 = %d", page.LastID, lastID)
			}
			if page.Updated != wantUpdated-2 {
				t.Errorf("次のページのUpdated = %d、期待値 = %d", page.Updated, wantUpdated-2)
			}

			for i, c := range cases {
				got := getReconcileRowState(t, tx, target, ids[i])
				if got.value != c.want {
					t.Errorf("%s: %s = %+v、期待値 = %+v", c.name, target.column, got.value, c.want)
				}
				if !got.updatedAt.Equal(past) {
					t.Errorf("%s: updated_at = %v、期待値 = %v (照合で変えない)", c.name, got.updatedAt, past)
				}
			}

			// 照合し終えたあとにもう一度照合しても、書き換える行は無い。
			page, err = reconcile(ctx, firstID-1, len(ids))
			if err != nil {
				t.Fatalf("照合のエラー = %v", err)
			}
			if page.Updated != 0 {
				t.Errorf("2回目の照合のUpdated = %d、期待値 = 0", page.Updated)
			}

			// 最後の行より後ろには行が無いため、ページは空になる。
			page, err = reconcile(ctx, lastID, len(ids))
			if err != nil {
				t.Fatalf("照合のエラー = %v", err)
			}
			if page.LastID != 0 || page.Updated != 0 {
				t.Errorf("空のページ = %+v、期待値 = LastIDとUpdatedが0", page)
			}
		})
	}
}
