# Annictサポーター Stripe移行 設計書

## 概要

AnnictサポーターのサブスクリプションサービスをGumroadからStripeへ移行する。
同時に、サポーター機能をRails版からGo版へ移行する。

現在、Annictサポーターは月額¥290または年額¥2,900で提供されており、Gumroadを通じて決済処理を行っている。
しかし、Gumroadの手数料やサービスの安定性を考慮し、より広く使われているStripeへの移行を行う。

移行期間中は、既存のGumroadサポーターを維持しつつ、新規登録はStripeのみで受け付ける。
Gumroad経由のサブスクリプションは自動更新を停止し、期限が切れた時点でサポーター資格を終了する。
継続を希望するユーザーは、Stripe経由で再登録する必要がある。

**目的**:

- 決済プラットフォームをStripeに統一し、運用コストを削減する
- Stripe Checkout/Customer Portalを活用してユーザー体験を向上させる
- サポーター機能をGo版に移行し、Rails依存を減らす
- 将来的な機能拡張（複数プランなど）への基盤を整備する

**背景**:

- Gumroadの手数料が高い（10%）のに対し、Stripeは3.6%程度
- Gumroadは海外サービスのため、日本のユーザーにとって馴染みが薄い
- Stripeは日本語対応が充実しており、Customer Portalでユーザー自身が管理可能
- Go版への移行が進行中であり、新機能はGo版で実装する方針

## 要件

### 機能要件

#### Stripe決済機能

- ユーザーはStripe Checkoutを通じてサポーターに登録できる
- ユーザーはStripe Customer Portalを通じてサブスクリプションを管理できる（支払い方法変更、キャンセルなど）
- システムはStripe Webhookを通じてサブスクリプションの状態変更を検知し、サポーター状態を更新する

#### プラン構成

- 月額プラン: ¥290/月
- 年額プラン: ¥2,900/年（2ヶ月分無料）
- 支払い方法: クレジットカード
- 通貨: 日本円（JPY）で直接決済

#### プラン変更

ユーザーはStripe Customer Portalを通じて月額⇔年額のプラン変更が可能。

**プラン変更の種類**:

| 変更方向  | 種別           | 適用タイミング | 日割り計算                           |
| --------- | -------------- | -------------- | ------------------------------------ |
| 月額→年額 | アップグレード | 即時           | 月額の残り期間を年額に充当           |
| 年額→月額 | ダウングレード | 請求期間終了時 | 年額の残り期間をクレジットとして保持 |

**Webhook処理**:

- プラン変更は `customer.subscription.updated` イベントで通知される
- `stripe_price_id` の更新のみで対応可能（追加実装不要）

#### サポーター判定

- システムはGumroad経由のサポーターとStripe経由のサポーターの両方を判定できる
- Go版: `UserRepository.IsSupporter()`メソッドは、いずれかの経路でアクティブなサブスクリプションがあれば`true`を返す
- Rails版: 既存の`User#supporter?`メソッドを拡張し、Stripeサポーターも判定できるようにする
  - これにより、Go版への完全移行前でも、Rails版のページ（広告非表示、バッジ表示など）でStripeサポーターが正しく認識される

#### Gumroad移行

- 既存のGumroadサポーターは、サブスクリプション期限まで引き続きサポーターとして扱う
- Gumroad経由の新規登録は停止する（UIからリンクを削除）
- Gumroad同期タスク（`supporters:sync_with_gumroad`）は移行完了まで**Rails側で**継続実行する
- Gumroadサブスクリプションの自動更新は行わない（ユーザーがGumroad側でキャンセルしなくても、期限到来で終了）

#### サポーターページのGo版移行

- サポーターページ（`/supporters`）をGo版で実装する
- リバースプロキシのホワイトリストに`/supporters`を追加する
- 既存のGumroadサポーター情報の表示もGo版で実装する

#### UI変更

- サポーターページ（`/supporters`）にStripe決済ボタンを追加する
- Gumroad連携セクションは削除し、Stripe Customer Portalへのリンクに置き換える
- 既存のGumroadサポーターには、Stripeへの移行を促すメッセージを表示する

**サポーターページの表示条件分岐**:

| ユーザー状態                    | 表示内容                                                           |
| ------------------------------- | ------------------------------------------------------------------ |
| 未ログイン                      | サポーター特典の説明、ログイン促進                                 |
| ログイン済み・非サポーター      | Stripe決済ボタン（月額/年額）                                      |
| Gumroadサポーター（アクティブ） | Gumroad移行案内メッセージ + 次回更新日/終了日の表示                |
| Stripeサポーター（アクティブ）  | サブスクリプション情報 + Customer Portalへのリンク                 |
| Gumroad + Stripe両方アクティブ  | Stripeサブスクリプション情報を優先表示 + Customer Portalへのリンク |

**Gumroad移行案内メッセージ**（サポーターページに表示、メール通知は行わない）:

日本語:

```
サポータープログラムへのご参加ありがとうございます！
現在のサポーター登録はGumroad経由で行われています

Gumroadでの新規登録は終了しており、現在はStripe (決済サービス) での登録に移行しています
現在の契約期間中は引き続きサポーター特典をご利用いただけます。契約終了後も継続される場合は、Gumroadでの契約終了後にこのページからStripeでご登録ください
(Gumroadでの契約終了後、このページにStripeで登録するためのボタンが表示されるようになります)
```

English:

```
Thank you for joining the supporter program!
Your current registration is through Gumroad

New registrations via Gumroad have ended, and we have transitioned to Stripe (payment service) for new subscriptions
You can continue to enjoy supporter benefits during your current subscription period. If you wish to continue after your subscription ends, please register with Stripe from this page after your Gumroad subscription ends
(After your Gumroad subscription ends, a button to register with Stripe will appear on this page)
```

**Checkout完了/キャンセル時のメッセージ表示**:

クエリパラメータに基づいてページ上部にメッセージを表示する：

| パラメータ       | メッセージ（日本語）                     | メッセージ（英語）                    |
| ---------------- | ---------------------------------------- | ------------------------------------- |
| `?success=true`  | 「サポーター登録ありがとうございます！」 | "Thank you for becoming a supporter!" |
| `?canceled=true` | 「決済がキャンセルされました」           | "Payment was canceled"                |

### 非機能要件

#### セキュリティ

- Stripe Webhookの署名検証を必ず行い、不正なリクエストを拒否する
- Stripe API KeyはSecretsとして管理し、ソースコードにコミットしない
- Customer Portalへのリダイレクトは認証済みユーザーのみ許可する

#### 可用性・信頼性

- Webhook処理の冪等性を確保する（同じイベントを複数回受信しても問題ない）
- Webhook処理に失敗した場合は、Stripeの自動リトライに任せる
- 決済処理中のエラーは適切にハンドリングし、ユーザーにわかりやすいメッセージを表示する

#### 監視・エラー通知

**Webhook処理失敗時の通知**:

- Sentryへエラー送信（既存のSentry連携を使用）
- エラー内容: イベントタイプ、stripe_event_id、エラーメッセージ

**Stripeダッシュボードのアラート設定**:

- Settings > Communication preferences でアカウント通知を有効化
  - 支払い関連通知
  - 紛争（Dispute）通知
  - セキュリティ/リスク警告
- Developers > Webhooks で配信状況を定期確認

#### 障害時の対応方針

