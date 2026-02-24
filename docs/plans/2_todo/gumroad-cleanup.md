# Gumroad関連コード削除 設計書

## 概要

Stripe移行完了後、Gumroad関連のコードとタスクを削除する。

Annictサポーターの決済プラットフォームをGumroadからStripeへ移行した後、Gumroad経由の既存サブスクリプションが全て期限切れになった時点で、Gumroad関連のコードを完全に削除する。

**目的**:

- 不要になったGumroad関連コードを削除し、コードベースをシンプルに保つ
- 使われなくなった同期タスクを停止し、運用負荷を軽減する
- Go版への移行を完了させ、Rails版の依存を減らす

**背景**:

- Stripe移行により、新規サポーター登録はStripe経由のみとなる
- 既存のGumroadサポーターはサブスクリプション期限まで引き続きサポーターとして扱われる
- 全Gumroadサブスクリプションの期限切れ後は、Gumroad関連コードは不要となる

**前提条件**:

- [Annictサポーター Stripe移行](../1_doing/stripe-supporter-migration.md)が完了していること

## 要件

### 機能要件

#### 終了条件の判定

- 全てのGumroadサブスクリプションが期限切れになったことを確認する
- 確認方法: `gumroad_subscribers`テーブルで`gumroad_ended_at`が現在日時より過去のレコードのみであること

```sql
-- アクティブなGumroadサブスクリプションが存在しないことを確認
SELECT COUNT(*) FROM gumroad_subscribers
WHERE (gumroad_cancelled_at IS NULL OR gumroad_cancelled_at > NOW())
  AND (gumroad_ended_at IS NULL OR gumroad_ended_at > NOW());
-- 結果が0であれば、全サブスクリプションが終了済み
```

#### Gumroad同期タスクの停止

- `supporters:sync_with_gumroad` Rakeタスクの定期実行を停止する
- 定期実行の設定を削除する

#### Rails側のGumroad関連コード削除

- Gumroad OmniAuth連携の削除
- Gumroad APIクライアントの削除
- Gumroad関連のモデル・コントローラーの削除

#### Go版のGumroad関連コード削除

- `GumroadSubscriberRepository`の削除
- サポーター判定ロジックからGumroad判定を削除
- サポーターページのGumroad関連表示を削除

#### データベースのクリーンアップ

- `gumroad_subscribers`テーブルの削除（オプション、履歴として残しても可）
- `users.gumroad_subscriber_id`カラムの削除

### 非機能要件

#### 安全性

- 削除前に全Gumroadサブスクリプションが終了済みであることを必ず確認する
- 削除作業はメンテナンス時間帯に行う
- バックアップを取得してから作業する

## 設計

### 削除対象ファイル（Rails版）

| 種別 | ファイル | 説明 |
|-----|---------|------|
| Model | `app/models/gumroad_subscriber.rb` | サブスクリプションモデル |
| Model | `app/models/gumroad_client.rb` | Gumroad APIクライアント |
| Concern | `app/models/concerns/supportable.rb` | 共通ロジック（Gumroad部分のみ削除） |
| Form | `app/models/forms/supporter_registration_form.rb` | 登録フォーム |
| Creator | `app/models/creators/supporter_registration_creator.rb` | 登録処理 |
| Updater | `app/models/updaters/supporter_updater.rb` | 更新処理 |
| Controller | `app/controllers/callbacks_controller.rb` | OAuthコールバック（Gumroad部分のみ削除） |
| Rake | `lib/tasks/supporters.rake` | 同期タスク |
| Config | `config/initializers/omniauth.rb` | OmniAuth設定（Gumroad部分のみ削除） |

### 削除対象ファイル（Go版）

| 種別 | ファイル | 説明 |
|-----|---------|------|
| Repository | `internal/repository/gumroad_subscriber.go` | Gumroadサブスクライバーリポジトリ |
| Query | `internal/query/gumroad_subscriber.sql` | sqlcクエリファイル |
| Template | `internal/templates/pages/supporters/` | Gumroad関連表示部分 |

### データベース変更

```sql
-- gumroad_subscribersテーブルの削除（オプション）
DROP TABLE IF EXISTS gumroad_subscribers;

-- usersテーブルからgumroad_subscriber_idカラムを削除
ALTER TABLE users DROP COLUMN IF EXISTS gumroad_subscriber_id;
```

## タスクリスト

### フェーズ 1: 事前確認

- [ ] **1-1**: Gumroadサブスクリプション終了確認

  - 全Gumroadサブスクリプションが期限切れであることを確認
  - 確認用SQLクエリの実行
  - **想定作業時間**: 約 10 分

### フェーズ 2: 同期タスク停止

- [ ] **2-1**: Gumroad同期タスクの停止

  - 定期実行設定の削除
  - 動作確認
  - **想定作業時間**: 約 15 分

### フェーズ 3: Rails側クリーンアップ

- [ ] **3-1**: Gumroad OmniAuth連携の削除

  - OmniAuthプロバイダー設定からGumroadを削除
  - CallbacksControllerからGumroad処理を削除
  - 関連ルーティングの削除
  - **想定ファイル数**: 約 4 ファイル（実装 4 + テスト 0）
  - **想定行数**: 約 80 行（実装 80 行 + テスト 0 行）

- [ ] **3-2**: Gumroad関連モデル・サービスの削除

  - `GumroadSubscriber`モデルの削除
  - `GumroadClient`の削除
  - 関連フォーム・クリエイター・アップデーターの削除
  - **想定ファイル数**: 約 6 ファイル（実装 6 + テスト 0）
  - **想定行数**: 約 200 行（実装 200 行 + テスト 0 行）

- [ ] **3-3**: User#supporter?からGumroad判定を削除

  - `Supportable` concernからGumroad判定ロジックを削除
  - Stripe判定のみに簡略化
  - テストの更新
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 60 行（実装 30 行 + テスト 30 行）

- [ ] **3-4**: Gumroad同期タスクの削除

  - `lib/tasks/supporters.rake`の削除
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

### フェーズ 4: Go版クリーンアップ

- [ ] **4-1**: GumroadSubscriberRepositoryの削除

  - `internal/repository/gumroad_subscriber.go`の削除
  - 関連sqlcクエリの削除
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 120 行（実装 50 行 + テスト 70 行）

- [ ] **4-2**: サポーター判定ロジックの簡略化

  - `UserRepository.IsSupporter()`からGumroad判定を削除
  - Stripe判定のみに簡略化
  - テストの更新
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 40 行（実装 20 行 + テスト 20 行）

- [ ] **4-3**: サポーターページのGumroad関連表示削除

  - Gumroad移行案内メッセージの削除
  - Gumroadサポーター向け表示分岐の削除
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 80 行（実装 60 行 + テスト 20 行）

### フェーズ 5: データベースクリーンアップ

- [ ] **5-1**: usersテーブルからgumroad_subscriber_idカラムを削除

  - マイグレーションファイルの作成
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 15 行（実装 15 行 + テスト 0 行）

- [ ] **5-2**: gumroad_subscribersテーブルの削除（オプション）

  - 履歴として残す場合はスキップ
  - 削除する場合はマイグレーションファイルを作成
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **Gumroadサブスクリプションデータのアーカイブ**: 必要に応じて別途対応
- **Gumroad関連の履歴データ移行**: 過去の決済履歴などはGumroad側で確認可能

## 参考資料

- [Annictサポーター Stripe移行 設計書](../1_doing/stripe-supporter-migration.md)
