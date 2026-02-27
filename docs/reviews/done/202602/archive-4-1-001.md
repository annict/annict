# コードレビュー: archive-4-1

## レビュー情報

| 項目                       | 内容                                       |
| -------------------------- | ------------------------------------------ |
| レビュー日                 | 2026-02-27                                 |
| 対象ブランチ               | archive-4-1                                |
| ベースブランチ             | archive                                    |
| 作業計画書（指定があれば） | docs/plans/1_doing/work-episode-archive.md |
| 変更ファイル数             | 16 ファイル                                |
| 変更行数（実装）           | +567 / -11 行（自動生成ファイル除く）      |
| 変更行数（テスト）         | +60 / -8 行                                |

## 参照するガイドライン

- [@CLAUDE.md#レビュー時に参照するガイドライン](/workspace/CLAUDE.md) - ガイドライン一覧

## 変更ファイル一覧

### 実装ファイル

- [x] `go/cmd/server/main.go`
- [x] `go/internal/handler/db_work/handler.go`
- [x] `go/internal/handler/db_work/new.go`
- [x] `go/internal/model/number_format.go`
- [x] `go/internal/query/queries/number_formats.sql`
- [x] `go/internal/query/number_formats.sql.go`
- [x] `go/internal/query/querier.go`
- [x] `go/internal/repository/number_format.go`
- [x] `go/internal/viewmodel/db_work.go`
- [x] `go/internal/templates/pages/db_works/new.templ`
- [x] `go/internal/templates/pages/db_works/new_templ.go`
- [x] `go/internal/i18n/locales/ja.toml`
- [x] `go/internal/i18n/locales/en.toml`

### テストファイル

- [x] `go/internal/handler/db_work/handler_test.go`

### 設定・その他

- [x] `go/internal/handler/popular_work/index_test.go`
- [x] `docs/plans/1_doing/work-episode-archive.md`

## ファイルごとのレビュー結果

問題のあるファイルはありません。全ファイルがガイドラインに準拠しています。

## 設計との整合性チェック

作業計画書のタスク 4-1 に記載された要件を確認しました：

| 要件                                                              | 状態        | 備考                             |
| ----------------------------------------------------------------- | ----------- | -------------------------------- |
| 新規作成フォームハンドラー (GET /db/works/new)                    | ✅ 実装済み | `new.go` で実装                  |
| フォームテンプレート（全フィールド）                              | ✅ 実装済み | `new.templ` に全フィールドを網羅 |
| セレクトボックス用データ取得（media, season, number_format など） | ✅ 実装済み | viewmodel と repository で実装   |

想定サイズとの比較：

- 想定: 約 6 ファイル（実装 4 + テスト 2）、約 250 行（実装 150 行 + テスト 100 行）
- 実際: 16 ファイル（自動生成・翻訳含む）、実装コード約 567 行 + テスト約 60 行
- 翻訳ファイル（ja.toml, en.toml）の 272 行と自動生成ファイル（new_templ.go）の 585 行を除くと、実装コードは妥当な規模

## 設計改善の提案

設計改善の提案はありません。

## 総合評価

**評価**: Approve

**総評**:

タスク 4-1（作品作成フォームのハンドラー・テンプレート実装）として、作業計画書の要件を満たす実装がされています。

**良い点**:

- **ガイドライン準拠**: ハンドラーの標準ファイル名（`handler.go`, `new.go`）、構造体ベースのテンプレート引数パターン（`NewPageData`）、ViewModel での変換など、すべてのガイドラインに準拠
- **3層アーキテクチャ**: NumberFormat の Model → Repository → Handler の流れが正しく、Query への直接依存もない
- **国際化の徹底**: 全フォームラベル・ページタイトルが `templates.T(ctx, ...)` 経由で翻訳されている。翻訳キーの命名規則（`db_works_form_*`）も適切
- **CSRF 対策**: フォームに `csrf_token` の hidden input が含まれている
- **テスト**: TestNew テストが追加され、フォーム要素（form, action, method, CSRF token, 各フィールド名）の存在を確認している
- **エラーハンドリング**: `slog.ErrorContext` を使用した適切なログ出力
- **既存コードとの一貫性**: `testutil.SetupTestDB` パターンの使用、`int64` 型の ID フィールドなど、既存コードベースとの一貫性を維持
