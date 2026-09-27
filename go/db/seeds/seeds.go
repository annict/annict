// Package seedsはマスターデータのCSVを埋め込み、投入処理から参照できるようにする。
// 埋め込みは親ディレクトリを参照できないため、CSVを読む側のパッケージではなくCSVと
// 同じ階層に置いている。
package seeds

import "embed"

// FSはマスターデータのCSVを保持する。ファイル名は投入先のテーブル名に対応する。
//
//go:embed *.csv
var FS embed.FS