**基本方針**: ロールバックは行わず、問題を特定して追加修正で対応する。

**理由**:

- Stripe側で作成されたサブスクリプションはDBロールバックしても消えないため、不整合が発生するリスクがある
- サポーター機能（広告非表示、バッジ表示など）はクリティカルではなく、一時的な問題があっても手動でサポーター状態を維持可能
- DBマイグレーションのロールバックはデータ損失リスクがあり、問題を修正する方が安全

**問題発生時の対応**:

| 問題                             | 対応方法                                                                                                        |
| -------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| Webhook処理失敗                  | `stripe_webhook_events`テーブルで失敗イベントを確認し、手動で再処理または`stripe_subscribers`レコードを直接作成 |
| サポーター判定が正しくない       | `users.stripe_subscriber_id`と`stripe_subscribers.stripe_status`を確認・修正                                    |
| Checkout後にサポーターにならない | Stripeダッシュボードでサブスクリプション状態を確認し、`stripe_subscribers`レコードを手動作成                    |
| 決済は成功したがDBに反映されない | Stripeダッシュボードのイベントログを確認し、`stripe_subscribers`レコードを手動作成                              |

**手動でのサポーター状態維持**:

```sql
-- サポーターとして手動登録（緊急時）
INSERT INTO stripe_subscribers (stripe_customer_id, stripe_subscription_id, stripe_price_id, stripe_status, ...)
VALUES ('cus_xxx', 'sub_xxx', 'price_xxx', 'active', ...);

UPDATE users SET stripe_subscriber_id = <新しいID> WHERE id = <ユーザーID>;
```

#### 支払い失敗時の対応

**Stripeの自動リトライ（Smart Retries）**:

- Stripeは支払い失敗時に自動的にリトライを行う（デフォルト: 2週間で最大8回）
- リトライ中、サブスクリプションは`past_due`状態となる
- `past_due`状態でもサポーター機能は継続して利用可能（猶予期間）
- 全リトライ失敗後、Stripe側で`canceled`に自動変更される

**ステータス遷移（Stripe側で自動実行）**:

```
active → (支払い失敗) → past_due → (リトライ成功) → active
                                  → (リトライ全失敗) → canceled
```

**メール通知（Stripe組み込み機能を使用）**:

- Annict側での実装は不要
- Stripe Dashboard (Settings > Emails > Subscriptions emails) で設定
- 「Failed payment」メールを有効化することで、支払い失敗時に自動通知

**サポーター機能の利用可否**:

| stripe_status | サポーター機能 | 説明                               |
| ------------- | -------------- | ---------------------------------- |
| `active`      | 利用可能       | 通常状態                           |
| `past_due`    | 利用可能       | 支払い遅延中だが猶予期間として継続 |
| `canceled`    | 利用不可       | キャンセル済み                     |
| `unpaid`      | 利用不可       | 未払い（設定による）               |

#### 保守性

- Go版の既存アーキテクチャ（Handler、UseCase、Repository）に従う
- テストは実データベースを使用し、Stripe APIはモックする

## 設計

### 技術スタック

- **決済プラットフォーム**: Stripe
- **Goライブラリ**: `github.com/stripe/stripe-go/v82`（公式）
- **決済UI**: Stripe Checkout（ホスト型決済ページ）
- **サブスク管理UI**: Stripe Customer Portal

#### Stripe Checkout vs Payment Links

Stripeには複数の決済導入方法があるが、本実装ではStripe Checkoutを採用する。

| 機能           | Stripe Checkout              | Payment Links           |
| -------------- | ---------------------------- | ----------------------- |
| セッション作成 | API（プログラマティック）    | ダッシュボードでURL生成 |
| metadata設定   | user_idなど自由に設定可能    | client_reference_idのみ |
| Webhook連携    | metadataでユーザー特定が容易 | 紐付けが難しい          |
| 用途           | アプリ組み込み型決済         | シンプルな販売・寄付    |

**Checkoutを選択した理由**:

Annictではログインユーザーとサブスクリプションを紐付ける必要がある。
Checkoutセッション作成時に`metadata`に`user_id`を含めることで、
Webhook受信時にどのユーザーがサポーターになったかを特定できる。

```go
// Checkoutセッション作成時
session, _ := checkout.Session.New(&stripe.CheckoutSessionParams{
    Metadata: map[string]string{
        "user_id": strconv.FormatInt(currentUser.ID, 10),
    },
    // ...
})
```

Payment Linksはノーコードで簡単に作成できるが、ユーザーとの紐付けが難しいため不採用。

### アーキテクチャ

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層                                          │
│ - internal/handler/supporters/     (サポーターページ)    │
│ - internal/handler/webhooks/stripe/ (Webhook受信)       │
│ - internal/templates/pages/supporters/ (テンプレート)    │
└─────────────────────────────────────────────────────────┘
         ↓ 依存
┌─────────────────────────────────────────────────────────┐
│ Application層                                           │
│ - internal/usecase/create_stripe_subscriber.go          │
│ - internal/usecase/update_stripe_subscriber.go          │
└─────────────────────────────────────────────────────────┘
         ↓ 依存
┌─────────────────────────────────────────────────────────┐
│ Domain/Infrastructure層                                  │
│ - internal/repository/stripe_subscriber.go              │
│ - internal/repository/gumroad_subscriber.go             │
│ - internal/query/ (sqlc生成コード)                       │
│ - internal/stripe/ (Stripe APIクライアント)              │
└─────────────────────────────────────────────────────────┘
```

### データベース設計

#### stripe_subscribersテーブル（新規作成）

```sql
CREATE TABLE public.stripe_subscribers (
    id bigint NOT NULL,
    stripe_customer_id character varying NOT NULL,      -- Stripeの顧客ID (cus_xxx)
    stripe_subscription_id character varying NOT NULL,  -- StripeのサブスクリプションID (sub_xxx)
    stripe_price_id character varying NOT NULL,         -- Stripeの価格ID (price_xxx)
    stripe_status character varying NOT NULL,           -- サブスクリプション状態 (active, canceled, past_due, etc.)
    stripe_current_period_start timestamp without time zone NOT NULL,  -- 現在の請求期間開始
    stripe_current_period_end timestamp without time zone NOT NULL,    -- 現在の請求期間終了
    stripe_cancel_at timestamp without time zone,       -- キャンセル予定日（期間終了時にキャンセルの場合）
    stripe_canceled_at timestamp without time zone,     -- 実際にキャンセルされた日時
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);

