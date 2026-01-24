package testutil

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// シーズン名のenum値（Rails互換）
const (
	SeasonWinter = 1
	SeasonSpring = 2
	SeasonSummer = 3
	SeasonAutumn = 4
)

// WorkBuilder は作品テストデータのビルダー
type WorkBuilder struct {
	tx         *sql.Tx
	t          *testing.T
	id         int64
	title      string
	seasonName int32 // enum値 (1:winter, 2:spring, 3:summer, 4:autumn)
	seasonYear int32
}

// NewWorkBuilder は新しいWorkBuilderを作成します
func NewWorkBuilder(t *testing.T, tx *sql.Tx) *WorkBuilder {
	return &WorkBuilder{
		tx:         tx,
		t:          t,
		id:         1,
		title:      "テストアニメ",
		seasonName: SeasonSpring,
		seasonYear: 2024,
	}
}

// WithID は作品IDを設定します
func (b *WorkBuilder) WithID(id int64) *WorkBuilder {
	b.id = id
	return b
}

// WithTitle は作品タイトルを設定します
func (b *WorkBuilder) WithTitle(title string) *WorkBuilder {
	b.title = title
	return b
}

// WithSeason はシーズンを設定します
// seasonNameは SeasonWinter(1), SeasonSpring(2), SeasonSummer(3), SeasonAutumn(4) のいずれか
func (b *WorkBuilder) WithSeason(year int32, seasonName int32) *WorkBuilder {
	b.seasonName = seasonName
	b.seasonYear = year
	return b
}

// Build はテスト用の作品データをデータベースに作成します
func (b *WorkBuilder) Build() int64 {
	b.t.Helper()

	query := `
		INSERT INTO works (
			title, title_kana, media, official_site_url,
			wikipedia_url, season_year, season_name,
			watchers_count, episodes_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9,
			$10, $11
		) RETURNING id
	`

	var id int64
	err := b.tx.QueryRow(
		query,
		b.title,      // $1
		"",           // $2 title_kana (NOT NULL制約あり)
		0,            // $3 media (0 = tv in Rails enum)
		"",           // $4 official_site_url
		"",           // $5 wikipedia_url
		b.seasonYear, // $6 season_year (int32)
		b.seasonName, // $7 season_name (int32, enum値)
		100,          // $8 watchers_count
		12,           // $9 episodes_count
		time.Now(),   // $10 created_at
		time.Now(),   // $11 updated_at
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("作品データの作成に失敗しました: seasonYear=%v, seasonName=%v, error=%v", b.seasonYear, b.seasonName, err)
	}

	return id
}

// EpisodeBuilder はエピソードテストデータのビルダー
type EpisodeBuilder struct {
	tx     *sql.Tx
	t      *testing.T
	workID int64
	number string
	title  string
}

// NewEpisodeBuilder は新しいEpisodeBuilderを作成します
func NewEpisodeBuilder(t *testing.T, tx *sql.Tx, workID int64) *EpisodeBuilder {
	return &EpisodeBuilder{
		tx:     tx,
		t:      t,
		workID: workID,
		number: "1",
		title:  "第1話",
	}
}

// WithNumber はエピソード番号を設定します
func (b *EpisodeBuilder) WithNumber(number string) *EpisodeBuilder {
	b.number = number
	return b
}

// WithTitle はエピソードタイトルを設定します
func (b *EpisodeBuilder) WithTitle(title string) *EpisodeBuilder {
	b.title = title
	return b
}

// Build はテスト用のエピソードデータをデータベースに作成します
func (b *EpisodeBuilder) Build() int64 {
	b.t.Helper()

	query := `
		INSERT INTO episodes (
			work_id, number, sort_number, title,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6
		) RETURNING id
	`

	var id int64
	sortNumber := 10 // デフォルトのソート番号
	err := b.tx.QueryRow(
		query,
		b.workID,
		b.number,
		sortNumber,
		b.title,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("エピソードデータの作成に失敗しました: %v", err)
	}

	return id
}

