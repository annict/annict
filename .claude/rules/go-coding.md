---
paths:
  - "go/**/*.{go,templ}"
---

# コーディング規約

このドキュメントは、Go 版 Wikino のコーディング規約を説明します。

## Go コード

- **インデント**: タブを使用（Go 標準）
- **フォーマット**: `gofmt` を使用して自動フォーマット
- **コメント**: 自動生成されたコード以外は日本語でコメントを記述する
  - 関数やメソッドの説明は日本語で記述
  - インラインコメントも日本語で統一
  - sqlc 等のツールが生成したコードのコメントは変更しない

## コメントのガイドライン

コメントのガイドラインは Korylus 共通です。詳細は [`.claude/rules/common.md`](common.md) の「コメントのガイドライン」セクションを参照してください。

## ログ出力

ログ出力は Go 1.21 で標準ライブラリに追加された `log/slog` パッケージを使用します。

**基本ルール**:

- `log/slog` を使用する（構造化ログで一貫性のあるログ出力を実現）
- `log` パッケージは使用禁止（`log.Printf`, `log.Println`, `log.Fatalf` は使用しない）

**使用する関数**:

| 関数                                           | 用途                               |
| ---------------------------------------------- | ---------------------------------- |
| `slog.Info(msg, key, value, ...)`              | 通常の情報ログ（コンテキストなし） |
| `slog.Warn(msg, key, value, ...)`              | 警告ログ（コンテキストなし）       |
| `slog.Error(msg, key, value, ...)`             | エラーログ（コンテキストなし）     |
| `slog.InfoContext(ctx, msg, key, value, ...)`  | 通常の情報ログ（コンテキストあり） |
| `slog.WarnContext(ctx, msg, key, value, ...)`  | 警告ログ（コンテキストあり）       |
| `slog.ErrorContext(ctx, msg, key, value, ...)` | エラーログ（コンテキストあり）     |

**コンテキストの使い分け**:

- **コンテキストが利用可能な場合**（Handler, UseCase など）: `slog.InfoContext(ctx, ...)` を使用
- **コンテキストが利用不可能な場合**（`main.go` など）: `slog.Info(...)` を使用

**ログレベルの選択基準**:

| レベル       | 用途                                     | 例                            |
| ------------ | ---------------------------------------- | ----------------------------- |
| `slog.Debug` | デバッグ情報（通常は出力しない）         | 変数の値、処理の詳細          |
| `slog.Info`  | 通常の情報（サーバー起動、処理完了など） | サーバー起動、DB 接続成功     |
| `slog.Warn`  | 警告（処理は継続するが注意が必要）       | 非推奨機能の使用、リトライ    |
| `slog.Error` | エラー（処理が失敗した場合）             | DB 接続失敗、API 呼び出し失敗 |

**良い例**:

```go
// コンテキストありの場合（Handler, UseCase）
slog.InfoContext(ctx, "ユーザーがログインしました", "user_id", userID)
slog.ErrorContext(ctx, "パスワードリセットメールの送信に失敗", "error", err, "email", email)

// コンテキストなしの場合（main.go）
slog.Info("サーバーを起動します", "port", cfg.Port)
slog.Error("データベース接続に失敗", "error", err)
```

**悪い例（使用禁止）**:

```go
// log パッケージは使用禁止
log.Printf("ユーザーがログインしました: %d", userID)
log.Println("サーバーを起動します")
log.Fatalf("データベース接続に失敗: %v", err)
```

**致命的エラーの処理**:

`log.Fatalf` の代わりに `slog.Error` + `os.Exit(1)` を使用します。

```go
// 悪い例
log.Fatalf("データベース接続に失敗: %v", err)

// 良い例
slog.Error("データベース接続に失敗", "error", err)
os.Exit(1)
```