-- インデックス
CREATE UNIQUE INDEX index_stripe_subscribers_on_stripe_customer_id ON public.stripe_subscribers USING btree (stripe_customer_id);
CREATE UNIQUE INDEX index_stripe_subscribers_on_stripe_subscription_id ON public.stripe_subscribers USING btree (stripe_subscription_id);
CREATE INDEX index_stripe_subscribers_on_stripe_status ON public.stripe_subscribers USING btree (stripe_status);
```

#### usersテーブルへの変更

```sql
-- stripe_subscriber_id カラムを追加
ALTER TABLE public.users ADD COLUMN stripe_subscriber_id bigint;
CREATE INDEX index_users_on_stripe_subscriber_id ON public.users USING btree (stripe_subscriber_id);
ALTER TABLE public.users ADD CONSTRAINT fk_users_stripe_subscriber FOREIGN KEY (stripe_subscriber_id) REFERENCES public.stripe_subscribers(id);
```

#### stripe_webhook_eventsテーブル（新規作成）

Webhook処理の冪等性を確保し、イベントの処理状態を管理するためのテーブル。

```sql
CREATE TABLE public.stripe_webhook_events (
    id bigint NOT NULL,
    stripe_event_id character varying NOT NULL,           -- [Stripe] イベントID (evt_xxx)
    stripe_event_type character varying NOT NULL,         -- [Stripe] イベントタイプ (checkout.session.completed など)
    stripe_payload jsonb NOT NULL,                        -- [Stripe] イベント全体のJSONデータ（調査・デバッグ用）
    status character varying NOT NULL DEFAULT 'pending',  -- [Annict] 処理状態 (pending, processed, failed, skipped)
    error_message text,                                   -- [Annict] エラー時のメッセージ
    received_at timestamp without time zone NOT NULL,     -- [Annict] 受信日時
    processed_at timestamp without time zone,             -- [Annict] 処理完了日時
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);

-- インデックス
CREATE UNIQUE INDEX index_stripe_webhook_events_on_stripe_event_id ON public.stripe_webhook_events USING btree (stripe_event_id);
CREATE INDEX index_stripe_webhook_events_on_stripe_event_type ON public.stripe_webhook_events USING btree (stripe_event_type);
CREATE INDEX index_stripe_webhook_events_on_status ON public.stripe_webhook_events USING btree (status);
CREATE INDEX index_stripe_webhook_events_on_received_at ON public.stripe_webhook_events USING btree (received_at);
```

**statusの値**:

| 値          | 説明                             |
| ----------- | -------------------------------- |
| `pending`   | 受信済み、処理待ち               |
| `processed` | 処理完了                         |
| `failed`    | 処理失敗                         |
| `skipped`   | 処理対象外のイベント（ログのみ） |

### API設計

#### Stripe Webhook エンドポイント

```
POST /webhooks/stripe
```

**処理するイベント**:

| イベント                        | 処理内容                                   |
| ------------------------------- | ------------------------------------------ |
| `checkout.session.completed`    | サブスクリプション作成、ユーザーとの紐付け |
| `customer.subscription.updated` | サブスクリプション状態の更新               |
| `customer.subscription.deleted` | サブスクリプションの終了処理               |
| `invoice.payment_succeeded`     | 支払い成功の記録（ログ用）                 |
| `invoice.payment_failed`        | 支払い失敗の通知（オプション）             |

**Webhook処理フロー（冪等性の確保）**:

```
1. リクエスト受信
2. Stripe署名検証（失敗なら400を返す）
3. stripe_event_idでstripe_webhook_eventsを検索
   - 既存レコードあり → 200を返して終了（冪等性）
   - 既存レコードなし → 次へ
4. stripe_webhook_eventsにレコード作成（status=pending）
5. イベントタイプに応じた処理を実行
   - 処理対象外のイベント → status=skipped
   - 処理成功 → status=processed, processed_at=now
   - 処理失敗 → status=failed, error_message=エラー内容
6. 200を返す（Stripeへの応答）
```

**トランザクション処理**:

- サブスクリプション処理内のDB操作（StripeSubscriber作成/更新 + User紐付け）は同一DBトランザクション内で実行
- イベント保存・ステータス更新は冪等性チェックにより整合性を担保
  - 処理開始前: stripe_event_idで既存レコードを検索し、処理済み（processed/skipped）ならスキップ
  - 処理完了後: statusをprocessed/failed/skippedに更新
- 外部API呼び出し（Stripe API）はトランザクション外で実行
  - 理由: トランザクション内で外部APIを呼ぶと、API遅延やタイムアウトがDBコネクションを長時間占有してしまうため
- 処理が失敗した場合もイベントレコードは保持（status=failedで記録）
- Annictのサブスクリプション処理は比較的シンプルなため、同期処理で実装（非同期キューは不要）

#### サポーターページ

```
GET /supporters
```

- 認証任意（未ログインでも閲覧可能、ログイン時は追加情報表示）
- サポーター特典の説明
- Stripe決済ボタン（月額/年額）
- ログイン済みサポーターには管理情報を表示

#### Stripe Checkout セッション作成

```
POST /supporters/checkout
```

- 認証必須
- リクエスト: `plan=monthly` or `plan=yearly`
- レスポンス: Stripe CheckoutページへのリダイレクトURL

**重複サブスクリプション防止**:

Checkoutセッション作成時に以下のチェックを行い、重複登録を防止する：

1. ユーザーに紐づく`stripe_subscriber_id`が存在し、アクティブな場合はエラーを返す
2. Stripe APIで既存サブスクリプションを確認（バックアップ）

```go
// 既存サブスクリプションのチェック例
params := &stripe.SubscriptionListParams{
    Customer: stripe.String(stripeCustomerID),
}
params.Filters.AddFilter("status", "", "active")
iter := subscription.List(params)
if iter.Next() {
    // 既にアクティブなサブスクリプションが存在する
    // Customer Portalへリダイレクトするか、エラーメッセージを表示
}
```

**Stripe Dashboard設定（バックアップ）**:

Settings > Payments > Checkout and Payment links で、既存サブスクリプション保持者のリダイレクト設定を有効化する。これにより、アプリ側チェックをすり抜けた場合でも重複を防止できる。

**成功/キャンセル時のリダイレクトURL**:

Checkoutセッション作成時に以下のURLを設定する：

| パラメータ    | URL                         | 用途                             |
| ------------- | --------------------------- | -------------------------------- |
| `success_url` | `/supporters?success=true`  | 決済完了後のリダイレクト先       |
| `cancel_url`  | `/supporters?canceled=true` | 決済キャンセル時のリダイレクト先 |

```go
// Checkoutセッション作成時の設定例
params := &stripe.CheckoutSessionParams{
    SuccessURL: stripe.String(cfg.BaseURL + "/supporters?success=true"),
    CancelURL:  stripe.String(cfg.BaseURL + "/supporters?canceled=true"),
    // ...
}
```

#### Stripe Customer Portal

```
POST /supporters/portal
```

- 認証必須
- Stripe Customer Portalへのリダイレクト

### コード設計

#### ディレクトリ構造

```
internal/
├── handler/
│   ├── supporters/
│   │   ├── handler.go      # Handler構造体と依存性
│   │   └── show.go         # GET /supporters
│   ├── supporters_checkout/
│   │   ├── handler.go      # Handler構造体と依存性
│   │   ├── create.go       # POST /supporters/checkout
│   │   └── request.go      # CreateRequest DTO
│   ├── supporters_portal/
│   │   ├── handler.go      # Handler構造体と依存性
│   │   └── create.go       # POST /supporters/portal
│   └── webhooks/
│       └── stripe/
│           ├── handler.go  # Handler構造体
│           └── create.go   # POST /webhooks/stripe
├── usecase/
│   ├── create_stripe_subscriber.go
│   ├── update_stripe_subscriber.go
│   └── delete_stripe_subscriber.go
├── repository/
│   ├── stripe_subscriber.go
│   ├── stripe_webhook_event.go
│   └── gumroad_subscriber.go
├── stripe/
│   └── client.go           # Stripe APIラッパー
└── templates/
    └── pages/
        └── supporters/
            ├── show.templ
            └── show_templ.go
