// Package masterはマスターデータ (チャンネルグループ・チャンネル・エピソード番号
// フォーマット・都道府県) を、埋め込んだCSVからデータベースへ投入する。
//
// 投入は主キーを指定したupsertで、対象のテーブル以外には触れない。既存の開発データを
// 消さずに何度でも実行できるため、日常的なセットアップ手順から呼べる。対象テーブルを
// 空にしてから作り直すinternal/seederとは、この点で役割が分かれる。
package master

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/lib/pq"

	"github.com/annict/annict/go/db/seeds"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/seeder"
)

// tablesは投入先のテーブルを外部キーの依存順に並べたもの。CSVのファイル名はテーブル名に
// 対応し、どの列を投入するかはCSVのヘッダーが決める。
var tables = []string{
	"channel_groups",
	"channels",
	"number_formats",
	"prefectures",
}

// idColumnは投入先のテーブルに共通する主キーの列名。CSVが持つIDをそのまま使うことで、
// 再実行が新しい行を増やさずに同じ行を更新する形になる。
const idColumn = "id"

// timestampColumnsはCSVに含まれないときだけ、INSERTで現在時刻を入れる列。
// number_formatsとprefecturesのCSVはこの2列を持たないが、テーブル側はNOT NULLである。
// 既存行の更新では触らない。CSVに由来しない値を毎回書き換えると、同じCSVを投入し直した
// だけで行が変わってしまうため。
var timestampColumns = []string{"created_at", "updated_at"}

// columnConvertersはCSVの表記がPostgreSQLの入力形式と一致しない列の変換。ここに無い列は
// 文字列のままドライバへ渡し、型の解釈は挿入先の列に委ねる。
var columnConverters = map[string]map[string]func(string) (any, error){
	"number_formats": {"data": parseTextArray},
}

// Runはマスターデータをデータベースへ投入する。
//
// 環境ガードはデータベースに触れるどの処理よりも先に実行する。CLIのタスクガードも同じ
// 確認を行うが、このパッケージを直接呼ぶ経路でも開発用以外のデータを書き換えないように
// するため、ここでも繰り返す。
func Run(ctx context.Context, cfg *config.Config, db *sql.DB) error {
	if err := seeder.EnsureSeedableEnv(cfg); err != nil {
		return err
	}

	slog.InfoContext(ctx, "マスターデータを投入します", "env", cfg.Env)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() {
		// コミット済みの場合もRollbackは呼ばれるが、ErrTxDoneが返るだけで影響は無い
		_ = tx.Rollback()
	}()

	if err := run(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}

// runは投入の本体で、テーブルを外部キーの依存順に処理する。トランザクションを引数で
// 受けるため、テストはロールバックされるトランザクションの中で同じ経路を実行できる。
func run(ctx context.Context, tx query.DBTX) error {
	for _, table := range tables {
		count, err := seedTable(ctx, tx, table)
		if err != nil {
			return fmt.Errorf("%sの投入に失敗しました: %w", table, err)
		}

		slog.InfoContext(ctx, "マスターデータを投入しました", "table", table, "rows", count)
	}

	return nil
}

// seedTableは1テーブル分のCSVを投入し、投入した行数を返す。
func seedTable(ctx context.Context, tx query.DBTX, table string) (int, error) {
	header, rows, err := readCSV(table)
	if err != nil {
		return 0, err
	}

	nullable, err := nullableColumns(ctx, tx, table)
	if err != nil {
		return 0, err
	}

	stmt := upsertStatement(table, header)
	for i, row := range rows {
		args, err := rowArgs(table, header, row, nullable)
		if err != nil {
			// ヘッダーが1行目のため、データ行の行番号は添字に2を足したものになる
			return 0, fmt.Errorf("CSVの%d行目の変換に失敗しました: %w", i+2, err)
		}

		if _, err := tx.ExecContext(ctx, stmt, args...); err != nil {
			return 0, fmt.Errorf("CSVの%d行目の投入に失敗しました: %w", i+2, err)
		}
	}

	if err := resetSequence(ctx, tx, table); err != nil {
		return 0, err
	}

	return len(rows), nil
}

// readCSVは埋め込んだCSVをヘッダーとデータ行に分けて返す。
func readCSV(table string) (header []string, rows [][]string, err error) {
	file, err := seeds.FS.Open(table + ".csv")
	if err != nil {
		return nil, nil, fmt.Errorf("CSVを開けませんでした: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("CSVのクローズに失敗しました: %w", closeErr)
		}
	}()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("CSVの読み込みに失敗しました: %w", err)
	}
	if len(records) == 0 {
		return nil, nil, errors.New("CSVが空です")
	}

	header = records[0]
	if !slices.Contains(header, idColumn) {
		return nil, nil, fmt.Errorf("CSVのヘッダーに%s列がありません", idColumn)
	}

	return header, records[1:], nil
}

