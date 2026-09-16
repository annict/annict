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

	"github.com/annict/annict/go/internal/model"
)

// シーズン名のenum値 (Rails互換)
const (
	SeasonWinter = 1
	SeasonSpring = 2
	SeasonSummer = 3
	SeasonAutumn = 4
)

// WorkBuilderは作品テストデータのビルダー
type WorkBuilder struct {
	tx        *sql.Tx
	t         *testing.T
	id        model.WorkID
	title     string
	titleKana string
	titleEn   string
	// Railsのmedia enum: tv=1, ova=2, movie=3, web=4, other=0。
	media      int32
	seasonName int32 // enum値 (1:winter, 2:spring, 3:summer, 4:autumn)
	seasonYear int32
	noSeason   bool // trueの場合、season_year/season_nameをNULLにする
	noEpisodes bool // no_episodesカラムの値
	// 外部サービスのID (NULL許容)。nilの場合はカラムをNULLのままにする。
	scTid      *int32
	malAnimeID *int32
	// manualEpisodesCountは作品の予定総話数 (works.manual_episodes_count)。
	// nilの場合はカラムをNULLのままにします。
	manualEpisodesCount *int32
	watchersCount       int32
	// unpublishedAt / deletedAtはUnpublishable / SoftDeletableの状態カラムを
	// 設定します。nilの場合はカラムをNULLのまま (公開 / 未削除) にします。
	unpublishedAt *time.Time
	deletedAt     *time.Time
}

// NewWorkBuilderは新しいWorkBuilderを作成します
func NewWorkBuilder(t *testing.T, tx *sql.Tx) *WorkBuilder {
	return &WorkBuilder{
		tx:            tx,
		t:             t,
		id:            1,
		title:         "テストアニメ",
		seasonName:    SeasonSpring,
		seasonYear:    2024,
		watchersCount: 100,
	}
}

// WithIDは作品IDを設定します
func (b *WorkBuilder) WithID(id model.WorkID) *WorkBuilder {
	b.id = id
	return b
}

// WithTitleは作品タイトルを設定します
func (b *WorkBuilder) WithTitle(title string) *WorkBuilder {
	b.title = title
	return b
}

// WithTitleKanaはふりがなタイトルを設定します。
func (b *WorkBuilder) WithTitleKana(titleKana string) *WorkBuilder {
	b.titleKana = titleKana
	return b
}

// WithTitleEnは英語タイトルを設定します。
func (b *WorkBuilder) WithTitleEn(titleEn string) *WorkBuilder {
	b.titleEn = titleEn
	return b
}

// WithMediaはメディア種別を設定します (Rails enum: tv=1, ova=2, movie=3, web=4, other=0)。
func (b *WorkBuilder) WithMedia(media int32) *WorkBuilder {
	b.media = media
	return b
}

// WithScTidはしょぼいカレンダーの番組ID (works.sc_tid) を設定します。
func (b *WorkBuilder) WithScTid(scTid int32) *WorkBuilder {
	b.scTid = &scTid
	return b
}

// WithMalAnimeIDはMyAnimeListのアニメID (works.mal_anime_id) を設定します。
func (b *WorkBuilder) WithMalAnimeID(malAnimeID int32) *WorkBuilder {
	b.malAnimeID = &malAnimeID
	return b
}

// WithSeasonはシーズンを設定します
// seasonNameはSeasonWinter(1), SeasonSpring(2), SeasonSummer(3), SeasonAutumn(4) のいずれか
func (b *WorkBuilder) WithSeason(year int32, seasonName int32) *WorkBuilder {
	b.seasonName = seasonName
	b.seasonYear = year
	b.noSeason = false
	return b
}

// WithNoSeasonはシーズン情報なしに設定します
func (b *WorkBuilder) WithNoSeason() *WorkBuilder {
	b.noSeason = true
	return b
}

// WithNoEpisodesはno_episodesフラグを設定します
func (b *WorkBuilder) WithNoEpisodes(noEpisodes bool) *WorkBuilder {
	b.noEpisodes = noEpisodes
	return b
}

// WithUnpublishedAtはworks.unpublished_atを設定し、作品を非公開 (アーカイブ、
// Unpublishable) とします。既定ではNULL (公開) のままにします。
func (b *WorkBuilder) WithUnpublishedAt(unpublishedAt time.Time) *WorkBuilder {
	b.unpublishedAt = &unpublishedAt
	return b
}

// WithDeletedAtはworks.deleted_atを設定し、作品をソフトデリート
// (SoftDeletable) とします。既定ではNULL (未削除) のままにします。
func (b *WorkBuilder) WithDeletedAt(deletedAt time.Time) *WorkBuilder {
	b.deletedAt = &deletedAt
	return b
}

// WithManualEpisodesCountは作品が最終的に持つ予定の話数
// (works.manual_episodes_count) を設定します。既定ではNULL (不明) のままにします。
func (b *WorkBuilder) WithManualEpisodesCount(manualEpisodesCount int32) *WorkBuilder {
	b.manualEpisodesCount = &manualEpisodesCount
	return b
}

