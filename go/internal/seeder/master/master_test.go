package master

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/testutil"
)

// TestRun_InsertsAllCSVRowsはCSVの全行が投入されることを、対象テーブルすべてについて
// 確認する。CSVごとにヘッダーの構成が異なるため、1テーブルだけでは列の組み立てが壊れても
// 気付けない。
func TestRun_InsertsAllCSVRows(t *testing.T) {
	t.Parallel()

	tx := setupMasterTx(t)
	ctx := context.Background()

	if err := run(ctx, tx); err != nil {
		t.Fatalf("run()のエラー = %v", err)
	}

	for _, table := range tables {
		ids := csvIDs(t, table)

		var count int
		stmt := fmt.Sprintf("SELECT count(*) FROM %s WHERE id = ANY($1)", pq.QuoteIdentifier(table))
		if err := tx.QueryRowContext(ctx, stmt, pq.Array(ids)).Scan(&count); err != nil {
			t.Fatalf("%sの件数の取得に失敗した: %v", table, err)
		}
		if count != len(ids) {
			t.Errorf("%sの投入済みの件数 = %d、期待値 = %d", table, count, len(ids))
		}
	}
}

// TestRun_ConvertsCSVValuesはCSVの表記とPostgreSQLの値の対応を確認する。CSVは空の値で
// NULLと空文字の両方を表し、配列をJSONと同じ表記で持ち、テーブルによっては時刻の列を
// 持たない。いずれも文字列のまま渡すだけでは正しい値にならない。
func TestRun_ConvertsCSVValues(t *testing.T) {
	t.Parallel()

	tx := setupMasterTx(t)
	ctx := context.Background()

	if err := run(ctx, tx); err != nil {
		t.Fatalf("run()のエラー = %v", err)
	}

	// 空の値がNULLになる列 (nullable) と、空文字のまま入る列 (NOT NULL)
	var (
		nameAlter string
		deletedAt sql.NullTime
	)
	if err := tx.QueryRowContext(ctx, "SELECT name_alter, deleted_at FROM channels WHERE id = 1").Scan(&nameAlter, &deletedAt); err != nil {
		t.Fatalf("channelsの取得に失敗した: %v", err)
	}
	if nameAlter != "" {
		t.Errorf("channels.name_alter = %q、期待値 = %q", nameAlter, "")
	}
	if deletedAt.Valid {
		t.Errorf("channels.deleted_at = %v、期待値 = NULL", deletedAt.Time)
	}

	// JSONと同じ表記の配列がPostgreSQLの配列になる
	var (
		empty  pq.StringArray
		filled pq.StringArray
	)
	if err := tx.QueryRowContext(ctx, "SELECT data FROM number_formats WHERE id = 1").Scan(&empty); err != nil {
		t.Fatalf("number_formatsの取得に失敗した: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("number_formats.data (id = 1) の要素数 = %d、期待値 = 0", len(empty))
	}
	if err := tx.QueryRowContext(ctx, "SELECT data FROM number_formats WHERE id = 2").Scan(&filled); err != nil {
		t.Fatalf("number_formatsの取得に失敗した: %v", err)
	}
	if len(filled) == 0 || filled[0] != "第一話" {
		t.Errorf("number_formats.data (id = 2) = %v、期待値 = 先頭が%qの配列", filled, "第一話")
	}

	// CSVに無いcreated_at / updated_atが現在時刻で埋まる
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx, "SELECT created_at FROM prefectures WHERE id = 1").Scan(&createdAt); err != nil {
		t.Fatalf("prefecturesの取得に失敗した: %v", err)
	}
	if createdAt.IsZero() {
		t.Error("prefectures.created_atがゼロ値だった")
	}
}

// TestRun_OverwritesChangedRowsは、変更された行がCSVの内容へ戻ることを確認する。
// マスターデータの投入は同じ行を更新し続ける形で冪等にしているため、行が増えないことと
// 内容が揃うことの両方が要る。
func TestRun_OverwritesChangedRows(t *testing.T) {
	t.Parallel()

	tx := setupMasterTx(t)
	ctx := context.Background()

	if err := run(ctx, tx); err != nil {
		t.Fatalf("1回目のrun()のエラー = %v", err)
	}

	var want string
	if err := tx.QueryRowContext(ctx, "SELECT name FROM channels WHERE id = 1").Scan(&want); err != nil {
		t.Fatalf("channelsの取得に失敗した: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE channels SET name = '変更後' WHERE id = 1"); err != nil {
		t.Fatalf("channelsの更新に失敗した: %v", err)
	}

	if err := run(ctx, tx); err != nil {
		t.Fatalf("2回目のrun()のエラー = %v", err)
	}

	var got string
	if err := tx.QueryRowContext(ctx, "SELECT name FROM channels WHERE id = 1").Scan(&got); err != nil {
		t.Fatalf("channelsの取得に失敗した: %v", err)
	}
	if got != want {
		t.Errorf("channels.name = %q、期待値 = %q", got, want)
	}

	ids := csvIDs(t, "channels")

	var count int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM channels WHERE id = ANY($1)", pq.Array(ids)).Scan(&count); err != nil {
		t.Fatalf("channelsの件数の取得に失敗した: %v", err)
	}
	if count != len(ids) {
		t.Errorf("2回目の投入後のchannelsの件数 = %d、期待値 = %d", count, len(ids))
	}
}

