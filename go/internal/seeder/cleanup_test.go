package seeder

import (
	"context"
	"database/sql"
	"net/url"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

func TestCleanupExistingData_RemovesRelatedRowsBeforeAnimeSync(t *testing.T) {
	t.Parallel()

	db := newSeederTestDB(t)
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("テストデータのトランザクション開始に失敗: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	channelID := testutil.NewChannelBuilder(t, tx).Build()
	programID := testutil.NewProgramBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).Build()
	testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithEpisodeID(episodeID).
		WithChannelID(channelID).WithProgramID(programID).Build()
	testutil.NewCastBuilder(t, tx, workID).Build()
	testutil.NewStaffBuilder(t, tx, workID).Build()
	testutil.NewWorkImageBuilder(t, tx, workID).Build()

	var animeID int64
	if err := tx.QueryRowContext(ctx, "SELECT anime_id FROM works WHERE id = $1", int64(workID)).Scan(&animeID); err != nil {
		t.Fatalf("作品のanime_idの取得に失敗: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO trailers (work_id, anime_id, url, title, created_at, updated_at)
		VALUES ($1, $2, 'https://example.com/pv', 'PV', NOW(), NOW())
	`, int64(workID), animeID); err != nil {
		t.Fatalf("PVの作成に失敗: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		WITH inserted_series AS (
			INSERT INTO series (name, created_at, updated_at)
			VALUES ('片付けのテスト', NOW(), NOW()) RETURNING id
		)
		INSERT INTO series_works (series_id, work_id, anime_id, created_at, updated_at)
		SELECT id, $1, $2, NOW(), NOW() FROM inserted_series
	`, int64(workID), animeID); err != nil {
		t.Fatalf("シリーズの作品の作成に失敗: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO anime_classifications (anime_id, kind) VALUES ($1, 'work')", animeID); err != nil {
		t.Fatalf("作品の分類の作成に失敗: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("テストデータのコミットに失敗: %v", err)
	}

	// 後からNOT NULLが付いた状態でも、片付け後の照合でNULLを書かないことを確かめる。
	relatedTables := []string{"casts", "staffs", "trailers", "work_images", "slots", "programs", "series_works"}
	for _, table := range relatedTables {
		stmt := "ALTER TABLE " + pq.QuoteIdentifier(table) + " ADD CONSTRAINT seed_cleanup_anime_id_not_null CHECK (anime_id IS NOT NULL) NOT VALID"
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("%sの検証用制約の追加に失敗: %v", table, err)
		}
	}
	if err := cleanupExistingData(ctx, db); err != nil {
		t.Fatalf("cleanupExistingData()のエラー = %v", err)
	}
	for _, table := range append(relatedTables, "works", "episodes") {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+pq.QuoteIdentifier(table)).Scan(&count); err != nil {
			t.Fatalf("%sの件数の取得に失敗: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("%sの件数 = %d、期待値 = 0", table, count)
		}
	}
	for _, target := range []struct {
		table string
		query string
	}{
		{table: "animes", query: "SELECT COUNT(*) FROM animes WHERE id = $1"},
		{table: "anime_classifications", query: "SELECT COUNT(*) FROM anime_classifications WHERE anime_id = $1"},
	} {
		var count int
		if err := db.QueryRowContext(ctx, target.query, animeID).Scan(&count); err != nil {
			t.Fatalf("%sの件数の取得に失敗: %v", target.table, err)
		}
		if count != 1 {
			t.Errorf("%sの件数 = %d、期待値 = 1 (保持する)", target.table, count)
		}
	}

	q := query.New(db)
	relatedSync := usecase.NewSyncRelatedAnimeIDsUsecase(
		repository.NewCastRepository(q), repository.NewStaffRepository(q),
		repository.NewTrailerRepository(q), repository.NewWorkImageRepository(q),
		repository.NewVodTitleRepository(q), repository.NewSlotRepository(q),
		repository.NewProgramRepository(q), repository.NewSeriesWorkRepository(q), 1000,
	)
	if _, err := relatedSync.Execute(ctx); err != nil {
		t.Fatalf("片付け後の関連テーブルの照合に失敗: %v", err)
	}
}

// newSeederTestDBは、全件削除が共有DBの他パッケージへ影響しないよう、使い捨てのDBを作る。
// スキーマは正本のschema.sqlをpsqlで適用する。
func newSeederTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbName := "annict_test_seeder_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	admin := testutil.GetTestDB()
	if _, err := admin.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName)); err != nil {
		t.Fatalf("使い捨てDBの作成に失敗: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE " + pq.QuoteIdentifier(dbName)); err != nil {
			t.Errorf("使い捨てDBの削除に失敗: %v", err)
		}
	})
	dbURL, err := url.Parse(testutil.TestDatabaseURL())
	if err != nil {
		t.Fatalf("テストDBの接続URLの解析に失敗: %v", err)
	}
	dbURL.Path = "/" + dbName
	db, err := sql.Open("postgres", dbURL.String())
	if err != nil {
		t.Fatalf("使い捨てDBの接続に失敗: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("使い捨てDBの接続のクローズに失敗: %v", err)
		}
	})
	cmd := exec.Command("psql", "--no-psqlrc", "--dbname", dbURL.String(), "--set", "ON_ERROR_STOP=1", "--file", "../../db/schema.sql")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("使い捨てDBへのスキーマ適用に失敗: %v\n%s", err, output)
	}
	return db
}