```

#### Repository

```go
// internal/repository/stripe_subscriber.go
type StripeSubscriberRepository struct {
    queries *query.Queries
}

func (r *StripeSubscriberRepository) Create(ctx context.Context, params CreateStripeSubscriberParams) (*query.StripeSubscriber, error)
func (r *StripeSubscriberRepository) GetByStripeCustomerID(ctx context.Context, customerID string) (*query.StripeSubscriber, error)
func (r *StripeSubscriberRepository) GetByStripeSubscriptionID(ctx context.Context, subscriptionID string) (*query.StripeSubscriber, error)
func (r *StripeSubscriberRepository) Update(ctx context.Context, id int64, params UpdateStripeSubscriberParams) error

// IsActive は active または past_due 状態をアクティブとして扱う
// past_due は支払い遅延中だが、Stripeがリトライ中のため猶予期間として利用可能
func (r *StripeSubscriberRepository) IsActive(s *query.StripeSubscriber) bool {
    return s.StripeStatus == "active" || s.StripeStatus == "past_due"
}

// internal/repository/gumroad_subscriber.go
type GumroadSubscriberRepository struct {
    queries *query.Queries
}

func (r *GumroadSubscriberRepository) GetByID(ctx context.Context, id int64) (*query.GumroadSubscriber, error)
func (r *GumroadSubscriberRepository) IsActive(subscriber *query.GumroadSubscriber) bool
```

#### UseCase

```go
// internal/usecase/create_stripe_subscriber.go
type CreateStripeSubscriberUsecase struct {
    stripeSubscriberRepo *repository.StripeSubscriberRepository
    userRepo             *repository.UserRepository
    db                   *sql.DB
}

func (u *CreateStripeSubscriberUsecase) Execute(ctx context.Context, input CreateStripeSubscriberInput) error
```

#### Handler

```go
// internal/handler/supporters/handler.go
// サポーターページ表示用Handler
type Handler struct {
    cfg                   *config.Config
    sessionManager        *session.Manager
    imageHelper           *image.Helper
    stripeSubscriberRepo  *repository.StripeSubscriberRepository
    gumroadSubscriberRepo *repository.GumroadSubscriberRepository
    stripeCfg             *annictStripe.Config
    stripeClient          *stripe.Client
}

// Show handles GET /supporters
func (h *Handler) Show(w http.ResponseWriter, r *http.Request)
```

```go
// internal/handler/supporters_checkout/handler.go
// Stripe Checkoutセッション作成用Handler
type Handler struct {
    cfg                  *config.Config
    sessionManager       *session.Manager
    stripeSubscriberRepo *repository.StripeSubscriberRepository
    stripeCfg            *annictStripe.Config
    stripeClient         *stripe.Client
}

// Create handles POST /supporters/checkout
func (h *Handler) Create(w http.ResponseWriter, r *http.Request)
```

```go
// internal/handler/supporters_portal/handler.go
// Stripe Customer Portal用Handler
type Handler struct {
    cfg                  *config.Config
    sessionManager       *session.Manager
    stripeSubscriberRepo *repository.StripeSubscriberRepository
    stripeClient         *stripe.Client
}

// Create handles POST /supporters/portal
func (h *Handler) Create(w http.ResponseWriter, r *http.Request)
```

#### サポーター判定ロジック

```go
// internal/repository/user.go に追加
func (r *UserRepository) IsSupporter(ctx context.Context, user *query.User) (bool, error) {
    // Stripeサブスクリプションをチェック
    if user.StripeSubscriberID.Valid {
        stripeSubscriber, err := r.stripeSubscriberRepo.GetByID(ctx, user.StripeSubscriberID.Int64)
        if err == nil && r.stripeSubscriberRepo.IsActive(stripeSubscriber) {
            return true, nil
        }
    }

    // Gumroadサブスクリプションをチェック（移行期間中）
    if user.GumroadSubscriberID.Valid {
        gumroadSubscriber, err := r.gumroadSubscriberRepo.GetByID(ctx, user.GumroadSubscriberID.Int64)
        if err == nil && r.gumroadSubscriberRepo.IsActive(gumroadSubscriber) {
            return true, nil
        }
    }

    return false, nil
}
```

### Stripeプロダクト設定

Stripeダッシュボードで以下を設定:

1. **Product**: "Annict Supporters"
2. **Prices**:
   - Monthly: ¥290/月（税込、`tax_behavior: inclusive`）
   - Yearly: ¥2,900/年（税込、`tax_behavior: inclusive`）
3. **Customer Portal設定**:
   - サブスクリプションのキャンセル許可
   - 支払い方法の変更許可
   - **請求履歴の表示**（領収書PDFダウンロード可能）
   - **プランの変更を許可**（月額⇔年額の切り替え）
   - プラン変更時の日割り計算（proration）を有効化
   - ダウングレードは請求期間終了時に適用（「ダウングレードを管理」を有効化）

### 税金・インボイス設定

**消費税の扱い**:

- **Stripe Taxは使用しない**: 日本国内のみ、単一税率（10%）のため不要
- **内税（税込価格）**: Price作成時に `tax_behavior: inclusive` を設定
- 表示価格（¥290、¥2,900）がそのまま請求額となる

**インボイス（領収書）**:

- Customer Portalで請求履歴の確認・領収書PDFダウンロードを有効化
- Stripe Billingの標準機能を使用（追加実装不要）
- 設定: Dashboard > Settings > Customer portal > Invoice history を有効化

### 環境変数

```
ANNICT_STRIPE_SECRET_KEY=sk_xxx          # Stripe Secret Key
ANNICT_STRIPE_WEBHOOK_SECRET=whsec_xxx   # Webhook署名検証用シークレット
ANNICT_STRIPE_PRICE_MONTHLY_ID=price_xxx # 月額プランの価格ID
ANNICT_STRIPE_PRICE_YEARLY_ID=price_yyy  # 年額プランの価格ID
```

### テスト環境

開発・テスト環境ではStripe Test Modeを使用し、本番環境ではLive Modeを使用する。

**Test Mode vs Live Mode**:

| 項目         | Test Mode（開発・テスト）          | Live Mode（本番） |
| ------------ | ---------------------------------- | ----------------- |
| APIキー      | `sk_test_xxx`                      | `sk_live_xxx`     |
| 決済         | シミュレーション（実際の課金なし） | 実際の課金        |
| テストカード | `4242 4242 4242 4242` 等を使用可能 | 実際のカードのみ  |
| データ       | 本番データとは完全に分離           | 本番データ        |

**環境ごとの設定**:

| 環境                 | Mode      | 用途               |
| -------------------- | --------- | ------------------ |
| 開発環境（ローカル） | Test Mode | 開発・デバッグ     |
| 本番環境             | Live Mode | 実際のサービス提供 |

**テスト環境のセットアップ手順**:

1. Stripe Dashboardで「Test mode」に切り替え
2. Test Mode用のProduct/Priceを作成
3. Test Mode用のAPIキーを取得（`sk_test_xxx`）
4. Customer Portalの設定（Test Mode用）
5. Webhook endpointの登録（開発環境URL or Stripe CLI）

**ローカル開発でのWebhookテスト**:

Stripe CLIを使用してローカル環境でWebhookをテストできる:

```sh
# Stripe CLIでWebhookをローカルに転送
stripe listen --forward-to localhost:8080/webhooks/stripe

