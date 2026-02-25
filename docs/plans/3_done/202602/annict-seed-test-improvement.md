# Annict シードテストの常時実行化 設計書

<!--
このテンプレートの使い方:
1. このファイルを `docs/designs/2_todo/` ディレクトリにコピー
   例: cp docs/designs/template.md docs/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 実装ガイドラインの参照

### Go版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@annict/go/CLAUDE.md](/annict/go/CLAUDE.md) - 全体的なコーディング規約

## 概要

Annict Go版のシードデータ生成テスト（`TestCreateUserUsecase_LargeBatch`、`TestCreateWorkUsecase_LargeBatch`）は、実行時間が長いため `testing.Short()` でCIからスキップされている。これらのテストを高速化し、常にCI上で実行されるようにする。

**目的**:

- シードテストをCIで常時実行し、リグレッションを早期に検出する
- `testing.Short()` によるスキップを廃止し、テストカバレッジの漏れを防ぐ

**背景**:

- 現在 `make test` は `-short` フラグ付きで実行されるため、LargeBatchテストはCIで実行されていない
- シードデータ生成ロジック（マルチ行INSERT、bcryptハッシュ化の並列処理など）は複雑であり、テストによる品質保証が重要
- テストが遅い原因は特定されており、対処可能である

## 要件

### 機能要件

- `TestCreateUserUsecase_LargeBatch` が `-short` フラグの有無に関わらず常に実行される
- `TestCreateWorkUsecase_LargeBatch` が `-short` フラグの有無に関わらず常に実行される
- マルチチャンクでのバッチ処理（100件ずつのマルチ行INSERT）が引き続きテストされる
- bcryptによるパスワードハッシュ化が引き続きテストされる

### 非機能要件

- **パフォーマンス**: 各LargeBatchテストが10秒以内に完了すること
- **保守性**: テスト用bcryptコスト設定が1箇所で管理され、本番コードに影響を与えないこと

## 設計

### 現状の分析

#### ボトルネックの特定

**`TestCreateUserUsecase_LargeBatch`**（主なボトルネック）:

- 2500人のユーザーを作成
- 各ユーザーに対して `bcrypt.GenerateFromPassword` を実行（`bcrypt.DefaultCost` = 10）
- 並列化済み（`runtime.NumCPU() * 2` のgoroutine）だが、bcrypt自体のコストが高い
- 1回のbcrypt（コスト10）≈ 約50-100ms → 2500回 ≈ 数十秒〜数分（並列化してもCPUコア数で割る程度）

**`TestCreateWorkUsecase_LargeBatch`**:

- 2500の作品を作成
- bcryptは不使用
- マルチ行INSERT（100件ずつ）× 25チャンク
- DB操作のみなので比較的高速だが、件数が多いため時間がかかる

#### 現在のコード構造

```
/annict/go/internal/
├── auth/
│   └── password.go          # HashPassword() - bcrypt.DefaultCost を使用
├── usecase/seed/
│   ├── create_user.go       # createMultipleUsers() - auth.HashPassword()を呼び出し
│   ├── create_user_test.go  # TestCreateUserUsecase_LargeBatch - testing.Short()でスキップ
│   ├── create_work.go       # createMultipleWorks() - bcrypt不使用
│   └── create_work_test.go  # TestCreateWorkUsecase_LargeBatch - testing.Short()でスキップ
└── testutil/
    └── db.go                # SetupTestDB() - sync.Onceでコネクションプーリング済み
```

### 改善方針

2つのアプローチを組み合わせて高速化する:

#### 1. テスト用bcryptコスト削減

`auth.HashPassword` にテスト用の低コストオプションを導入する。

**現状**:

```go
// auth/password.go
func HashPassword(plainPassword string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
    // ...
}
```

**改善後**:

```go
// auth/password.go

// bcryptCost はbcryptのコスト値。テスト時はSetBcryptCostForTestで変更可能
var bcryptCost = bcrypt.DefaultCost

// SetBcryptCostForTest はテスト用にbcryptコストを変更する
// テスト以外からの呼び出しは想定していない
func SetBcryptCostForTest(cost int) {
    bcryptCost = cost
}

func HashPassword(plainPassword string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcryptCost)
    // ...
}
```

**テスト側**:

