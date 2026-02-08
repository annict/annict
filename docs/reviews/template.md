# コードレビュー: [PR/ブランチ名]

<!--
このテンプレートの使い方:
1. このファイルを `docs/reviews/` ディレクトリにコピー
   - ファイル名: `{ブランチ名}-{サフィックス}.md`
   - サフィックスにはファイル作成時の日時タイムスタンプを付与する（形式: `YYYYMMDDHHMM`）
   - 完了後 `docs/reviews/done/` に移動するため、同一ブランチの再レビュー時にファイル名が重複しないようにする
   - 例: `sign-up-1-1-202602061430.md`
2. ブランチ情報を記入
3. 差分を取得して変更ファイル一覧を更新
4. 各ファイルに対してガイドラインに従っているかチェック
5. 問題点や改善提案を記録（問題がないファイルは「ファイルごとのレビュー結果」に記載しない）

**レビュー後のワークフロー**:
1. レビュアーがこのドキュメントを作成
2. 開発者が各問題点の「対応方針」に回答を記入（判断が必要な項目のみ）
3. 開発者が「追記したので対応をお願いします」とレビュアーに伝える
4. レビュアーが回答に基づいて修正を実施
-->

## レビュー情報

| 項目                   | 内容              |
| ---------------------- | ----------------- |
| レビュー日             | YYYY-MM-DD        |
| 対象ブランチ           | [current-branch]  |
| ベースブランチ         | [base-branch]     |
| 設計書（指定があれば） | [design-doc-path] |
| 変更ファイル数         | N ファイル        |
| 変更行数（実装）       | +N / -N 行        |
| 変更行数（テスト）     | +N / -N 行        |

## 参照するガイドライン

<!--
レビュー時に参照するガイドラインドキュメント。
変更ファイルの種類に応じて該当するガイドラインを確認してください。
-->

### 共通

- [@CLAUDE.md](/workspace/CLAUDE.md) - プロジェクト全体のガイド（コミットメッセージ、コメント、PRガイドライン）

### Go版

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go版の開発ガイド（コーディング規約、テスト戦略）
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTPハンドラーガイドライン
- [@go/docs/templ-guide.md](/workspace/go/docs/templ-guide.md) - templテンプレートガイド
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - リクエストバリデーションガイド
- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化（I18n）ガイド
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - セキュリティガイドライン

### Rails版

- [@rails/CLAUDE.md](/workspace/rails/CLAUDE.md) - Rails版の開発ガイド（コーディング規約、テスト戦略）
- [@rails/docs/architecture-guide.md](/workspace/rails/docs/architecture-guide.md) - アーキテクチャガイド
- [@rails/docs/security-guide.md](/workspace/rails/docs/security-guide.md) - セキュリティガイドライン
- [@rails/docs/testing-guide.md](/workspace/rails/docs/testing-guide.md) - テスト戦略ガイド

## 変更ファイル一覧

<!--
以下のコマンドで変更ファイル一覧を取得:
git diff <base-branch>...<current-branch> --name-only

ファイルの種類ごとにグループ化すると見やすくなります:
- 実装ファイル
- テストファイル
- 設定ファイル
- ドキュメント
-->

### 実装ファイル

- [ ] `path/to/file1.go`
- [ ] `path/to/file2.go`

### テストファイル

- [ ] `path/to/file1_test.go`
- [ ] `path/to/file2_test.go`

### 設定・その他

- [ ] `path/to/config.yaml`

## ファイルごとのレビュー結果

<!--
**重要**: 問題のあるファイルのみ記載してください。
問題がないファイルは「変更ファイル一覧」のチェックボックスにチェックを入れるだけで十分です。

**レビュー手順**:
1. ファイルを読み込む
2. ファイルの種類に応じた該当ガイドラインをチェック
3. ガイドラインに従っていない箇所を問題点として記録
4. 問題点にはガイドラインへの参照を含める

**問題点のフォーマット**:

各問題点には必ず「対応方針」と「回答」の項目を含めてください:

- **修正方針が明確な場合**: 修正案を記載し、「対応方針」には「修正案の通り修正する」等の選択肢を用意する
- **判断が必要な場合**: 修正案に加えて複数の選択肢を「対応方針」に記載し、開発者に回答を求める

「対応方針」と「回答」は**すべての問題点に必須**です。これにより、開発者が判断結果を記録でき、レビュアーが対応状況を追跡できます。

「対応方針」の回答の書き方:
- 選択肢がある場合: 該当する選択肢に `[x]` を付ける
- 自由記述の場合: 「回答」欄に記入
-->

### `path/to/file1.go`（例: 修正方針が明確な場合でも「対応方針」「回答」は必須）

**ステータス**: 要修正

**チェックしたガイドライン**:

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - コーディング規約

**問題点・改善提案**:

- **[@go/CLAUDE.md#ログ出力]**: `log.Printf`を使用している箇所がある

  ```go
  // 問題のあるコード
  log.Printf("ユーザーがログインしました: %d", userID)
  ```

  **修正案**:

  ```go
  // 修正後のコード
  slog.InfoContext(ctx, "ユーザーがログインしました", "user_id", userID)
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->

  - [ ] 案Aの通り完全一致に変更する
  - [ ] パスの末尾に `/` も許容するパターンに変更する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `path/to/file2.go`（例: 判断が必要な場合）

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - CSRF対策

**問題点・改善提案**:

- **[@go/docs/security-guide.md#CSRF対策]**: CSRFスキップパスのマッチ方式が過度に広い

  ```go
  // 問題のあるコード
  if strings.HasPrefix(r.URL.Path, path) {
  ```

  **修正案**:

  ```go
  // 修正後のコード（案A）
  if r.URL.Path == path {
  ```

  **対応方針**:

  <!-- 開発者が回答を記入してください -->

  - [ ] 案Aの通り完全一致に変更する
  - [ ] パスの末尾に `/` も許容するパターンに変更する
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

## 総合評価

<!--
レビュー全体の総合評価を記述します。

評価基準:
- **Approve**: 問題なし、マージ可能
- **Request Changes**: 修正が必要、修正後に再レビュー
- **Comment**: 軽微な指摘のみ、修正は任意
-->

**評価**: [Approve / Request Changes / Comment]

**総評**:

[レビュー全体を通しての所感、良かった点、改善が必要な点などを記述]