// upsertStatementはCSVのヘッダーが示す列を、主キーの衝突時に上書きするINSERT文へ組む。
// CSVに無いtimestampColumnsはINSERTのときだけ現在時刻で埋めるため、更新の対象にしない。
func upsertStatement(table string, header []string) string {
	columns := make([]string, 0, len(header)+len(timestampColumns))
	values := make([]string, 0, len(header)+len(timestampColumns))
	updates := make([]string, 0, len(header))

	for i, column := range header {
		columns = append(columns, pq.QuoteIdentifier(column))
		values = append(values, fmt.Sprintf("$%d", i+1))

		if column == idColumn {
			continue
		}
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", pq.QuoteIdentifier(column), pq.QuoteIdentifier(column)))
	}

	for _, column := range timestampColumns {
		if slices.Contains(header, column) {
			continue
		}
		columns = append(columns, pq.QuoteIdentifier(column))
		values = append(values, "now()")
	}

	conflict := "DO NOTHING"
	if len(updates) > 0 {
		conflict = "DO UPDATE SET " + strings.Join(updates, ", ")
	}

	// テーブル名は固定リスト、列名は埋め込んだCSVのヘッダー由来で、いずれも識別子として引用している
	return fmt.Sprintf( // #nosec G201
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) %s",
		pq.QuoteIdentifier(table),
		strings.Join(columns, ", "),
		strings.Join(values, ", "),
		pq.QuoteIdentifier(idColumn),
		conflict,
	)
}

// rowArgsはCSVの1行をINSERT文のプレースホルダーに対応する引数へ変換する。
func rowArgs(table string, header, row []string, nullable map[string]bool) ([]any, error) {
	args := make([]any, len(header))
	for i, column := range header {
		value := row[i]

		if convert, ok := columnConverters[table][column]; ok {
			converted, err := convert(value)
			if err != nil {
				return nil, fmt.Errorf("%s列の変換に失敗しました: %w", column, err)
			}
			args[i] = converted

			continue
		}

		if value == "" && nullable[column] {
			args[i] = nil

			continue
		}

		args[i] = value
	}

	return args, nil
}

// nullableColumnsは対象テーブルのうちNULLを許す列を返す。
//
// Goのencoding/csvは引用の有無を返さないため、CSVの空の値が空文字とNULLのどちらを表して
// いるかをCSVからは判別できない。空文字をNULLに倒すかどうかは列の制約から決める。
// そうしないと、NOT NULLの列に空文字を持つCSV (channels.name_alter、
// number_formats.format) がNULL違反で投入できない。
func nullableColumns(ctx context.Context, tx query.DBTX, table string) (map[string]bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT column_name, is_nullable = 'YES'
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
	`, table)
	if err != nil {
		return nil, fmt.Errorf("列の定義の取得に失敗しました: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	nullable := map[string]bool{}
	for rows.Next() {
		var (
			column     string
			isNullable bool
		)
		if err := rows.Scan(&column, &isNullable); err != nil {
			return nil, fmt.Errorf("列の定義の読み取りに失敗しました: %w", err)
		}
		nullable[column] = isNullable
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("列の定義の読み取りに失敗しました: %w", err)
	}
	if len(nullable) == 0 {
		return nil, fmt.Errorf("テーブル %sが見つかりません", table)
	}

	return nullable, nil
}

// resetSequenceは主キーを指定したINSERTで取り残されたシーケンスを、テーブルの最大IDへ
// 合わせる。これをしないと、次に採番されるIDが投入済みの行と衝突する。
//
// 合わせる先はCSVの最大IDではなくテーブルの最大IDのため、CSVより後に追加された行が
// 残っていればその先から採番される。その行が削除されている場合や、別のトランザクションが
// 未コミットで保持している場合は最大IDに現れないため、シーケンスはその位置まで下がる。
// setvalはトランザクションの外で効くため、この呼び出しはロールバックしても戻らない。
func resetSequence(ctx context.Context, tx query.DBTX, table string) error {
	// テーブル名は固定リスト由来で、識別子として引用している
	stmt := fmt.Sprintf( // #nosec G201
		"SELECT setval($1::regclass, MAX(%s)) FROM %s",
		pq.QuoteIdentifier(idColumn),
		pq.QuoteIdentifier(table),
	)

	if _, err := tx.ExecContext(ctx, stmt, table+"_id_seq"); err != nil {
		return fmt.Errorf("シーケンスの更新に失敗しました: %w", err)
	}

	return nil
}

// parseTextArrayはCSVの配列表記をPostgreSQLの配列へ変換する。CSVはRailsがArray#to_sで
// 書き出しており、JSONの配列と同じ表記になる。
func parseTextArray(value string) (any, error) {
	var elements []string
	if err := json.Unmarshal([]byte(value), &elements); err != nil {
		return nil, fmt.Errorf("配列として解釈できません (%q): %w", value, err)
	}

	return pq.StringArray(elements), nil
}