// UserBuilder はユーザーテストデータのビルダー
type UserBuilder struct {
	tx                 *sql.Tx
	t                  *testing.T
	username           string
	email              string
	encryptedPassword  string
	locale             string
	stripeSubscriberID *int64
}

// NewUserBuilder は新しいUserBuilderを作成します
func NewUserBuilder(t *testing.T, tx *sql.Tx) *UserBuilder {
	// ユニークなIDを生成（テスト間の衝突を避ける）
	uniqueID := uuid.New().String()[:8]
	uniqueUsername := fmt.Sprintf("testuser_%s", uniqueID)
	uniqueEmail := fmt.Sprintf("test_%s@example.com", uniqueID)

	return &UserBuilder{
		tx:                tx,
		t:                 t,
		username:          uniqueUsername,
		email:             uniqueEmail,
		encryptedPassword: "encrypted_test_password",
		locale:            "ja",
	}
}

// WithUsername はユーザー名を設定します
func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.username = username
	return b
}

// WithEmail はメールアドレスを設定します
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}

// WithEncryptedPassword はハッシュ化されたパスワードを設定します
func (b *UserBuilder) WithEncryptedPassword(password string) *UserBuilder {
	b.encryptedPassword = password
	return b
}

// WithLocale はロケールを設定します
func (b *UserBuilder) WithLocale(locale string) *UserBuilder {
	b.locale = locale
	return b
}

// WithStripeSubscriberID はStripeサブスクライバーIDを設定します
func (b *UserBuilder) WithStripeSubscriberID(id *int64) *UserBuilder {
	b.stripeSubscriberID = id
	return b
}