// WithWatchersCountはウォッチャー数を設定します
func (b *WorkBuilder) WithWatchersCount(count int32) *WorkBuilder {
	b.watchersCount = count
	return b
}

// Buildはテスト用の作品データをデータベースに作成します
func (b *WorkBuilder) Build() model.WorkID {
	b.t.Helper()

	q := `
		INSERT INTO works (
			title, title_kana, title_en, media, official_site_url,
			wikipedia_url, season_year, season_name,
			watchers_count, episodes_count, no_episodes,
			sc_tid, mal_anime_id, manual_episodes_count, unpublished_at, deleted_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18
		) RETURNING id
	`

	var seasonYear, seasonName interface{}
	if b.noSeason {
		seasonYear = nil
		seasonName = nil
	} else {
		seasonYear = b.seasonYear
		seasonName = b.seasonName
	}

	var scTid, malAnimeID, manualEpisodesCount interface{}
	if b.scTid != nil {
		scTid = *b.scTid
	}
	if b.malAnimeID != nil {
		malAnimeID = *b.malAnimeID
	}
	if b.manualEpisodesCount != nil {
		manualEpisodesCount = *b.manualEpisodesCount
	}

	var unpublishedAt, deletedAt interface{}
	if b.unpublishedAt != nil {
		unpublishedAt = *b.unpublishedAt
	}
	if b.deletedAt != nil {
		deletedAt = *b.deletedAt
	}

	var id int64
	err := b.tx.QueryRow(
		q,
		b.title,
		b.titleKana,
		b.titleEn,
		b.media,
		"",
		"",
		seasonYear,
		seasonName,
		b.watchersCount,
		12,
		b.noEpisodes,
		scTid,
		malAnimeID,
		manualEpisodesCount,
		unpublishedAt,
		deletedAt,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("作品データの作成に失敗しました: seasonYear=%v, seasonName=%v, error=%v", b.seasonYear, b.seasonName, err)
	}

	return model.WorkID(id)
}

// EpisodeBuilderはエピソードテストデータのビルダー
type EpisodeBuilder struct {
	tx            *sql.Tx
	t             *testing.T
	workID        model.WorkID
	number        string
	title         string
	unpublishedAt *time.Time
	deletedAt     *time.Time
}

// NewEpisodeBuilderは新しいEpisodeBuilderを作成します
func NewEpisodeBuilder(t *testing.T, tx *sql.Tx, workID model.WorkID) *EpisodeBuilder {
	return &EpisodeBuilder{
		tx:     tx,
		t:      t,
		workID: workID,
		number: "1",
		title:  "第1話",
	}
}

// WithNumberはエピソード番号を設定します
func (b *EpisodeBuilder) WithNumber(number string) *EpisodeBuilder {
	b.number = number
	return b
}

// WithTitleはエピソードタイトルを設定します
func (b *EpisodeBuilder) WithTitle(title string) *EpisodeBuilder {
	b.title = title
	return b
}

// WithUnpublishedAtはepisodes.unpublished_atを設定し、エピソードを非公開
// (アーカイブ、Unpublishable) とします。既定ではNULL (公開) のままにします。
func (b *EpisodeBuilder) WithUnpublishedAt(unpublishedAt time.Time) *EpisodeBuilder {
	b.unpublishedAt = &unpublishedAt
	return b
}

// WithDeletedAtはepisodes.deleted_atを設定し、エピソードをソフトデリート
// (SoftDeletable) とします。既定ではNULL (未削除) のままにします。
func (b *EpisodeBuilder) WithDeletedAt(deletedAt time.Time) *EpisodeBuilder {
	b.deletedAt = &deletedAt
	return b
}

// Buildはテスト用のエピソードデータをデータベースに作成します
func (b *EpisodeBuilder) Build() model.EpisodeID {
	b.t.Helper()

	query := `
		INSERT INTO episodes (
			work_id, number, sort_number, title,
			unpublished_at, deleted_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, $8
		) RETURNING id
	`

	var unpublishedAt, deletedAt interface{}
	if b.unpublishedAt != nil {
		unpublishedAt = *b.unpublishedAt
	}
	if b.deletedAt != nil {
		deletedAt = *b.deletedAt
	}

	var id int64
	sortNumber := 10 // デフォルトのソート番号
	err := b.tx.QueryRow(
		query,
		int64(b.workID),
		b.number,
		sortNumber,
		b.title,
		unpublishedAt,
		deletedAt,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("エピソードデータの作成に失敗しました: %v", err)
	}

	return model.EpisodeID(id)
}

// UserBuilderはユーザーテストデータのビルダー
type UserBuilder struct {
	tx                 *sql.Tx
	t                  *testing.T
	username           string
	email              string
	encryptedPassword  string
	locale             string
	role               int32
	stripeSubscriberID *model.StripeSubscriberID
}