# Webhook署名シークレットが表示される（whsec_xxx）
# この値を環境変数に設定して使用
```

**テスト用クレジットカード番号**:

| カード番号            | 用途                 |
| --------------------- | -------------------- |
| `4242 4242 4242 4242` | 成功する決済         |
| `4000 0000 0000 0002` | カード拒否           |
| `4000 0000 0000 3220` | 3Dセキュア認証が必要 |

**環境変数の管理**:

- 開発環境: `.env` ファイルにTest Mode APIキーを設定
- 本番環境: Dokku環境変数にLive Mode APIキーを設定
- Price IDは環境ごとに異なる（Test/Live Modeで別々に作成されるため）

### リバースプロキシ設定の変更

`internal/middleware/reverse_proxy.go` のホワイトリストに追加:

```go
var goHandledPaths = []string{
    // ... 既存のパス
    "/supporters",        // サポーターページ
    "/webhooks/stripe",   // Stripe Webhook
}
```

## タスクリスト

### フェーズ 1: インフラ準備

- [x] **1-1**: stripe-go ライブラリの追加と初期設定
  - `go get github.com/stripe/stripe-go/v82`
  - `internal/stripe/client.go` を作成
  - 環境変数の設定（`.env.example` 更新）
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

- [x] **1-2**: stripe_subscribersテーブルの作成とsqlc生成
  - dbmateマイグレーションファイルの作成
  - sqlcクエリファイルの作成
  - `sqlc generate` でGoコード生成
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 80 行（実装 80 行 + テスト 0 行）

- [x] **1-3**: stripe_webhook_eventsテーブルの作成とsqlc生成
  - dbmateマイグレーションファイルの作成（冪等性確保用テーブル）
  - sqlcクエリファイルの作成（Create, GetByStripeEventID, UpdateStatus）
  - `sqlc generate` でGoコード生成
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 60 行（実装 60 行 + テスト 0 行）

- [x] **1-4**: StripeSubscriberRepository の作成
  - `internal/repository/stripe_subscriber.go` を作成
  - CRUD操作と `IsActive()` メソッドの実装
  - テストの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 60 行 + テスト 90 行）

- [x] **1-4b**: StripeWebhookEventRepository の作成
  - `internal/repository/stripe_webhook_event.go` を作成
  - Create, GetByStripeEventID, UpdateStatus メソッドの実装
  - MarkAsProcessed, MarkAsFailed, MarkAsSkipped, Exists ヘルパーメソッドの実装
  - Webhook処理の冪等性確保に必要
  - テストの作成
  - テストヘルパー StripeWebhookEventBuilder の追加（fixtures.go）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 120 行（実装 50 行 + テスト 70 行）

- [x] **1-5**: usersテーブルへのstripe_subscriber_id追加
  - dbmateマイグレーション（stripe_subscriber_idカラム追加）
  - sqlcクエリの更新
  - UserRepository に関連メソッド追加
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 80 行（実装 30 行 + テスト 50 行）

- [x] **1-6**: Rails版User#supporter?のStripe対応
  - Rails版Userモデルに`belongs_to :stripe_subscriber`追加
  - `supporter?`メソッドを拡張してStripe判定も含める
  - StripeSubscriberモデルの作成（Go版と同じテーブルを参照）
  - 広告非表示、バッジ表示などの既存機能がStripeサポーターでも動作することを確認
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [x] **1-7**: Stripe Dashboard設定
  - **自動リトライ設定の確認** (Settings > Billing > Subscriptions and emails):
    - Smart Retriesの有効化を確認
    - リトライスケジュールの確認（デフォルト: 2週間で最大8回）
    - 全リトライ失敗後の動作を「Cancel subscription」に設定
  - **メール通知の設定** (Settings > Emails > Subscriptions emails):
    - 「Failed payment」メールを有効化
    - メール文面のカスタマイズ（必要に応じて）
  - **Customer Portal設定** (Settings > Billing > Customer portal):
    - 「プランの変更」を有効化（月額⇔年額の切り替え許可）
    - 「サブスクリプションの更新に比例配分を適用」を有効化
    - 「ダウングレードを管理」を有効化（請求期間終了時に適用）
    - 「Invoice history」を有効化（請求履歴・領収書PDFダウンロード）
  - **想定作業時間**: 約 20 分

- [x] **1-8**: Stripe Test Mode環境のセットアップ
  - **Stripe Dashboardでの設定（Test Mode）**:
    - Test Modeに切り替え
    - Product「Annict Supporters」を作成
    - Price（月額¥290、年額¥2,900）を作成
    - Customer Portalの設定（1-7と同様の設定をTest Modeで）
  - **開発環境の設定**:
    - Test Mode APIキー（`sk_test_xxx`）を取得
    - `.env.example` にTest Mode用の設定例を追加
    - `.env` にTest Mode APIキーを設定
  - **Stripe CLIのセットアップ**:
    - Stripe CLIのインストール確認
    - `stripe login` でアカウント連携
    - Webhook転送のテスト（`stripe listen --forward-to localhost:8080/webhooks/stripe`）
  - **想定作業時間**: 約 30 分

- [x] **1-9**: Stripeダッシュボードのアラート設定確認
  - **Communication preferences** (Settings > Communication preferences):
    - 支払い関連通知の有効化確認
    - 紛争（Dispute）通知の有効化確認
    - セキュリティ/リスク警告の有効化確認
  - **Webhook配信状況の確認方法**:
    - Developers > Webhooks で配信状況を確認する手順の確認
    - 失敗時の再送方法の確認
  - **想定作業時間**: 約 15 分

### フェーズ 2: Gumroad既存機能のGo版移行

- [x] **2-1**: GumroadSubscriberRepository の作成
  - `internal/repository/gumroad_subscriber.go` を作成
  - `GetByID()` と `IsActive()` メソッドの実装
  - sqlcクエリの追加
  - テストの作成
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 120 行（実装 50 行 + テスト 70 行）

- [x] **2-2**: サポーター判定ロジックの実装
  - UserRepository に `IsSupporter()` メソッド追加
  - Gumroad/Stripe両方の判定に対応
  - テストの作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 40 行 + テスト 60 行）

- [x] **2-3**: サポーターページのGo版基本実装
  - `internal/handler/supporters/` ディレクトリ作成
  - `show.go` で基本的なページ表示を実装
  - templテンプレートの作成（特典説明、ユーザー状態に応じた表示分岐）
  - リバースプロキシのホワイトリストに `/supporters` 追加
  - **表示条件分岐の実装**:
    - 未ログイン: サポーター特典の説明、ログイン促進
    - ログイン済み・非サポーター: Stripe決済ボタン
    - Gumroadサポーター: 移行不要の案内メッセージ + 次回更新日/終了日
    - Stripeサポーター: サブスクリプション情報 + Customer Portalリンク
  - **Checkout完了/キャンセル時のメッセージ表示**:
    - `?success=true`: 「サポーター登録ありがとうございます！」
    - `?canceled=true`: 「決済がキャンセルされました」
  - **想定ファイル数**: 約 5 ファイル（実装 4 + テスト 1）
  - **想定行数**: 約 320 行（実装 240 行 + テスト 80 行）

### フェーズ 3: Stripe決済機能

- [x] **3-1**: Stripe Webhook受信エンドポイントの作成
  - `internal/handler/webhooks/stripe/` ディレクトリ作成
  - 署名検証の実装
  - **冪等性処理の実装**:
    - stripe_event_idでstripe_webhook_eventsを検索
    - 既存レコードがあれば200を返して終了
    - 新規ならレコード作成後にイベント処理
    - 処理結果に応じてstatus更新（processed/failed/skipped）
  - **エラー通知の実装**:
    - 処理失敗時にSentryへエラー送信
    - エラー内容: イベントタイプ、stripe_event_id、エラーメッセージ
  - イベントディスパッチャーの基本構造
  - リバースプロキシのホワイトリストに `/webhooks/stripe` 追加
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 2）
  - **想定行数**: 約 300 行（実装 130 行 + テスト 170 行）

- [x] **3-2**: checkout.session.completedイベント処理
  - `CreateStripeSubscriberUsecase` の作成
  - StripeSubscriberレコード作成ロジック
  - Userとの紐付け処理
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 180 行（実装 70 行 + テスト 110 行）

- [x] **3-3**: customer.subscription.updated/deletedイベント処理
  - `UpdateStripeSubscriberUsecase` の作成
  - サブスクリプション状態更新ロジック
  - キャンセル処理（Userとの紐付け解除）
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 180 行（実装 70 行 + テスト 110 行）

- [x] **3-4**: Stripe Checkoutセッション作成
  - `internal/handler/supporters/checkout.go` の作成
  - Stripe Checkout Session API呼び出し
  - metadataにuser_idを含める（Webhook処理用）
  - **重複サブスクリプション防止チェックの実装**:
    - ユーザーに紐づく`stripe_subscriber_id`が存在しアクティブな場合はエラーを返す
    - Stripe APIで既存サブスクリプションを確認（バックアップ）
    - 既存サブスクリプションがある場合はCustomer Portalへリダイレクト
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 200 行（実装 80 行 + テスト 120 行）

- [x] **3-5**: Stripe Customer Portal連携
  - `internal/handler/supporters/portal.go` の作成
  - Stripe Billing Portal Session API呼び出し
  - 認証チェック（Stripeサポーターのみ）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 40 行 + テスト 60 行）

### フェーズ 4: UI更新

- [x] **4-1**: サポーターページのStripe決済ボタン追加
  - 月額/年額プラン選択UI
  - Checkoutボタンの追加
  - プラン比較表示
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [x] **4-2**: Gumroadサポーター向け表示の実装
  - Gumroadサポーター（アクティブ）向けの移行案内メッセージ表示
    - 日本語・英語両方のメッセージを用意（i18n対応）
    - 次回更新日/終了日の表示
  - Gumroad情報の取得と表示（GumroadSubscriberRepository経由）
  - メール通知は不要（サポーターページでの表示のみ）
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 80 行（実装 80 行 + テスト 0 行）

- [x] **4-3**: Stripeサポーター情報表示セクションの追加
  - サブスクリプション状態の表示
  - Customer Portalへのリンク
  - 次回請求日の表示
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

- [x] **4-4**: 国際化対応（日本語/英語）
  - `internal/i18n/locales/ja.toml` にサポーター関連メッセージ追加
  - `internal/i18n/locales/en.toml` にサポーター関連メッセージ追加
  - テンプレートでの翻訳適用
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

### フェーズ 5: Rails側クリーンアップ

- [x] **5-1**: Rails側のサポーターページの削除
  - Go版のリバースプロキシが`/supporters`をホワイトリストに含んでいるため、リダイレクト設定は不要
  - Rails側のサポーターページ関連ファイルを削除:
    - `app/controllers/supporters_controller.rb`
    - `app/views/supporters/` ディレクトリ
    - `spec/requests/supporters/` ディレクトリ
    - `config/routes.rb` の `/supporters` ルーティング
    - `config/locales/` のサポーターページ関連翻訳
  - **削除ファイル数**: 5 ファイル + 関連ルーティング・翻訳

- [x] **5-2**: Rails側のGumroad OmniAuth連携の無効化
  - OmniAuth Gumroadプロバイダーの削除
  - CallbacksControllerからGumroad処理を削除
  - 関連ルーティングの削除
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

- [x] **5-3**: 特定商取引法に基づく表記の更新
  - `app/views/pages/legal.html.erb` を更新
  - 「お支払い方法」から「Gumroad」を削除
    - 変更前: 「決済サービス「Stripe」または「Gumroad」を利用したクレジットカード決済」
    - 変更後: 「決済サービス「Stripe」を利用したクレジットカード決済」
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 5 行（実装 5 行 + テスト 0 行）

### フェーズ 5a: サポーターページUI/UX改善

- [x] **5a-1**: レスポンシブ対応
  - モバイル表示時にサイドバーを非表示にする
  - モバイル表示時に下部にタブバーを表示する
  - 既存のGo版レイアウトのレスポンシブパターンに従う
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

- [x] **5a-2**: フッターのデザイン調整
  - サポーターページのフッターデザインを調整
  - 他のGo版ページとデザインを統一
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

- [x] **5a-3**: メインコンテンツのデザイン調整
  - サポーターページのメインコンテンツ領域のデザインを調整
  - カード、ボタン、テキストなどのスタイリングを改善
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 80 行（実装 80 行 + テスト 0 行）

- [x] **5a-4**: サイドバーの通知リンクに未読バッジを表示
  - Go版サイドバーの通知リンクに未読通知数のバッジを表示
  - Rails版 `main_sidebar_component.rb` と同様の実装
    - `current_user.notifications_count > 0` の場合にバッジを表示
  - `viewmodel.User` に `NotificationsCount` フィールドを追加
  - `UserRepository` で通知数を取得するクエリを追加
  - サイドバーコンポーネントでバッジを条件付きレンダリング
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 80 行（実装 50 行 + テスト 30 行）

### フェーズ 5b: Stripe SDK推奨パターンへの移行

現在の実装では `stripe.Key = xxx` によるグローバル変数設定（Legacyパターン）を使用している。
stripe-go ライブラリの推奨パターン `stripe.NewClient()` に移行する。

**参考**: [stripe-go README](https://github.com/stripe/stripe-go/blob/master/README.md)

- [x] **5b-1**: Stripeクライアントの推奨パターンへの移行
  - `internal/stripe/client.go` のリファクタリング
    - `Init` 関数を削除
    - `NewClient(secretKey string) *stripe.Client` 関数を追加
    - `stripe.NewClient(secretKey)` を使用してクライアントを作成
  - `cmd/server/main.go` の更新
    - `annictStripe.NewClient(cfg.StripeSecretKey)` でクライアントを作成
    - クライアントを各ハンドラー/ユースケースに渡す
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

- [x] **5b-2**: supporters Handlerの推奨パターンへの移行
  - `internal/handler/supporters/handler.go` の更新
    - `*stripe.Client` を依存性として追加
  - `internal/handler/supporters/checkout.go` のリファクタリング
    - 変更前: `session.New(params)`
    - 変更後: `h.stripeClient.V1.CheckoutSessions.Create(ctx, params)`
  - `internal/handler/supporters/portal.go` のリファクタリング
    - 変更前: `portalsession.New(params)`
    - 変更後: `h.stripeClient.V1.BillingPortalSessions.Create(ctx, params)`
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

- [x] **5b-3**: CreateStripeSubscriber Usecaseの推奨パターンへの移行
  - `internal/usecase/create_stripe_subscriber.go` のリファクタリング
    - `*stripe.Client` を依存性として追加
    - 変更前: `subscription.Get(input.StripeSubscriptionID, nil)`
    - 変更後: `uc.stripeClient.V1Subscriptions.Retrieve(ctx, input.StripeSubscriptionID, nil)`
  - `cmd/server/main.go` の更新
    - `NewCreateStripeSubscriberUsecase` にStripeクライアントを渡す
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

### フェーズ 5c: ハンドラー命名規則の修正

Go版のハンドラーガイドライン（`go/CLAUDE.md`）に従い、標準ファイル名（8種類）以外のファイルを使用している箇所を修正する。

**参考**: [Go版ハンドラーガイド](/workspace/go/docs/handler-guide.md)

- [x] **5c-1**: supporters/checkout ハンドラーのディレクトリ分離
  - `internal/handler/supporters_checkout/` ディレクトリを作成
  - `internal/handler/supporters_checkout/handler.go` を作成（Handler構造体と依存性）
  - `internal/handler/supporters_checkout/request.go` を作成（CreateRequest DTO）
  - `internal/handler/supporters_checkout/create.go` を作成（POST /supporters/checkout）
  - 既存の `internal/handler/supporters/checkout.go` を削除
  - テストファイルも移動（`checkout_test.go` → `supporters_checkout/create_test.go`, `request_test.go`）
  - `cmd/server/main.go` のルーティングを更新
  - **実ファイル数**: 5 ファイル（実装 3 + テスト 2）

- [x] **5c-2**: supporters/portal ハンドラーのディレクトリ分離
  - `internal/handler/supporters_portal/` ディレクトリを作成
  - `internal/handler/supporters_portal/handler.go` を作成（Handler構造体と依存性）
  - `internal/handler/supporters_portal/create.go` を作成（POST /supporters/portal）
    - メソッド名を `Portal` から `Create` に変更
  - 既存の `internal/handler/supporters/portal.go` を削除
  - テストファイルも移動（`portal_test.go` → `supporters_portal/create_test.go`）
  - `cmd/server/main.go` のルーティングを更新
  - **想定ファイル数**: 約 4 ファイル（実装 2 + テスト 2）
  - **想定行数**: 約 150 行（実装 90 行 + テスト 250 行、移動のため実質変更は少ない）

- [x] **5c-3**: 設計書のコード設計セクションを現在の実装に合わせて更新
  - タスク5c-1/5c-2でハンドラー構造が変更されたため、設計書の「コード設計」セクションを更新
  - **更新箇所**:
    - 「ディレクトリ構造」: `supporters/checkout.go`, `supporters/portal.go` を削除し、`supporters_checkout/`, `supporters_portal/` ディレクトリを追加
    - 「Handler」: `supporters/handler.go` のコード例から `Checkout`, `Portal` メソッドを削除
    - `supporters_checkout/handler.go` と `supporters_portal/handler.go` のコード例を追加
  - **想定ファイル数**: 1 ファイル（設計書のみ）
  - **想定行数**: 約 30 行（ドキュメント更新のみ）

### フェーズ 5d: Webhook処理のトランザクション修正

現在の実装では、Usecaseでトランザクションを開始しているが、Repositoryに渡していないため、
実際には各DB操作が別々の接続で実行されている。また、Webhookイベント処理とサブスクリプション処理が
別トランザクションで実行されており、設計書の「同一DBトランザクション内で実行」という記載と矛盾している。

**現状の問題点**:

1. **CreateStripeSubscriberUsecase**: トランザクションを開始しているが、RepositoryがWithTxをサポートしていない
2. **DeleteStripeSubscriberUsecase**: 同上
3. **Webhookハンドラー**: UseCase実行とMarkAsProcessedが別トランザクション

- [x] **5d-1**: RepositoryにWithTxメソッドを追加
  - `internal/repository/stripe_subscriber.go` にWithTxメソッドを追加
    - `WithTx(tx *sql.Tx) *StripeSubscriberRepository` メソッドを追加
    - 内部の `queries` を `queries.WithTx(tx)` に置き換えた新しいRepositoryを返す
  - `internal/repository/stripe_webhook_event.go` にWithTxメソッドを追加
    - 同様のパターンで実装
  - `internal/repository/user.go` にWithTxメソッドを追加
    - 同様のパターンで実装
  - 既存のテストが通ることを確認
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

- [x] **5d-2**: CreateStripeSubscriberUsecaseのトランザクション修正
  - `internal/usecase/create_stripe_subscriber.go` のリファクタリング
    - 修正前: `tx` を作成するが Repository に渡していない
    - 修正後: `stripeSubscriberRepo.WithTx(tx)` と `userRepo.WithTx(tx)` を使用
  - テストの更新
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 15 行 + テスト 15 行）

- [x] **5d-3**: DeleteStripeSubscriberUsecaseのトランザクション修正
  - `internal/usecase/delete_stripe_subscriber.go` のリファクタリング
    - 修正前: `tx` を作成するが Repository に渡していない
    - 修正後: `stripeSubscriberRepo.WithTx(tx)` と `userRepo.WithTx(tx)` を使用
  - テストの更新
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 15 行 + テスト 15 行）

- [x] **5d-4**: 設計書のトランザクション処理セクションの修正
  - 設計書の「トランザクション処理」セクション（366-369行付近）を実際の設計に合わせて修正
  - **修正内容**:
    - 修正前: 「イベント保存とサブスクリプション処理は同一DBトランザクション内で実行」
    - 修正後: 「サブスクリプション処理内のDB操作（StripeSubscriber作成/更新 + User紐付け）は同一DBトランザクション内で実行。イベント保存・ステータス更新は冪等性チェックにより整合性を担保」
  - 外部API呼び出しをトランザクション外に置く理由を補足
  - **想定ファイル数**: 1 ファイル（設計書のみ）
  - **想定行数**: 約 10 行（ドキュメント更新のみ）

- [x] **5d-5**: Webhook冪等性チェックの改善
  - `internal/handler/webhooks/stripe/create.go` の冪等性チェックを改善
  - **現状の問題**:
    - イベントが「存在するか」だけをチェックしている
    - status=pending/failedのイベントも「存在する」と判定され、再処理されない
    - 処理途中でクラッシュした場合、Stripeからの再送で復旧できない
  - **修正内容**:
    - `Exists` の代わりに `GetByStripeEventID` でイベントを取得
    - statusを確認し、処理済み（processed/skipped）の場合のみスキップ
    - status=pending/failedの場合は再処理を試みる
  - テストの更新
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 60 行（実装 30 行 + テスト 30 行）

### フェーズ 6: 本番リリース準備

- [x] **6-1**: 本番Stripe環境の設定
  - **Stripeダッシュボードでの設定（Live Mode）**:
    - Live Modeに切り替え
    - Product「Annict Supporters」を作成
    - Price（月額¥290、年額¥2,900、`tax_behavior: inclusive`）を作成
    - Customer Portalの設定（フェーズ1-7と同様）
    - Webhookエンドポイントの登録（`https://annict.com/webhooks/stripe`）
  - **APIキーの取得と設定**:
    - Live Mode APIキー（`sk_live_xxx`）を取得
    - Webhook署名シークレット（`whsec_xxx`）を取得
  - **想定作業時間**: 約 30 分