// Build はテスト用のユーザーデータをデータベースに作成します
func (b *UserBuilder) Build() int64 {
	b.t.Helper()

	query := `
		INSERT INTO users (
			username, email, role, locale,
			created_at, updated_at,
			encrypted_password, sign_in_count,
			time_zone, allowed_locales
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, $8,
			$9, $10
		) RETURNING id
	`

	var id int64
	err := b.tx.QueryRow(
		query,
		b.username,
		b.email,
		0,        // role (0 = user, 1 = admin, 2 = editor)
		b.locale, // locale
		time.Now(),
		time.Now(),
		b.encryptedPassword,            // encrypted_password
		0,                              // sign_in_count
		"Asia/Tokyo",                   // time_zone
		pq.Array([]string{"ja", "en"}), // allowed_locales (PostgreSQL配列)
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("ユーザーデータの作成に失敗しました: %v", err)
	}

	// プロフィールを作成（CompleteSignUpUsecaseと同様）
	_, err = b.tx.Exec(`
		INSERT INTO profiles (user_id, name, description, created_at, updated_at, background_image_animated)
		VALUES ($1, $2, '', NOW(), NOW(), false)
	`, id, b.username)
	if err != nil {
		b.t.Fatalf("プロフィールデータの作成に失敗しました: %v", err)
	}

	// 設定を作成（CompleteSignUpUsecaseと同様）
	_, err = b.tx.Exec(`
		INSERT INTO settings (
			user_id,
			privacy_policy_agreed,
			hide_record_body,
			slots_sort_type,
			display_option_work_list,
			display_option_user_work_list,
			records_sort_type,
			display_option_record_list,
			share_record_to_twitter,
			share_record_to_facebook,
			share_review_to_twitter,
			share_review_to_facebook,
			hide_supporter_badge,
			share_status_to_twitter,
			share_status_to_facebook,
			timeline_mode,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, id, true, true, "", "list_detailed", "grid_detailed", "created_at_desc", "all_comments",
		false, false, false, false, false, false, false, "following", time.Now(), time.Now())
	if err != nil {
		b.t.Fatalf("設定データの作成に失敗しました: %v", err)
	}

	// メール通知設定を作成（CompleteSignUpUsecaseと同様）
	unsubscriptionKey := fmt.Sprintf("%s-%s", uuid.New().String(), uuid.New().String())
	_, err = b.tx.Exec(`
		INSERT INTO email_notifications (
			user_id,
			unsubscription_key,
			event_followed_user,
			event_liked_episode_record,
			event_friends_joined,
			event_next_season_came,
			event_favorite_works_added,
			event_related_works_added,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, unsubscriptionKey, true, true, true, true, true, true, time.Now(), time.Now())
	if err != nil {
		b.t.Fatalf("メール通知設定データの作成に失敗しました: %v", err)
	}

	// StripeSubscriberIDが設定されている場合は更新
	if b.stripeSubscriberID != nil {
		_, err = b.tx.Exec(`UPDATE users SET stripe_subscriber_id = $1 WHERE id = $2`, *b.stripeSubscriberID, id)
		if err != nil {
			b.t.Fatalf("ユーザーのStripeSubscriberID更新に失敗しました: %v", err)
		}
	}

	return id
}

// UserResult はテスト用のユーザー結果
type UserResult struct {
	ID                 int64
	Username           string
	Email              string
	StripeSubscriberID sql.NullInt64
}

// BuildWithResult はテスト用のユーザーデータをデータベースに作成し、結果を返します
func (b *UserBuilder) BuildWithResult() UserResult {
	b.t.Helper()
	id := b.Build()

	var stripeSubID sql.NullInt64
	if b.stripeSubscriberID != nil {
		stripeSubID = sql.NullInt64{Int64: *b.stripeSubscriberID, Valid: true}
	}

	return UserResult{
		ID:                 id,
		Username:           b.username,
		Email:              b.email,
		StripeSubscriberID: stripeSubID,
	}
}

// WorkImageBuilder は作品画像テストデータのビルダー
type WorkImageBuilder struct {
	tx        *sql.Tx
	t         *testing.T
	workID    int64
	userID    int64
	imageData string
}

// NewWorkImageBuilder は新しいWorkImageBuilderを作成します
func NewWorkImageBuilder(t *testing.T, tx *sql.Tx, workID int64) *WorkImageBuilder {
	// テストユーザーを作成
	userID := CreateTestUser(t, tx, fmt.Sprintf("image_uploader_%d", workID))

	return &WorkImageBuilder{
		tx:     tx,
		t:      t,
		workID: workID,
		userID: userID,
		imageData: `{
			"id": "workimage/12345.jpg",
			"storage": "shrine",
			"metadata": {
				"size": 123456,
				"filename": "test.jpg",
				"mime_type": "image/jpeg"
			},
			"derivatives": {
				"recommended_url": "workimage/recommended/12345.jpg",
				"facebook_og_image_url": "workimage/facebook/12345.jpg",
				"twitter_image_url": "workimage/twitter/12345.jpg"
			}
		}`,
	}
}

// WithImageData は画像データJSONを設定します
func (b *WorkImageBuilder) WithImageData(imageData string) *WorkImageBuilder {
	b.imageData = imageData
	return b
}

// Build はテスト用の作品画像データをデータベースに作成します
func (b *WorkImageBuilder) Build() int64 {
	b.t.Helper()

	query := `
		INSERT INTO work_images (
			work_id, user_id, image_data, copyright, asin, color_rgb, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id
	`

	var id int64
	err := b.tx.QueryRow(
		query,
		b.workID,
		b.userID,
		b.imageData,
		"",            // copyright (NOT NULL制約あり)
		"",            // asin (NOT NULL制約あり)
		"255,255,255", // color_rgb (NOT NULL制約あり)
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("作品画像データの作成に失敗しました: %v", err)
	}

	return id
}

// SessionBuilder はセッションテストデータのビルダー
type SessionBuilder struct {
	tx        *sql.Tx
	t         *testing.T
	sessionID string
	userID    int64
	data      string
}

// NewSessionBuilder は新しいSessionBuilderを作成します
func NewSessionBuilder(t *testing.T, tx *sql.Tx) *SessionBuilder {
	return &SessionBuilder{
		tx:        tx,
		t:         t,
		sessionID: "test_session_id",
		userID:    0,
		data:      "",
	}
}

// WithSessionID はセッションIDを設定します
func (b *SessionBuilder) WithSessionID(sessionID string) *SessionBuilder {
	b.sessionID = sessionID
	return b
}

// WithUserID はユーザーIDを設定します
func (b *SessionBuilder) WithUserID(userID int64) *SessionBuilder {
	b.userID = userID
	// セッションデータにユーザーIDを含める（Rails/Rack互換フォーマット）
	// "warden.user.user.key": [[userID], "authenticatable_salt"]
	b.data = fmt.Sprintf(`{"warden.user.user.key": [[%d], "salt"]}`, userID)
	return b
}

// Build はテスト用のセッションデータをデータベースに作成します
// Rails/Rackの仕様に合わせて、private ID（"2::" + SHA256(publicID)）をデータベースに保存し、
// public IDを返します
func (b *SessionBuilder) Build() string {
	b.t.Helper()

	// public IDからprivate IDを生成（Rails/Rack互換）
	privateID := b.generatePrivateID(b.sessionID)

	query := `
		INSERT INTO sessions (
			session_id, data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4
		) ON CONFLICT (session_id)
		DO UPDATE SET
			data = EXCLUDED.data,
			updated_at = EXCLUDED.updated_at
		RETURNING session_id
	`

	var returnedPrivateID string
	err := b.tx.QueryRow(
		query,
		privateID, // private IDをデータベースに保存
		b.data,
		time.Now(),
		time.Now(),
	).Scan(&returnedPrivateID)

	if err != nil {
		b.t.Fatalf("セッションデータの作成に失敗しました: %v", err)
	}

	// public IDを返す（テストで使用するため）
	return b.sessionID
}

// generatePrivateID はpublic IDからprivate IDを生成
// Rails/Rackの実装と互換性のある形式: "2::" + SHA256(publicID)
func (b *SessionBuilder) generatePrivateID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}

