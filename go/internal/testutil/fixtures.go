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
	tx                *sql.Tx
	t                 *testing.T
	username          string
	email             string
	encryptedPassword string
	locale            string
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

	return id
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