- [x] **6-2**: 本番環境変数の設定
  - Dokku環境変数に以下を設定:
    - `ANNICT_STRIPE_SECRET_KEY`: Live Mode Secret Key
    - `ANNICT_STRIPE_WEBHOOK_SECRET`: Webhook署名シークレット
    - `ANNICT_STRIPE_PRICE_MONTHLY_ID`: 月額Price ID（Live Mode）
    - `ANNICT_STRIPE_PRICE_YEARLY_ID`: 年額Price ID（Live Mode）
  - 設定値の確認（typoチェック）
  - **想定作業時間**: 約 15 分

- [x] **6-3**: インフラ確認
  - **SSL証明書**:
    - Webhookエンドポイント（`/webhooks/stripe`）がHTTPSでアクセス可能であることを確認
    - 証明書の有効期限確認
  - **DNS**:
    - `annict.com` のDNSレコードが正しく設定されていることを確認
  - **ファイアウォール**:
    - StripeからのWebhook受信が許可されていることを確認
    - （通常は追加設定不要だが、念のため確認）
  - **想定作業時間**: 約 10 分

- [x] **6-4**: リリース手順書の作成
  - **リリース前チェックリスト**:
    - [ ] Stripeダッシュボード設定完了（Product/Price/Webhook/Portal）
    - [ ] 環境変数設定完了（全5項目）
    - [ ] SSL証明書有効
    - [ ] Test Modeでの動作確認完了
  - **リリース手順**:
    1. DBマイグレーション実行
    2. Go版アプリケーションデプロイ
    3. Rails版アプリケーションデプロイ（`User#supporter?`拡張）
    4. 動作確認（テストアカウントでの決済フロー）
    5. 監視開始
  - **ロールバック手順**（別タスクで詳細化）
  - **想定作業時間**: 約 30 分