// CreateTestWork は簡単にテスト用作品を作成するヘルパー関数
func CreateTestWork(t *testing.T, tx *sql.Tx, title string) int64 {
	t.Helper()
	return NewWorkBuilder(t, tx).WithTitle(title).Build()
}

// CreateTestWorkWithImage は画像付きの作品を作成するヘルパー関数
func CreateTestWorkWithImage(t *testing.T, tx *sql.Tx, title string) (workID int64, imageID int64) {
	t.Helper()
	workID = NewWorkBuilder(t, tx).WithTitle(title).Build()
	imageID = NewWorkImageBuilder(t, tx, workID).Build()
	return workID, imageID
}

// CreateTestUser は簡単にテスト用ユーザーを作成するヘルパー関数
func CreateTestUser(t *testing.T, tx *sql.Tx, username string) int64 {
	t.Helper()
	return NewUserBuilder(t, tx).WithUsername(username).Build()
}

// CreateTestEpisode は簡単にテスト用エピソードを作成するヘルパー関数
func CreateTestEpisode(t *testing.T, tx *sql.Tx, workID int64, number string) int64 {
	t.Helper()
	return NewEpisodeBuilder(t, tx, workID).WithNumber(number).Build()
}

// StripeSubscriberBuilder はStripeサブスクライバーテストデータのビルダー
type StripeSubscriberBuilder struct {
	tx                       *sql.Tx
	t                        *testing.T
	stripeCustomerID         string
	stripeSubscriptionID     string
	stripePriceID            string
	stripeStatus             string
	stripeCurrentPeriodStart time.Time
	stripeCurrentPeriodEnd   time.Time
	stripeCancelAt           sql.NullTime
	stripeCanceledAt         sql.NullTime
}

// NewStripeSubscriberBuilder は新しいStripeSubscriberBuilderを作成します
func NewStripeSubscriberBuilder(t *testing.T, tx *sql.Tx) *StripeSubscriberBuilder {
	uniqueID := uuid.New().String()[:8]
	now := time.Now()

	return &StripeSubscriberBuilder{
		tx:                       tx,
		t:                        t,
		stripeCustomerID:         fmt.Sprintf("cus_test_%s", uniqueID),
		stripeSubscriptionID:     fmt.Sprintf("sub_test_%s", uniqueID),
		stripePriceID:            "price_monthly_test",
		stripeStatus:             "active",
		stripeCurrentPeriodStart: now,
		stripeCurrentPeriodEnd:   now.AddDate(0, 1, 0),
		stripeCancelAt:           sql.NullTime{},
		stripeCanceledAt:         sql.NullTime{},
	}
}