```go
// TestMainまたはテスト内でコストを最小に設定
func TestMain(m *testing.M) {
    auth.SetBcryptCostForTest(bcrypt.MinCost) // コスト4（デフォルト10の約64倍高速）
    os.Exit(m.Run())
}
```

**速度改善効果**:

- bcryptコスト10 → 4 で約64倍高速化（2^(10-4) = 64）
- 2500ユーザー × 50ms → 2500ユーザー × 0.78ms ≈ 約2秒（並列化前）
- 並列化により更に高速化（CPUコア数に応じて）

#### 2. バッチサイズの削減

テスト目的は「マルチチャンクでのバッチ処理が正しく動作すること」であり、2500件である必要はない。

**現状**: 2500件（100件チャンク × 25回）
**改善後**: 250件（100件チャンク × 3回）

250件にする理由:

- 100件チャンクが複数回（3回: 100 + 100 + 50）実行されることを検証できる
- 最後のチャンクが端数（50件）になるケースもカバー
- 2500件と比較して約10倍のINSERT高速化

#### 期待される改善効果

| テスト                   | 現状         | 改善後                            | 高速化      |
| ------------------------ | ------------ | --------------------------------- | ----------- |
| User LargeBatch (2500件) | 数十秒〜数分 | 数秒以内（250件 + bcryptコスト4） | 約100倍以上 |
| Work LargeBatch (2500件) | 数秒〜十数秒 | 1秒以内（250件）                  | 約10倍      |

### 関連設計書

この設計書は [Goテスト実行速度の改善](/workspace/docs/designs/1_doing/go-test-performance.md) のフェーズ4（Annict）と関連しています。特にタスク4-3「テスト用bcryptコスト削減」を前提としています。

本設計書では、bcryptコスト削減をAnnict側で先行実装する形で進めます。

## タスクリスト

### フェーズ 1: bcryptコスト削減とテスト高速化

- [x] **1-1**: [Go] テスト用bcryptコスト設定の導入
  - `auth/password.go` にパッケージ変数 `bcryptCost` と `SetBcryptCostForTest` 関数を追加
  - `HashPassword` 関数がパッケージ変数 `bcryptCost` を使用するように変更
  - `auth/password_test.go` に `SetBcryptCostForTest` のテストを追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 40 行（実装 15 行 + テスト 25 行）

- [x] **1-2**: [Go] シードテストのTestMainでbcryptコスト設定
  - `usecase/seed/` パッケージに `TestMain` を追加し、`auth.SetBcryptCostForTest(bcrypt.MinCost)` を呼び出す
  - 既存の `TestCreateUserUsecase_PasswordHashing` テストがbcrypt検証として引き続き動作することを確認
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 15 行（実装 15 行 + テスト 0 行）

### フェーズ 2: LargeBatchテストの常時実行化

- [x] **2-1**: [Go] LargeBatchテストのバッチサイズ削減と`testing.Short()`除去
  - `TestCreateUserUsecase_LargeBatch` のバッチサイズを2500→250に変更
  - `TestCreateWorkUsecase_LargeBatch` のバッチサイズを2500→250に変更
  - 両テストから `testing.Short()` によるスキップを削除
  - テスト名の変更は不要（テストの意図「大量データのバッチ処理テスト」は変わらない）
  - `make test`（`-short` フラグ付き）で両テストが実行されることを確認
  - **想定ファイル数**: 約 2 ファイル（実装 0 + テスト 2）
  - **想定行数**: 約 10 行（実装 0 行 + テスト 10 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **Makefile の `-short` フラグ削除**: 他のテストにも影響するため、[Goテスト実行速度の改善](/workspace/docs/designs/1_doing/go-test-performance.md) の全体タスクとして管理する
- **他プロジェクト（Wikino、Mewst）への横展開**: 各プロジェクトの状況に応じて個別に対応する
- **本番用bcryptコストの変更**: 本番環境のセキュリティには影響を与えない

## 参考資料

- [Goテスト実行速度の改善 設計書](/workspace/docs/designs/1_doing/go-test-performance.md)
- [bcrypt - Wikipedia](https://en.wikipedia.org/wiki/Bcrypt) - bcryptコストと計算時間の関係
- [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Go bcryptパッケージドキュメント