- [x] **6-5**: 本番リリース前最終確認
  - **機能確認（Staging環境またはTest Mode）**:
    - [ ] サポーターページ表示（未ログイン、ログイン済み、各ユーザー状態）
    - [ ] Stripe Checkout（月額/年額）
    - [ ] Webhook受信（checkout.session.completed）
    - [ ] Customer Portal（プラン変更、キャンセル、請求履歴）
    - [ ] Gumroadサポーター表示（移行案内メッセージ）
  - **チームへの周知**:
    - リリース日時の共有
    - 問題発生時の連絡先
  - **想定作業時間**: 約 60 分

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **複数プランのサポート**: 現時点では月額/年額の2プランのみ。将来的に上位プランを追加する可能性はあるが、今回は対象外
- **GumroadからStripeへの自動データ移行**: 既存Gumroadサポーターのデータを自動的にStripeに移行する機能は実装しない。ユーザーに再登録してもらう
- **Gumroad同期タスクのGo版移行**: 既存Gumroadサポーターの期限が全て切れるまではRails版の同期タスクを維持する
- **支払い失敗時のメール通知**: Stripe側の通知機能に委ねる
- **Rails側のGumroad関連コード完全削除**: 移行期間中は既存のGumroad関連コードを維持する（同期タスク等）
- **Canary版GraphQL APIのStripe対応**: Canary版GraphQL APIは現在使用されていないため、`is_supporter`や`display_supporter_badge`フィールドのStripe対応は行わない
- **サブスクリプション状態の監視**: 異常なキャンセル率の検知などは運用開始後に実際の利用状況を見てから検討する