// NewUserBuilderは新しいUserBuilderを作成します
func NewUserBuilder(t *testing.T, tx *sql.Tx) *UserBuilder {
	// ユニークなIDを生成 (テスト間の衝突を避ける)
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

// WithUsernameはユーザー名を設定します
func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.username = username
	return b
}

// WithEmailはメールアドレスを設定します
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}

// WithEncryptedPasswordはハッシュ化されたパスワードを設定します
func (b *UserBuilder) WithEncryptedPassword(password string) *UserBuilder {
	b.encryptedPassword = password
	return b
}

// WithLocaleはロケールを設定します
func (b *UserBuilder) WithLocale(locale string) *UserBuilder {
	b.locale = locale
	return b
}

// WithRoleはユーザーの権限を設定します (0: user, 1: admin, 2: editor)
func (b *UserBuilder) WithRole(role int32) *UserBuilder {
	b.role = role
	return b
}

// WithStripeSubscriberIDはStripeサブスクライバーIDを設定します
func (b *UserBuilder) WithStripeSubscriberID(id *model.StripeSubscriberID) *UserBuilder {
	b.stripeSubscriberID = id
	return b
}

// Buildはテスト用のユーザーデータをデータベースに作成します
func (b *UserBuilder) Build() model.UserID {
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
		b.role,   // role (0 = user, 1 = admin, 2 = editor)
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

	// プロフィールを作成 (CompleteSignUpUsecaseと同様)
	_, err = b.tx.Exec(`
		INSERT INTO profiles (user_id, name, description, created_at, updated_at, background_image_animated)
		VALUES ($1, $2, '', NOW(), NOW(), false)
	`, id, b.username)
	if err != nil {
		b.t.Fatalf("プロフィールデータの作成に失敗しました: %v", err)
	}

	// 設定を作成 (CompleteSignUpUsecaseと同様)
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

	// メール通知設定を作成 (CompleteSignUpUsecaseと同様)
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
		_, err = b.tx.Exec(`UPDATE users SET stripe_subscriber_id = $1 WHERE id = $2`, int64(*b.stripeSubscriberID), id)
		if err != nil {
			b.t.Fatalf("ユーザーのStripeSubscriberID更新に失敗しました: %v", err)
		}
	}

	return model.UserID(id)
}

// DeleteUserはNewUserBuilderが作ったユーザーを、builderが併せて挿入する行
// (プロフィール / 設定 / メール通知設定) と一緒に削除する。ロールバックされる
// トランザクションではなく共有プールにユーザーをコミットしたテストが、後始末に使う。これらの
// テーブルはON DELETE CASCADE無しでusersを参照しているため、ユーザー行だけを消そうとすると
// 失敗し、実行中ずっとすべての行が残ってしまう。
//
// builderの隣に置くのは両者を揃えて保つため。builderが挿入し始めたテーブルは、本関数が削除し
// 始めるべきテーブルでもある。
func DeleteUser(t *testing.T, db *sql.DB, userID model.UserID) {
	t.Helper()

	statements := []string{
		`DELETE FROM email_notifications WHERE user_id = $1`,
		`DELETE FROM settings WHERE user_id = $1`,
		`DELETE FROM profiles WHERE user_id = $1`,
		`DELETE FROM users WHERE id = $1`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement, int64(userID)); err != nil {
			t.Errorf("ユーザーの後始末に失敗 (%s): %v", statement, err)
		}
	}
}

// UserResultはテスト用のユーザー結果
type UserResult struct {
	ID                 model.UserID
	Username           string
	Email              string
	StripeSubscriberID *model.StripeSubscriberID
}

// BuildWithResultはテスト用のユーザーデータをデータベースに作成し、結果を返します
func (b *UserBuilder) BuildWithResult() UserResult {
	b.t.Helper()
	id := b.Build()

	return UserResult{
		ID:                 id,
		Username:           b.username,
		Email:              b.email,
		StripeSubscriberID: b.stripeSubscriberID,
	}
}

// WorkImageBuilderは作品画像テストデータのビルダー
type WorkImageBuilder struct {
	tx        *sql.Tx
	t         *testing.T
	workID    model.WorkID
	userID    model.UserID
	imageData string
}