// WithStripeCustomerID はStripe顧客IDを設定します
func (b *StripeSubscriberBuilder) WithStripeCustomerID(id string) *StripeSubscriberBuilder {
	b.stripeCustomerID = id
	return b
}

// WithStripeSubscriptionID はStripeサブスクリプションIDを設定します
func (b *StripeSubscriberBuilder) WithStripeSubscriptionID(id string) *StripeSubscriberBuilder {
	b.stripeSubscriptionID = id
	return b
}

// WithStripePriceID はStripe価格IDを設定します
func (b *StripeSubscriberBuilder) WithStripePriceID(id string) *StripeSubscriberBuilder {
	b.stripePriceID = id
	return b
}

// WithStripeStatus はStripeサブスクリプションステータスを設定します
func (b *StripeSubscriberBuilder) WithStripeStatus(status string) *StripeSubscriberBuilder {
	b.stripeStatus = status
	return b
}

// WithCurrentPeriod は現在の請求期間を設定します
func (b *StripeSubscriberBuilder) WithCurrentPeriod(start, end time.Time) *StripeSubscriberBuilder {
	b.stripeCurrentPeriodStart = start
	b.stripeCurrentPeriodEnd = end
	return b
}

// WithCancelAt はキャンセル予定日時を設定します
func (b *StripeSubscriberBuilder) WithCancelAt(cancelAt time.Time) *StripeSubscriberBuilder {
	b.stripeCancelAt = sql.NullTime{Time: cancelAt, Valid: true}
	return b
}

// WithCanceledAt は実際にキャンセルされた日時を設定します
func (b *StripeSubscriberBuilder) WithCanceledAt(canceledAt time.Time) *StripeSubscriberBuilder {
	b.stripeCanceledAt = sql.NullTime{Time: canceledAt, Valid: true}
	return b
}

// Build はテスト用のStripeサブスクライバーデータをデータベースに作成し、IDを返します
func (b *StripeSubscriberBuilder) Build() int64 {
	b.t.Helper()
	result := b.BuildWithResult()
	return result.ID
}

// BuildWithResult はテスト用のStripeサブスクライバーデータをデータベースに作成し、全フィールドを返します
func (b *StripeSubscriberBuilder) BuildWithResult() StripeSubscriberResult {
	b.t.Helper()

	q := `
		INSERT INTO stripe_subscribers (
			stripe_customer_id,
			stripe_subscription_id,
			stripe_price_id,
			stripe_status,
			stripe_current_period_start,
			stripe_current_period_end,
			stripe_cancel_at,
			stripe_canceled_at,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, stripe_customer_id, stripe_subscription_id, stripe_price_id, stripe_status,
		            stripe_current_period_start, stripe_current_period_end, stripe_cancel_at, stripe_canceled_at,
		            created_at, updated_at
	`

	now := time.Now()
	var result StripeSubscriberResult
	err := b.tx.QueryRow(
		q,
		b.stripeCustomerID,
		b.stripeSubscriptionID,
		b.stripePriceID,
		b.stripeStatus,
		b.stripeCurrentPeriodStart,
		b.stripeCurrentPeriodEnd,
		b.stripeCancelAt,
		b.stripeCanceledAt,
		now,
		now,
	).Scan(
		&result.ID,
		&result.StripeCustomerID,
		&result.StripeSubscriptionID,
		&result.StripePriceID,
		&result.StripeStatus,
		&result.StripeCurrentPeriodStart,
		&result.StripeCurrentPeriodEnd,
		&result.StripeCancelAt,
		&result.StripeCanceledAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		b.t.Fatalf("Stripeサブスクライバーデータの作成に失敗しました: %v", err)
	}

	return result
}

