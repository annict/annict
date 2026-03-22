# コードレビュー: feature-flag-2-1

## レビュー情報

| 項目                       | 内容                               |
| -------------------------- | ---------------------------------- |
| レビュー日                 | 2026-03-22                         |
| 対象ブランチ               | feature-flag-2-1                   |
| ベースブランチ             | archive                            |
| 作業計画書（指定があれば） | docs/plans/1_doing/feature-flag.md |
| 変更ファイル数             | 5 ファイル                         |
| 変更行数（実装）           | +139 / -8 行                       |
| 変更行数（テスト）         | +349 / -14 行                      |

## 参照するガイドライン

- [@CLAUDE.md#レビュー時に参照するガイドライン](/workspace/CLAUDE.md) - ガイドライン一覧
- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - Go版の開発ガイド
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - セキュリティガイドライン

## 変更ファイル一覧

### 実装ファイル

- [x] `go/internal/middleware/reverse_proxy.go`
- [x] `go/internal/session/token.go`
- [x] `go/cmd/server/main.go`

### テストファイル

- [x] `go/internal/middleware/reverse_proxy_test.go`

### 設定・その他

- [x] `docs/plans/1_doing/feature-flag.md`

## ファイルごとのレビュー結果

### `go/internal/middleware/reverse_proxy.go`

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - コーディング規約
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド（レイヤー間の依存関係）
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - Cookie設定、セキュリティ
- [docs/plans/1_doing/feature-flag.md](/workspace/docs/plans/1_doing/feature-flag.md) - 作業計画書

**問題点・改善提案**:

- **[作業計画書#ルーティングの流れ]**: `ensureDeviceToken` の戻り値が `Middleware` メソッド内で使用されていない

  `ensureDeviceToken` は新規生成したトークンを返すが、`Middleware` メソッド（198行目）では戻り値を無視している。その後 `isFeatureFlagEnabled`（270行目）が `r.Cookie()` からデバイストークンを読み取るが、新規生成されたトークンはレスポンスにのみ設定されリクエストには追加されないため、初回リクエスト時にデバイストークンを使ったフラグ判定ができない。

  ```go
  // 198行目: 戻り値が無視されている
  m.ensureDeviceToken(w, r)
  ```

  ```go
  // 270行目: 初回リクエスト時は空文字列になる
  deviceToken := ""
  if c, err := r.Cookie(DeviceTokenCookieName); err == nil {
      deviceToken = c.Value
  }
  ```

  実用上は初回リクエスト時に事前登録されたデバイストークンフラグが存在することはないため、直接的なバグではない。ただし、`ensureDeviceToken` の戻り値を `isFeatureFlagEnabled` に渡す設計のほうが意図が明確になる。

  **修正案**:

  ```go
  // Middleware メソッド内
  deviceToken := m.ensureDeviceToken(w, r)

  // isFeatureFlagEnabled の引数にデバイストークンを渡す
  if m.isFeatureFlagEnabled(r, deviceToken) {
  ```

  または、現状のままでも実用上問題がないことをコメントで明示する。

  **対応方針**:
  - [x] 案Aの通り `ensureDeviceToken` の戻り値を `isFeatureFlagEnabled` に渡すようリファクタリングする
  - [ ] 現状維持（初回リクエスト時のエッジケースは実用上問題なし）
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

### `go/internal/middleware/reverse_proxy_test.go`

**ステータス**: 要確認

**チェックしたガイドライン**:

- [@go/CLAUDE.md#テスト戦略](/workspace/go/CLAUDE.md) - テストのベストプラクティス
- [docs/plans/1_doing/feature-flag.md](/workspace/docs/plans/1_doing/feature-flag.md) - 作業計画書

**問題点・改善提案**:

- **[@go/CLAUDE.md#テスト戦略]**: `featureFlaggedPatterns` パッケージレベル変数のテスト内での書き換えに `t.Parallel()` がないことは正しいが、将来的にリスクがある

  複数のテスト関数（`TestIsFeatureFlagEnabled_MatchingPatternEnabled`, `TestIsFeatureFlagEnabled_MatchingPatternDisabled`, `TestIsFeatureFlagEnabled_ErrorFallsBackToFalse`, `TestReverseProxyMiddleware_FeatureFlagRouting`）がパッケージレベルの `featureFlaggedPatterns` 変数を一時的に書き換えている。現在は `t.Parallel()` を使用していないため問題ないが、将来的に並行テストを追加した際にデータ競合が発生するリスクがある。

  ```go
  // 1004行目: パッケージレベル変数の書き換え
  original := featureFlaggedPatterns
  featureFlaggedPatterns = []featureFlaggedPattern{
      {pattern: regexp.MustCompile(`^/test-feature(/.*)?$`), flag: "test_feature"},
  }
  defer func() { featureFlaggedPatterns = original }()
  ```

  現時点では `t.Parallel()` なしで動作しているため問題ないが、パッケージレベル変数の書き換えは一般的にテストのアンチパターンである。ただし、現在の `featureFlaggedPatterns` がパッケージレベル変数として設計されているため（各Go移行タスクで静的に追加される想定）、テスト時の書き換えは妥当なアプローチ。

  **修正案**:

  現状維持で問題ない。ただし、`featureFlaggedPatterns` を書き換えるテストには `t.Parallel()` を追加しないよう注意するコメントを追加することを検討。

  **対応方針**:
  - [x] 現状維持（`t.Parallel()` なしが正しい対処）
  - [ ] パッケージレベル変数を書き換えるテストにコメントを追加（`t.Parallel()` 不可の理由）
  - [ ] その他（下の回答欄に記入）

  **回答**:

  ```
  （ここに回答を記入）
  ```

## 設計改善の提案

設計改善の提案はありません。

## 総合評価

**評価**: Approve

**総評**:

作業計画書のタスク 2-1（リバースプロキシミドルウェアへのフィーチャーフラグ判定の統合）が仕様通りに実装されている。

**良い点**:

- **セキュリティ**: device_token Cookie の属性設定（HttpOnly, Secure, SameSite=Lax, 10年MaxAge）が作業計画書の要件を満たしている
- **信頼性**: `featureFlagRepo` が nil の場合やエラー発生時に Rails 版にフォールバックする設計が適切に実装されている
- **アーキテクチャ**: `featureFlagChecker` インターフェースによりテスタビリティが確保されている。Presentation層（ミドルウェア）からの `session.Manager` への依存も適切
- **テスト網羅性**: device_token 自動生成、既存トークン保持、開発環境の Secure=false、フラグ有効/無効/エラー/nil repo/統合テストなど、必要なケースが網羅されている
- **設計変更の反映**: `sessionUserResolver` インターフェースを不要とし `*session.Manager` を直接使用する設計変更が作業計画書に記録されている
- **既存テストへの影響**: `NewReverseProxyMiddleware` のシグネチャ変更に伴う既存テストの更新（`nil, nil` の追加）が漏れなく行われている

**指摘事項**:

- `ensureDeviceToken` の戻り値が `Middleware` で未使用である点は、実用上問題ないが設計の明確化のために検討の余地がある（必須ではない）
- パッケージレベル変数 `featureFlaggedPatterns` のテスト内書き換えは現状問題ないが、将来的な注意が必要（必須ではない）