// TestRun_KeepsTimestampsOnUpdateは、CSVに無い時刻の列が再実行で書き換わらないことを
// 確認する。CSVに由来しない値を毎回更新すると、同じCSVを投入し直しただけで行が変わる。
func TestRun_KeepsTimestampsOnUpdate(t *testing.T) {
	t.Parallel()

	tx := setupMasterTx(t)
	ctx := context.Background()

	if err := run(ctx, tx); err != nil {
		t.Fatalf("1回目のrun()のエラー = %v", err)
	}

	// now()は同じトランザクション内で変わらないため、上書きを検出できる過去日時を設定する。
	wantCreatedAt := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	wantUpdatedAt := time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, table := range []string{"number_formats", "prefectures"} {
		stmt := fmt.Sprintf("UPDATE %s SET created_at = $1, updated_at = $2 WHERE id = 1", pq.QuoteIdentifier(table))
		if _, err := tx.ExecContext(ctx, stmt, wantCreatedAt, wantUpdatedAt); err != nil {
			t.Fatalf("%sの時刻の更新に失敗した: %v", table, err)
		}
	}

	if err := run(ctx, tx); err != nil {
		t.Fatalf("2回目のrun()のエラー = %v", err)
	}

	for _, table := range []string{"number_formats", "prefectures"} {
		var createdAt, updatedAt time.Time
		stmt := fmt.Sprintf("SELECT created_at, updated_at FROM %s WHERE id = 1", pq.QuoteIdentifier(table))
		if err := tx.QueryRowContext(ctx, stmt).Scan(&createdAt, &updatedAt); err != nil {
			t.Fatalf("%sの時刻の取得に失敗した: %v", table, err)
		}
		if !createdAt.Equal(wantCreatedAt) {
			t.Errorf("再実行後の%s.created_at = %v、期待値 = %v", table, createdAt, wantCreatedAt)
		}
		if !updatedAt.Equal(wantUpdatedAt) {
			t.Errorf("再実行後の%s.updated_at = %v、期待値 = %v", table, updatedAt, wantUpdatedAt)
		}
	}
}

// TestRun_AdvancesSequenceは、投入後に採番されるIDが投入済みの行と衝突しないことを
// 確認する。CSVはIDを指定して投入するため、シーケンスは進まないままになる。
func TestRun_AdvancesSequence(t *testing.T) {
	t.Parallel()

	tx := setupMasterTx(t)
	ctx := context.Background()

	if err := run(ctx, tx); err != nil {
		t.Fatalf("run()のエラー = %v", err)
	}

	var maxID, nextID int64
	if err := tx.QueryRowContext(ctx, "SELECT MAX(id) FROM channels").Scan(&maxID); err != nil {
		t.Fatalf("channelsの最大IDの取得に失敗した: %v", err)
	}
	if err := tx.QueryRowContext(ctx, "SELECT nextval('channels_id_seq')").Scan(&nextID); err != nil {
		t.Fatalf("シーケンスの取得に失敗した: %v", err)
	}
	if nextID <= maxID {
		t.Errorf("次に採番されるID = %d、期待値 = %dより大きい値", nextID, maxID)
	}
}

// TestSetupMasterTx_IsolatesSequencesは、テストごとの使い捨てDBがテーブルだけでなく
// シーケンスも独立していることを確認する。setvalはロールバックで戻らないため、テーブルだけを
// 分離した構成に戻すと、シードの再投入が他のテストの採番を巻き戻す。
func TestSetupMasterTx_IsolatesSequences(t *testing.T) {
	t.Parallel()

	tx := setupMasterTx(t)
	otherTx := setupMasterTx(t)
	ctx := context.Background()
	for _, target := range []*sql.Tx{tx, otherTx} {
		if err := run(ctx, target); err != nil {
			t.Fatalf("run()のエラー = %v", err)
		}
	}

	var firstID int64
	if err := otherTx.QueryRowContext(ctx, "INSERT INTO channel_groups (name) VALUES ('1行目') RETURNING id").Scan(&firstID); err != nil {
		t.Fatalf("もう一方のDBの1行目の作成に失敗した: %v", err)
	}
	if err := run(ctx, tx); err != nil {
		t.Fatalf("再投入時のrun()のエラー = %v", err)
	}

	var secondID int64
	if err := otherTx.QueryRowContext(ctx, "INSERT INTO channel_groups (name) VALUES ('2行目') RETURNING id").Scan(&secondID); err != nil {
		t.Fatalf("もう一方のDBの2行目の作成に失敗した: %v", err)
	}
	if secondID != firstID+1 {
		t.Errorf("もう一方のDBの次のID = %d、期待値 = %d", secondID, firstID+1)
	}
}