// NewWorkImageBuilderは新しいWorkImageBuilderを作成します
func NewWorkImageBuilder(t *testing.T, tx *sql.Tx, workID model.WorkID) *WorkImageBuilder {
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

// WithImageDataは画像データJSONを設定します
func (b *WorkImageBuilder) WithImageData(imageData string) *WorkImageBuilder {
	b.imageData = imageData
	return b
}

// Buildはテスト用の作品画像データをデータベースに作成します
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
		int64(b.workID),
		int64(b.userID),
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

// SessionBuilderはセッションテストデータのビルダー
type SessionBuilder struct {
	tx        *sql.Tx
	t         *testing.T
	sessionID string
	userID    model.UserID
	data      string
}

// NewSessionBuilderは新しいSessionBuilderを作成します
func NewSessionBuilder(t *testing.T, tx *sql.Tx) *SessionBuilder {
	return &SessionBuilder{
		tx:        tx,
		t:         t,
		sessionID: "test_session_id",
		userID:    0,
		data:      "",
	}
}

// WithSessionIDはセッションIDを設定します
func (b *SessionBuilder) WithSessionID(sessionID string) *SessionBuilder {
	b.sessionID = sessionID
	return b
}

// WithUserIDはユーザーIDを設定します
func (b *SessionBuilder) WithUserID(userID model.UserID) *SessionBuilder {
	b.userID = userID
	// セッションデータにユーザーIDを含める (Rails/Rack互換フォーマット)
	// "warden.user.user.key": [[userID], "authenticatable_salt"]
	b.data = fmt.Sprintf(`{"warden.user.user.key": [[%d], "salt"]}`, userID)
	return b
}

// Buildはテスト用のセッションデータをデータベースに作成します
// Rails/Rackの仕様に合わせて、private ID ("2::" + SHA256(publicID)) をデータベースに保存し、
// public IDを返します
func (b *SessionBuilder) Build() string {
	b.t.Helper()

	// public IDからprivate IDを生成 (Rails/Rack互換)
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

	// public IDを返す (テストで使用するため)
	return b.sessionID
}

// generatePrivateIDはpublic IDからprivate IDを生成
// Rails/Rackの実装と互換性のある形式: "2::" + SHA256(publicID)
func (b *SessionBuilder) generatePrivateID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}

// CreateTestWorkは簡単にテスト用作品を作成するヘルパー関数
func CreateTestWork(t *testing.T, tx *sql.Tx, title string) model.WorkID {
	t.Helper()
	return NewWorkBuilder(t, tx).WithTitle(title).Build()
}

// CreateTestWorkWithImageは画像付きの作品を作成するヘルパー関数
func CreateTestWorkWithImage(t *testing.T, tx *sql.Tx, title string) (workID model.WorkID, imageID int64) {
	t.Helper()
	workID = NewWorkBuilder(t, tx).WithTitle(title).Build()
	imageID = NewWorkImageBuilder(t, tx, workID).Build()
	return workID, imageID
}

// CreateTestUserは簡単にテスト用ユーザーを作成するヘルパー関数
func CreateTestUser(t *testing.T, tx *sql.Tx, username string) model.UserID {
	t.Helper()
	return NewUserBuilder(t, tx).WithUsername(username).Build()
}

// CreateTestEpisodeは簡単にテスト用エピソードを作成するヘルパー関数
func CreateTestEpisode(t *testing.T, tx *sql.Tx, workID model.WorkID, number string) model.EpisodeID {
	t.Helper()
	return NewEpisodeBuilder(t, tx, workID).WithNumber(number).Build()
}

// StripeSubscriberBuilderはStripeサブスクライバーテストデータのビルダー
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

// NewStripeSubscriberBuilderは新しいStripeSubscriberBuilderを作成します
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

// WithStripeCustomerIDはStripe顧客IDを設定します
func (b *StripeSubscriberBuilder) WithStripeCustomerID(id string) *StripeSubscriberBuilder {
	b.stripeCustomerID = id
	return b
}

// WithStripeSubscriptionIDはStripeサブスクリプションIDを設定します
func (b *StripeSubscriberBuilder) WithStripeSubscriptionID(id string) *StripeSubscriberBuilder {
	b.stripeSubscriptionID = id
	return b
}

// WithStripePriceIDはStripe価格IDを設定します
func (b *StripeSubscriberBuilder) WithStripePriceID(id string) *StripeSubscriberBuilder {
	b.stripePriceID = id
	return b
}

// WithStripeStatusはStripeサブスクリプションステータスを設定します
func (b *StripeSubscriberBuilder) WithStripeStatus(status string) *StripeSubscriberBuilder {
	b.stripeStatus = status
	return b
}

// WithCurrentPeriodは現在の請求期間を設定します
func (b *StripeSubscriberBuilder) WithCurrentPeriod(start, end time.Time) *StripeSubscriberBuilder {
	b.stripeCurrentPeriodStart = start
	b.stripeCurrentPeriodEnd = end
	return b
}

// WithCancelAtはキャンセル予定日時を設定します
func (b *StripeSubscriberBuilder) WithCancelAt(cancelAt time.Time) *StripeSubscriberBuilder {
	b.stripeCancelAt = sql.NullTime{Time: cancelAt, Valid: true}
	return b
}

// WithCanceledAtは実際にキャンセルされた日時を設定します
func (b *StripeSubscriberBuilder) WithCanceledAt(canceledAt time.Time) *StripeSubscriberBuilder {
	b.stripeCanceledAt = sql.NullTime{Time: canceledAt, Valid: true}
	return b
}

// Buildはテスト用のStripeサブスクライバーデータをデータベースに作成し、IDを返します
func (b *StripeSubscriberBuilder) Build() model.StripeSubscriberID {
	b.t.Helper()
	result := b.BuildWithResult()
	return result.ID
}

// BuildWithResultはテスト用のStripeサブスクライバーデータをデータベースに作成し、全フィールドを返します
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
	var id int64
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
		&id,
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
	result.ID = model.StripeSubscriberID(id)

	return result
}