---

## 関連設計書

- [Gumroad関連コード削除](../2_todo/gumroad-cleanup.md) - Stripe移行完了後、全Gumroadサブスクリプションが終了した時点で実施

---

## 参考資料

- [Stripe Billing Subscriptions ドキュメント](https://docs.stripe.com/billing/subscriptions/build-subscriptions)
- [Stripe Webhooks ドキュメント](https://docs.stripe.com/webhooks)
- [Stripe Customer Portal ドキュメント](https://docs.stripe.com/customer-management/portal-deep-dive)
- [stripe-go ライブラリ](https://github.com/stripe/stripe-go)
- [Go版アーキテクチャガイド](/workspace/go/docs/architecture-guide.md)
- [Go版ハンドラーガイド](/workspace/go/docs/handler-guide.md)

---

## 既存実装の参考情報

### Go版の現状

| 項目                         | 状態                        |
| ---------------------------- | --------------------------- |
| サポーターページ             | 未実装（Rails版にプロキシ） |
| GumroadSubscriber Repository | 未実装                      |
| StripeSubscriber Repository  | 未実装                      |
| サポーター判定ロジック       | 未実装                      |
| Webhook処理                  | 未実装                      |

### Go版のリバースプロキシ設定

現在のホワイトリスト（`internal/middleware/reverse_proxy.go`）:

```go
var goHandledPaths = []string{
    "/static",
    "/health",
    "/manifest.json",
    "/sign_in/password",
    "/sign_in/code",
    "/sign_in",
    "/sign_up",
    "/password/reset",
    "/password/edit",
    "/password",
}
```

`/supporters` は含まれておらず、Rails版にプロキシされている。

### Rails版のGumroad実装構造（参考）

移行時に参考にすべき既存ファイル:

| 種別       | ファイル                                                | 説明                     |
| ---------- | ------------------------------------------------------- | ------------------------ |
| Model      | `app/models/gumroad_subscriber.rb`                      | サブスクリプションモデル |
| Model      | `app/models/gumroad_client.rb`                          | Gumroad APIクライアント  |
| Concern    | `app/models/concerns/supportable.rb`                    | 共通ロジック             |
| Form       | `app/models/forms/supporter_registration_form.rb`       | 登録フォーム             |
| Creator    | `app/models/creators/supporter_registration_creator.rb` | 登録処理                 |
| Updater    | `app/models/updaters/supporter_updater.rb`              | 更新処理                 |
| Controller | `app/controllers/callbacks_controller.rb`               | OAuthコールバック        |
| Controller | `app/controllers/supporters_controller.rb`              | サポーターページ         |
| View       | `app/views/supporters/_show_body.ja.html.erb`           | 日本語版ページ           |
| View       | `app/views/supporters/_show_body.en.html.erb`           | 英語版ページ             |
| Rake       | `lib/tasks/supporters.rake`                             | 同期タスク               |

### GumroadSubscriber.active? の実装（Rails版）

```ruby
# app/models/gumroad_subscriber.rb
def active?
  !gumroad_cancelled_at&.past? && !gumroad_ended_at&.past?
end
```

Go版でも同様のロジックを実装する:

```go
func (r *GumroadSubscriberRepository) IsActive(s *query.GumroadSubscriber) bool {
    now := time.Now()

    // キャンセル日時が過去でないこと
    if s.GumroadCancelledAt.Valid && s.GumroadCancelledAt.Time.Before(now) {
        return false
    }

    // 終了日時が過去でないこと
    if s.GumroadEndedAt.Valid && s.GumroadEndedAt.Time.Before(now) {
        return false
    }

    return true
}
```

### サポーター機能の利用箇所（Rails版）

Go版で同様の機能を実装する際の参考:

- **広告非表示**: `app/components/deprecated/adsense_component.rb`
- **視聴日時入力**: `app/components/deprecated/collapses/record_form_options_collapse_component.rb`
- **サポーターバッジ**: `app/components/deprecated/badges/supporter_badge_component.rb`
- **GraphQL API**: `app/graphql/canary/types/objects/user_type.rb`