// TestRun_RejectsNonSeedableEnvは、意図的にnilのデータベースハンドルを渡す。環境の判定は
// データベースに触れる前に行わなければならず、通過してしまう実装であればnilのハンドルで
// panicする。
func TestRun_RejectsNonSeedableEnv(t *testing.T) {
	t.Parallel()

	if err := Run(context.Background(), &config.Config{Env: "prod"}, nil); err == nil {
		t.Fatal("Run() = nil、期待値 = エラーあり")
	}
}

// TestUpsertStatementはCSVのヘッダーからINSERT文を組む部分を、データベースを介さずに
// 確認する。CSVに無い時刻の列の扱いと、更新対象の列の決まり方が対象。
func TestUpsertStatement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   []string
		contains []string
	}{
		{
			name:     "時刻の列を持つCSVはそのまま投入する",
			header:   []string{"id", "name", "created_at", "updated_at"},
			contains: []string{`("id", "name", "created_at", "updated_at")`, "VALUES ($1, $2, $3, $4)", `"name" = EXCLUDED."name"`, `"created_at" = EXCLUDED."created_at"`},
		},
		{
			name:     "時刻の列を持たないCSVはINSERTのときだけ現在時刻で埋める",
			header:   []string{"id", "name"},
			contains: []string{`("id", "name", "created_at", "updated_at")`, "VALUES ($1, $2, now(), now())", `DO UPDATE SET "name" = EXCLUDED."name"`},
		},
		{
			name:     "主キーだけのCSVは更新する列を持たない",
			header:   []string{"id"},
			contains: []string{`ON CONFLICT ("id") DO NOTHING`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stmt := upsertStatement("prefectures", tt.header)
			for _, want := range tt.contains {
				if !strings.Contains(stmt, want) {
					t.Errorf("upsertStatement() = %q、期待値 = %qを含む文", stmt, want)
				}
			}
			if strings.Contains(stmt, `"id" = EXCLUDED."id"`) {
				t.Errorf("upsertStatement() = %q、期待値 = 主キーを更新しない文", stmt)
			}
		})
	}
}

// TestParseTextArrayはCSVの配列表記の解釈を、データベースを介さずに確認する。表記が壊れた
// CSVを黙って空の配列として投入せず、エラーとして止まることが要る。
func TestParseTextArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    pq.StringArray
		wantErr bool
	}{
		{name: "空の配列", value: `[]`, want: pq.StringArray{}},
		{name: "要素を持つ配列", value: `["第一話", "第二話"]`, want: pq.StringArray{"第一話", "第二話"}},
		{name: "配列として解釈できない", value: `第一話`, wantErr: true},
		{name: "空の値", value: ``, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseTextArray(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseTextArray(%q) = %v、期待値 = エラーあり", tt.value, got)
				}

				return
			}
			if err != nil {
				t.Fatalf("parseTextArray(%q)のエラー = %v", tt.value, err)
			}

			array, ok := got.(pq.StringArray)
			if !ok {
				t.Fatalf("parseTextArray(%q)の型 = %T、期待値 = pq.StringArray", tt.value, got)
			}
			if !slices.Equal(array, tt.want) {
				t.Errorf("parseTextArray(%q) = %v、期待値 = %v", tt.value, array, tt.want)
			}
		})
	}
}

// csvIDsは対象テーブルのCSVが持つIDの一覧を返す。
func csvIDs(t *testing.T, table string) []int64 {
	t.Helper()

	header, rows, err := readCSV(table)
	if err != nil {
		t.Fatalf("readCSV(%q)のエラー = %v", table, err)
	}

	index := -1
	for i, column := range header {
		if column == idColumn {
			index = i
		}
	}
	if index < 0 {
		t.Fatalf("%sのCSVに%s列が無い", table, idColumn)
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		id, err := strconv.ParseInt(row[index], 10, 64)
		if err != nil {
			t.Fatalf("%sのIDを数値として読めない: %v", table, err)
		}
		ids = append(ids, id)
	}

	return ids
}

// setupMasterTxはテスト専用の使い捨てDBを用意し、トランザクションを返す。
// setvalはロールバックで戻らないため、共有テストDBからシーケンスとテーブルをまとめて隔離する。
// スキーマはTestMainが適用済みのテンプレートから複製して受け継ぐ。
func setupMasterTx(t *testing.T) *sql.Tx {
	t.Helper()

	dbName := newDBName("")
	admin := testutil.GetTestDB()
	stmt := fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", pq.QuoteIdentifier(dbName), pq.QuoteIdentifier(templateDBName))
	if _, err := admin.Exec(stmt); err != nil {
		t.Fatalf("使い捨てDBの作成に失敗した: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE " + pq.QuoteIdentifier(dbName)); err != nil {
			t.Errorf("使い捨てDBの削除に失敗した: %v", err)
		}
	})

	db, err := sql.Open("postgres", dbURLFor(dbName))
	if err != nil {
		t.Fatalf("使い捨てDBの接続に失敗した: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("使い捨てDBの接続のクローズに失敗した: %v", err)
		}
	})

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("トランザクションの開始に失敗した: %v", err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			t.Errorf("トランザクションのロールバックに失敗した: %v", err)
		}
	})

	return tx
}