// StripeSubscriberResult はテスト用のStripeサブスクライバー結果
type StripeSubscriberResult struct {
	ID                       int64
	StripeCustomerID         string
	StripeSubscriptionID     string
	StripePriceID            string
	StripeStatus             string
	StripeCurrentPeriodStart time.Time
	StripeCurrentPeriodEnd   time.Time
	StripeCancelAt           sql.NullTime
	StripeCanceledAt         sql.NullTime
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

// CreateTestStripeSubscriber は簡単にテスト用Stripeサブスクライバーを作成するヘルパー関数
func CreateTestStripeSubscriber(t *testing.T, tx *sql.Tx) int64 {
	t.Helper()
	return NewStripeSubscriberBuilder(t, tx).Build()
}

// CreateTestStripeSubscriberWithStatus は指定ステータスでテスト用Stripeサブスクライバーを作成するヘルパー関数
func CreateTestStripeSubscriberWithStatus(t *testing.T, tx *sql.Tx, status string) int64 {
	t.Helper()
	return NewStripeSubscriberBuilder(t, tx).WithStripeStatus(status).Build()
}

// StripeWebhookEventBuilder はStripe Webhookイベントテストデータのビルダー
type StripeWebhookEventBuilder struct {
	tx              *sql.Tx
	t               *testing.T
	stripeEventID   string
	stripeEventType string
	stripePayload   string
	status          string
	receivedAt      time.Time
}

// NewStripeWebhookEventBuilder は新しいStripeWebhookEventBuilderを作成します
func NewStripeWebhookEventBuilder(t *testing.T, tx *sql.Tx) *StripeWebhookEventBuilder {
	uniqueID := uuid.New().String()[:8]

	return &StripeWebhookEventBuilder{
		tx:              tx,
		t:               t,
		stripeEventID:   fmt.Sprintf("evt_test_%s", uniqueID),
		stripeEventType: "customer.subscription.created",
		stripePayload:   `{"id": "evt_test", "type": "customer.subscription.created"}`,
		status:          "pending",
		receivedAt:      time.Now(),
	}
}

// WithStripeEventID はStripeイベントIDを設定します
func (b *StripeWebhookEventBuilder) WithStripeEventID(id string) *StripeWebhookEventBuilder {
	b.stripeEventID = id
	return b
}

// WithStripeEventType はStripeイベントタイプを設定します
func (b *StripeWebhookEventBuilder) WithStripeEventType(eventType string) *StripeWebhookEventBuilder {
	b.stripeEventType = eventType
	return b
}

// WithStripePayload はStripeペイロードを設定します
func (b *StripeWebhookEventBuilder) WithStripePayload(payload string) *StripeWebhookEventBuilder {
	b.stripePayload = payload
	return b
}

// WithStatus はステータスを設定します
func (b *StripeWebhookEventBuilder) WithStatus(status string) *StripeWebhookEventBuilder {
	b.status = status
	return b
}

// WithReceivedAt は受信日時を設定します
func (b *StripeWebhookEventBuilder) WithReceivedAt(receivedAt time.Time) *StripeWebhookEventBuilder {
	b.receivedAt = receivedAt
	return b
}

// Build はテスト用のStripe Webhookイベントデータをデータベースに作成します
func (b *StripeWebhookEventBuilder) Build() int64 {
	b.t.Helper()

	query := `
		INSERT INTO stripe_webhook_events (
			stripe_event_id,
			stripe_event_type,
			stripe_payload,
			status,
			received_at,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id
	`

	var id int64
	err := b.tx.QueryRow(
		query,
		b.stripeEventID,
		b.stripeEventType,
		b.stripePayload,
		b.status,
		b.receivedAt,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("Stripe Webhookイベントデータの作成に失敗しました: %v", err)
	}

	return id
}

// CreateTestStripeWebhookEvent は簡単にテスト用Stripe Webhookイベントを作成するヘルパー関数
func CreateTestStripeWebhookEvent(t *testing.T, tx *sql.Tx) int64 {
	t.Helper()
	return NewStripeWebhookEventBuilder(t, tx).Build()
}

// CreateTestStripeWebhookEventWithStatus は指定ステータスでテスト用Stripe Webhookイベントを作成するヘルパー関数
func CreateTestStripeWebhookEventWithStatus(t *testing.T, tx *sql.Tx, status string) int64 {
	t.Helper()
	return NewStripeWebhookEventBuilder(t, tx).WithStatus(status).Build()
}

// GumroadSubscriberBuilder はGumroadサブスクライバーテストデータのビルダー
type GumroadSubscriberBuilder struct {
	tx                 *sql.Tx
	t                  *testing.T
	gumroadID          string
	gumroadProductID   string
	gumroadProductName string
	gumroadUserID      sql.NullString
	gumroadUserEmail   sql.NullString
	gumroadCreatedAt   time.Time
	gumroadCancelledAt sql.NullTime
	gumroadEndedAt     sql.NullTime
}

// NewGumroadSubscriberBuilder は新しいGumroadSubscriberBuilderを作成します
func NewGumroadSubscriberBuilder(t *testing.T, tx *sql.Tx) *GumroadSubscriberBuilder {
	uniqueID := uuid.New().String()[:8]
	now := time.Now()

	return &GumroadSubscriberBuilder{
		tx:                 tx,
		t:                  t,
		gumroadID:          fmt.Sprintf("gum_test_%s", uniqueID),
		gumroadProductID:   "product_test_123",
		gumroadProductName: "Annict Supporters",
		gumroadUserID:      sql.NullString{String: fmt.Sprintf("user_%s", uniqueID), Valid: true},
		gumroadUserEmail:   sql.NullString{String: fmt.Sprintf("test_%s@example.com", uniqueID), Valid: true},
		gumroadCreatedAt:   now.AddDate(-1, 0, 0),
		gumroadCancelledAt: sql.NullTime{},
		gumroadEndedAt:     sql.NullTime{},
	}
}

// WithGumroadID はGumroad IDを設定します
func (b *GumroadSubscriberBuilder) WithGumroadID(id string) *GumroadSubscriberBuilder {
	b.gumroadID = id
	return b
}

// WithGumroadCancelledAt はキャンセル日時を設定します
func (b *GumroadSubscriberBuilder) WithGumroadCancelledAt(cancelledAt time.Time) *GumroadSubscriberBuilder {
	b.gumroadCancelledAt = sql.NullTime{Time: cancelledAt, Valid: true}
	return b
}

// WithGumroadEndedAt は終了日時を設定します
func (b *GumroadSubscriberBuilder) WithGumroadEndedAt(endedAt time.Time) *GumroadSubscriberBuilder {
	b.gumroadEndedAt = sql.NullTime{Time: endedAt, Valid: true}
	return b
}

// Build はテスト用のGumroadサブスクライバーデータをデータベースに作成します
func (b *GumroadSubscriberBuilder) Build() int64 {
	b.t.Helper()

	query := `
		INSERT INTO gumroad_subscribers (
			gumroad_id,
			gumroad_product_id,
			gumroad_product_name,
			gumroad_user_id,
			gumroad_user_email,
			gumroad_purchase_ids,
			gumroad_created_at,
			gumroad_cancelled_at,
			gumroad_ended_at,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id
	`

	var id int64
	err := b.tx.QueryRow(
		query,
		b.gumroadID,
		b.gumroadProductID,
		b.gumroadProductName,
		b.gumroadUserID,
		b.gumroadUserEmail,
		pq.Array([]string{}),
		b.gumroadCreatedAt,
		b.gumroadCancelledAt,
		b.gumroadEndedAt,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("Gumroadサブスクライバーデータの作成に失敗しました: %v", err)
	}

	return id
}

// CreateTestGumroadSubscriber は簡単にテスト用Gumroadサブスクライバーを作成するヘルパー関数
func CreateTestGumroadSubscriber(t *testing.T, tx *sql.Tx) int64 {
	t.Helper()
	return NewGumroadSubscriberBuilder(t, tx).Build()
}