// StripeSubscriberResultはテスト用のStripeサブスクライバー結果
type StripeSubscriberResult struct {
	ID                       model.StripeSubscriberID
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

// CreateTestStripeSubscriberは簡単にテスト用Stripeサブスクライバーを作成するヘルパー関数
func CreateTestStripeSubscriber(t *testing.T, tx *sql.Tx) model.StripeSubscriberID {
	t.Helper()
	return NewStripeSubscriberBuilder(t, tx).Build()
}

// CreateTestStripeSubscriberWithStatusは指定ステータスでテスト用Stripeサブスクライバーを作成するヘルパー関数
func CreateTestStripeSubscriberWithStatus(t *testing.T, tx *sql.Tx, status string) model.StripeSubscriberID {
	t.Helper()
	return NewStripeSubscriberBuilder(t, tx).WithStripeStatus(status).Build()
}

// StripeWebhookEventBuilderはStripe Webhookイベントテストデータのビルダー
type StripeWebhookEventBuilder struct {
	tx              *sql.Tx
	t               *testing.T
	stripeEventID   string
	stripeEventType string
	stripePayload   string
	status          string
	receivedAt      time.Time
}

// NewStripeWebhookEventBuilderは新しいStripeWebhookEventBuilderを作成します
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

// WithStripeEventIDはStripeイベントIDを設定します
func (b *StripeWebhookEventBuilder) WithStripeEventID(id string) *StripeWebhookEventBuilder {
	b.stripeEventID = id
	return b
}

// WithStripeEventTypeはStripeイベントタイプを設定します
func (b *StripeWebhookEventBuilder) WithStripeEventType(eventType string) *StripeWebhookEventBuilder {
	b.stripeEventType = eventType
	return b
}

// WithStripePayloadはStripeペイロードを設定します
func (b *StripeWebhookEventBuilder) WithStripePayload(payload string) *StripeWebhookEventBuilder {
	b.stripePayload = payload
	return b
}

// WithStatusはステータスを設定します
func (b *StripeWebhookEventBuilder) WithStatus(status string) *StripeWebhookEventBuilder {
	b.status = status
	return b
}

// WithReceivedAtは受信日時を設定します
func (b *StripeWebhookEventBuilder) WithReceivedAt(receivedAt time.Time) *StripeWebhookEventBuilder {
	b.receivedAt = receivedAt
	return b
}

// Buildはテスト用のStripe Webhookイベントデータをデータベースに作成します
func (b *StripeWebhookEventBuilder) Build() model.StripeWebhookEventID {
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

	return model.StripeWebhookEventID(id)
}

// CreateTestStripeWebhookEventは簡単にテスト用Stripe Webhookイベントを作成するヘルパー関数
func CreateTestStripeWebhookEvent(t *testing.T, tx *sql.Tx) model.StripeWebhookEventID {
	t.Helper()
	return NewStripeWebhookEventBuilder(t, tx).Build()
}

// CreateTestStripeWebhookEventWithStatusは指定ステータスでテスト用Stripe Webhookイベントを作成するヘルパー関数
func CreateTestStripeWebhookEventWithStatus(t *testing.T, tx *sql.Tx, status string) model.StripeWebhookEventID {
	t.Helper()
	return NewStripeWebhookEventBuilder(t, tx).WithStatus(status).Build()
}

// GumroadSubscriberBuilderはGumroadサブスクライバーテストデータのビルダー
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

// NewGumroadSubscriberBuilderは新しいGumroadSubscriberBuilderを作成します
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

// WithGumroadIDはGumroad IDを設定します
func (b *GumroadSubscriberBuilder) WithGumroadID(id string) *GumroadSubscriberBuilder {
	b.gumroadID = id
	return b
}

// WithGumroadCancelledAtはキャンセル日時を設定します
func (b *GumroadSubscriberBuilder) WithGumroadCancelledAt(cancelledAt time.Time) *GumroadSubscriberBuilder {
	b.gumroadCancelledAt = sql.NullTime{Time: cancelledAt, Valid: true}
	return b
}

// WithGumroadEndedAtは終了日時を設定します
func (b *GumroadSubscriberBuilder) WithGumroadEndedAt(endedAt time.Time) *GumroadSubscriberBuilder {
	b.gumroadEndedAt = sql.NullTime{Time: endedAt, Valid: true}
	return b
}

// Buildはテスト用のGumroadサブスクライバーデータをデータベースに作成します
func (b *GumroadSubscriberBuilder) Build() model.GumroadSubscriberID {
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

	return model.GumroadSubscriberID(id)
}

// CreateTestGumroadSubscriberは簡単にテスト用Gumroadサブスクライバーを作成するヘルパー関数
func CreateTestGumroadSubscriber(t *testing.T, tx *sql.Tx) model.GumroadSubscriberID {
	t.Helper()
	return NewGumroadSubscriberBuilder(t, tx).Build()
}

// ChannelBuilderはチャンネルテストデータのビルダー
type ChannelBuilder struct {
	tx             *sql.Tx
	t              *testing.T
	name           string
	channelGroupID int64
}

// NewChannelBuilderは新しいChannelBuilderを作成します
func NewChannelBuilder(t *testing.T, tx *sql.Tx) *ChannelBuilder {
	uniqueID := uuid.New().String()[:8]
	return &ChannelBuilder{
		tx:   tx,
		t:    t,
		name: fmt.Sprintf("Channel_%s", uniqueID),
	}
}

// WithNameはチャンネル名を設定します
func (b *ChannelBuilder) WithName(name string) *ChannelBuilder {
	b.name = name
	return b
}

// WithChannelGroupIDはチャンネルグループIDを設定します
func (b *ChannelBuilder) WithChannelGroupID(channelGroupID int64) *ChannelBuilder {
	b.channelGroupID = channelGroupID
	return b
}

// Buildはテスト用のチャンネルデータをデータベースに作成します
func (b *ChannelBuilder) Build() int64 {
	b.t.Helper()

	// channel_group_idが設定されていない場合は作成
	channelGroupID := b.channelGroupID
	if channelGroupID == 0 {
		channelGroupID = NewChannelGroupBuilder(b.t, b.tx).Build()
	}

	query := `
		INSERT INTO channels (
			channel_group_id, name, aasm_state, vod, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id
	`

	var id int64
	err := b.tx.QueryRow(
		query,
		channelGroupID,
		b.name,
		"published",
		false,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("チャンネルデータの作成に失敗しました: %v", err)
	}

	return id
}

// ChannelGroupBuilderはチャンネルグループテストデータのビルダー
type ChannelGroupBuilder struct {
	tx   *sql.Tx
	t    *testing.T
	name string
}

// NewChannelGroupBuilderは新しいChannelGroupBuilderを作成します
func NewChannelGroupBuilder(t *testing.T, tx *sql.Tx) *ChannelGroupBuilder {
	uniqueID := uuid.New().String()[:8]
	return &ChannelGroupBuilder{
		tx:   tx,
		t:    t,
		name: fmt.Sprintf("ChannelGroup_%s", uniqueID),
	}
}

// WithNameはチャンネルグループ名を設定します
func (b *ChannelGroupBuilder) WithName(name string) *ChannelGroupBuilder {
	b.name = name
	return b
}

// Buildはテスト用のチャンネルグループデータをデータベースに作成します
func (b *ChannelGroupBuilder) Build() int64 {
	b.t.Helper()

	query := `
		INSERT INTO channel_groups (
			name, sort_number, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id
	`

	var id int64
	err := b.tx.QueryRow(
		query,
		b.name,
		1,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("チャンネルグループデータの作成に失敗しました: %v", err)
	}

	return id
}

// ProgramBuilderはプログラムテストデータのビルダー
type ProgramBuilder struct {
	tx        *sql.Tx
	t         *testing.T
	channelID int64
	workID    model.WorkID
}

// NewProgramBuilderは新しいProgramBuilderを作成します
func NewProgramBuilder(t *testing.T, tx *sql.Tx) *ProgramBuilder {
	return &ProgramBuilder{
		tx: tx,
		t:  t,
	}
}

// WithChannelIDはチャンネルIDを設定します
func (b *ProgramBuilder) WithChannelID(channelID int64) *ProgramBuilder {
	b.channelID = channelID
	return b
}

// WithWorkIDは作品IDを設定します
func (b *ProgramBuilder) WithWorkID(workID model.WorkID) *ProgramBuilder {
	b.workID = workID
	return b
}

// Buildはテスト用のプログラムデータをデータベースに作成します
func (b *ProgramBuilder) Build() int64 {
	b.t.Helper()

	query := `
		INSERT INTO programs (
			channel_id, work_id, aasm_state, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING id
	`

	var id int64
	err := b.tx.QueryRow(
		query,
		b.channelID,
		int64(b.workID),
		"published",
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("プログラムデータの作成に失敗しました: %v", err)
	}

	return id
}

// SlotBuilderは放送枠テストデータのビルダー
type SlotBuilder struct {
	tx        *sql.Tx
	t         *testing.T
	workID    model.WorkID
	channelID int64
	startedAt time.Time
	// episodeID / programIDはスロットが指す行 (slots.episode_id / slots.program_id)。
	// どちらもNULL許容で外部キーを持つため、nilの場合は存在しないIDを書かず
	// カラムをNULLのままにします。
	episodeID *model.EpisodeID
	programID *int64
	// numberはスロットが持つ話数 (slots.number)。nilの場合はカラムをNULLのままに
	// します (話数がまだ分からないスロットの状態)。
	number *int32
	// unpublishedAt / deletedAtはUnpublishable / SoftDeletableの状態カラムを
	// 設定します。nilの場合はカラムをNULL (公開中 / 未削除) のままにします。
	unpublishedAt *time.Time
	deletedAt     *time.Time
}

// NewSlotBuilderは新しいSlotBuilderを作成します
func NewSlotBuilder(t *testing.T, tx *sql.Tx) *SlotBuilder {
	return &SlotBuilder{
		tx:        tx,
		t:         t,
		startedAt: time.Now().Add(1 * time.Hour),
	}
}

// WithWorkIDは作品IDを設定します
func (b *SlotBuilder) WithWorkID(workID model.WorkID) *SlotBuilder {
	b.workID = workID
	return b
}

// WithEpisodeIDはエピソードIDを設定します
func (b *SlotBuilder) WithEpisodeID(episodeID model.EpisodeID) *SlotBuilder {
	b.episodeID = &episodeID
	return b
}

// WithChannelIDはチャンネルIDを設定します
func (b *SlotBuilder) WithChannelID(channelID int64) *SlotBuilder {
	b.channelID = channelID
	return b
}

// WithProgramIDはプログラムIDを設定します
func (b *SlotBuilder) WithProgramID(programID int64) *SlotBuilder {
	b.programID = &programID
	return b
}

// WithStartedAtは放送開始時刻を設定します
func (b *SlotBuilder) WithStartedAt(startedAt time.Time) *SlotBuilder {
	b.startedAt = startedAt
	return b
}

// WithNumberはslots.numberを設定します。自動生成がこのスロットを通じて到達する
// 話数を表します。既定ではNULL (話数未割り当て) のままにします。
func (b *SlotBuilder) WithNumber(number int32) *SlotBuilder {
	b.number = &number
	return b
}

// WithUnpublishedAtはslots.unpublished_atを設定し、スロットを非公開
// (Unpublishable) とします。既定ではNULL (公開) のままにします。
func (b *SlotBuilder) WithUnpublishedAt(unpublishedAt time.Time) *SlotBuilder {
	b.unpublishedAt = &unpublishedAt
	return b
}

// WithDeletedAtはslots.deleted_atを設定し、スロットをソフトデリート
// (SoftDeletable) とします。既定ではNULL (未削除) のままにします。
func (b *SlotBuilder) WithDeletedAt(deletedAt time.Time) *SlotBuilder {
	b.deletedAt = &deletedAt
	return b
}

// Buildはテスト用の放送枠データをデータベースに作成します
func (b *SlotBuilder) Build() model.SlotID {
	b.t.Helper()

	query := `
		INSERT INTO slots (
			work_id, episode_id, channel_id, program_id, started_at,
			number, unpublished_at, deleted_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10
		) RETURNING id
	`

	var episodeID, programID, number, unpublishedAt, deletedAt interface{}
	if b.episodeID != nil {
		episodeID = int64(*b.episodeID)
	}
	if b.programID != nil {
		programID = *b.programID
	}
	if b.number != nil {
		number = *b.number
	}
	if b.unpublishedAt != nil {
		unpublishedAt = *b.unpublishedAt
	}
	if b.deletedAt != nil {
		deletedAt = *b.deletedAt
	}

	var id int64
	err := b.tx.QueryRow(
		query,
		int64(b.workID),
		episodeID,
		b.channelID,
		programID,
		b.startedAt,
		number,
		unpublishedAt,
		deletedAt,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("放送枠データの作成に失敗しました: %v", err)
	}

	return model.SlotID(id)
}

// LibraryEntryBuilderはライブラリエントリテストデータのビルダー
type LibraryEntryBuilder struct {
	tx        *sql.Tx
	t         *testing.T
	userID    model.UserID
	workID    model.WorkID
	programID int64
	status    string
}

// NewLibraryEntryBuilderは新しいLibraryEntryBuilderを作成します
func NewLibraryEntryBuilder(t *testing.T, tx *sql.Tx) *LibraryEntryBuilder {
	return &LibraryEntryBuilder{
		tx:     tx,
		t:      t,
		status: "watching",
	}
}

// WithUserIDはユーザーIDを設定します
func (b *LibraryEntryBuilder) WithUserID(userID model.UserID) *LibraryEntryBuilder {
	b.userID = userID
	return b
}

// WithWorkIDは作品IDを設定します
func (b *LibraryEntryBuilder) WithWorkID(workID model.WorkID) *LibraryEntryBuilder {
	b.workID = workID
	return b
}

// WithProgramIDはプログラムIDを設定します
func (b *LibraryEntryBuilder) WithProgramID(programID int64) *LibraryEntryBuilder {
	b.programID = programID
	return b
}

// WithStatusはステータスを設定します (watching, wanna_watchなど)
func (b *LibraryEntryBuilder) WithStatus(status string) *LibraryEntryBuilder {
	b.status = status
	return b
}

// Buildはテスト用のライブラリエントリデータをデータベースに作成します
func (b *LibraryEntryBuilder) Build() int64 {
	b.t.Helper()

	// ステータスをRails enumの数値に変換 (kindカラム)
	kind := 0
	switch b.status {
	case "wanna_watch":
		kind = 1
	case "watching":
		kind = 2
	case "watched":
		kind = 3
	case "on_hold":
		kind = 4
	case "stop_watching":
		kind = 5
	}

	// まずstatusesテーブルにレコードを作成
	statusQuery := `
		INSERT INTO statuses (
			user_id, work_id, kind, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING id
	`
	var statusID int64
	err := b.tx.QueryRow(
		statusQuery,
		int64(b.userID),
		int64(b.workID),
		kind,
		time.Now(),
		time.Now(),
	).Scan(&statusID)
	if err != nil {
		b.t.Fatalf("ステータスデータの作成に失敗しました: %v", err)
	}

	// ライブラリエントリを作成 (status_idを設定)
	query := `
		INSERT INTO library_entries (
			user_id, work_id, program_id, kind, status_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id
	`

	var id int64
	err = b.tx.QueryRow(
		query,
		int64(b.userID),
		int64(b.workID),
		b.programID,
		kind,
		statusID,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		b.t.Fatalf("ライブラリエントリデータの作成に失敗しました: %v", err)
	}

	return id
}

// CastBuilderはキャストテストデータのビルダー
type CastBuilder struct {
	tx            *sql.Tx
	t             *testing.T
	workID        model.WorkID
	characterName string
	personName    string
}

// NewCastBuilderは新しいCastBuilderを作成します
func NewCastBuilder(t *testing.T, tx *sql.Tx, workID model.WorkID) *CastBuilder {
	return &CastBuilder{
		tx:            tx,
		t:             t,
		workID:        workID,
		characterName: "テストキャラクター",
		personName:    "テスト声優",
	}
}

// WithCharacterNameはキャラクター名を設定します
func (b *CastBuilder) WithCharacterName(name string) *CastBuilder {
	b.characterName = name
	return b
}

// WithPersonNameは人物名を設定します
func (b *CastBuilder) WithPersonName(name string) *CastBuilder {
	b.personName = name
	return b
}

// Buildはテスト用のキャストデータをデータベースに作成します
func (b *CastBuilder) Build() model.CastID {
	b.t.Helper()

	var characterID int64
	err := b.tx.QueryRow(`
		INSERT INTO characters (name, name_en, name_kana, series_id, created_at, updated_at)
		VALUES ($1, $2, '', NULL, NOW(), NOW())
		RETURNING id
	`, b.characterName, b.characterName).Scan(&characterID)
	if err != nil {
		b.t.Fatalf("キャラクターの作成に失敗: %v", err)
	}

	var personID int64
	err = b.tx.QueryRow(`
		INSERT INTO people (name, name_en, name_kana, created_at, updated_at)
		VALUES ($1, $2, '', NOW(), NOW())
		RETURNING id
	`, b.personName, b.personName).Scan(&personID)
	if err != nil {
		b.t.Fatalf("人物の作成に失敗: %v", err)
	}

	displayName := b.characterName + " as " + b.personName
	var castID int64
	err = b.tx.QueryRow(`
		INSERT INTO casts (work_id, character_id, person_id, name, name_en, sort_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, int64(b.workID), characterID, personID, displayName, displayName, 1).Scan(&castID)
	if err != nil {
		b.t.Fatalf("キャストの作成に失敗: %v", err)
	}

	return model.CastID(castID)
}

// StaffBuilderはスタッフテストデータのビルダー
type StaffBuilder struct {
	tx     *sql.Tx
	t      *testing.T
	workID model.WorkID
	name   string
	role   string
}

// NewStaffBuilderは新しいStaffBuilderを作成します
func NewStaffBuilder(t *testing.T, tx *sql.Tx, workID model.WorkID) *StaffBuilder {
	return &StaffBuilder{
		tx:     tx,
		t:      t,
		workID: workID,
		name:   "テストスタッフ",
		role:   "director",
	}
}

// WithNameはスタッフ名を設定します
func (b *StaffBuilder) WithName(name string) *StaffBuilder {
	b.name = name
	return b
}

// WithRoleはスタッフ役割を設定します (director, series_composition, other等)
func (b *StaffBuilder) WithRole(role string) *StaffBuilder {
	b.role = role
	return b
}

// Buildはテスト用のスタッフデータをデータベースに作成します
func (b *StaffBuilder) Build() model.StaffID {
	b.t.Helper()

	var personID int64
	err := b.tx.QueryRow(`
		INSERT INTO people (name, name_en, name_kana, created_at, updated_at)
		VALUES ($1, $2, '', NOW(), NOW())
		RETURNING id
	`, b.name, b.name).Scan(&personID)
	if err != nil {
		b.t.Fatalf("人物の作成に失敗: %v", err)
	}

	var staffID int64
	err = b.tx.QueryRow(`
		INSERT INTO staffs (work_id, name, name_en, role, role_other, role_other_en, resource_id, resource_type, sort_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '', '', $5, $6, 1, NOW(), NOW())
		RETURNING id
	`, int64(b.workID), b.name, b.name, b.role, personID, "Person").Scan(&staffID)
	if err != nil {
		b.t.Fatalf("スタッフの作成に失敗: %v", err)
	}

	return model.StaffID(staffID)
}
