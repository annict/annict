package repository_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/testutil"
)

// TestTouchSession_Success はセッションのupdated_atが更新されることをテスト
func TestTouchSession_Success(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSessionRepository(queries)

	// テスト用セッションを作成
	publicID := testutil.NewSessionBuilder(t, tx).
		WithSessionID("test_public_session_id").
		WithUserID(1).
		Build()

	// セッションの初期updated_atを取得
	privateID := generatePrivateID(publicID)
	initialSession, err := queries.GetSessionByID(context.Background(), privateID)
	if err != nil {
		t.Fatalf("初期セッションの取得に失敗: %v", err)
	}

	// 少し待ってからupdated_atを更新
	time.Sleep(100 * time.Millisecond)

	// TouchSessionを実行
	err = repo.TouchSession(context.Background(), publicID)
	if err != nil {
		t.Fatalf("TouchSessionに失敗: %v", err)
	}

	// セッションの更新後のupdated_atを取得
	updatedSession, err := queries.GetSessionByID(context.Background(), privateID)
	if err != nil {
		t.Fatalf("更新後のセッションの取得に失敗: %v", err)
	}

	// updated_atが更新されていることを確認
	if !updatedSession.UpdatedAt.After(initialSession.UpdatedAt) {
		t.Errorf("updated_atが更新されていません: initial=%v, updated=%v",
			initialSession.UpdatedAt, updatedSession.UpdatedAt)
	}
}

// TestTouchSession_NonExistentSession は存在しないセッションIDでもエラーが発生しないことをテスト
func TestTouchSession_NonExistentSession(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSessionRepository(queries)

	// 存在しないセッションIDでTouchSessionを実行
	err := repo.TouchSession(context.Background(), "non_existent_session_id")

	// エラーが発生しないことを確認（UPDATEは0行更新でもエラーにならない）
	if err != nil {
		t.Errorf("存在しないセッションIDでエラーが発生しました: %v", err)
	}
}

// TestGeneratePrivateID はprivate IDが正しいフォーマットで生成されることをテスト
func TestGeneratePrivateID(t *testing.T) {
	tests := []struct {
		name     string
		publicID string
		want     string
	}{
		{
			name:     "通常のセッションID",
			publicID: "test_session_id",
			want:     "2::" + hashString("test_session_id"),
		},
		{
			name:     "空文字列",
			publicID: "",
			want:     "2::" + hashString(""),
		},
		{
			name:     "長いセッションID",
			publicID: "very_long_session_id_with_many_characters_1234567890",
			want:     "2::" + hashString("very_long_session_id_with_many_characters_1234567890"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePrivateID(tt.publicID)
			if got != tt.want {
				t.Errorf("generatePrivateID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGeneratePrivateID_Format はprivate IDが "2::" で始まることをテスト
func TestGeneratePrivateID_Format(t *testing.T) {
	tests := []string{
		"test_id_1",
		"test_id_2",
		"another_session",
	}

	for _, publicID := range tests {
		t.Run(publicID, func(t *testing.T) {
			privateID := generatePrivateID(publicID)

			// "2::" で始まることを確認
			if len(privateID) < 3 || privateID[:3] != "2::" {
				t.Errorf("private IDが '2::' で始まっていません: %s", privateID)
			}

			// SHA256ハッシュの長さを確認（"2::" + 64文字のhex = 67文字）
			expectedLength := 3 + 64
			if len(privateID) != expectedLength {
				t.Errorf("private IDの長さが正しくありません: got %d, want %d", len(privateID), expectedLength)
			}
		})
	}
}

// hashString はテスト用のSHA256ハッシュ関数
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// generatePrivateID はテスト用のprivate ID生成関数（Repositoryの実装と同じロジック）
func generatePrivateID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("2::%s", hex.EncodeToString(hash[:]))
}

// TestGetSessionByID_Success はセッションを正常に取得できることをテスト
func TestGetSessionByID_Success(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSessionRepository(queries)

	// テスト用セッションを作成
	publicID := testutil.NewSessionBuilder(t, tx).
		WithSessionID("test_get_session_id").
		WithUserID(1).
		Build()

	// GetSessionByIDを実行
	session, err := repo.GetSessionByID(context.Background(), publicID)
	if err != nil {
		t.Fatalf("GetSessionByIDに失敗: %v", err)
	}

	// セッションが取得できたことを確認
	if session == nil {
		t.Fatal("セッションがnilです")
	}

	// private IDが正しいことを確認
	expectedPrivateID := generatePrivateID(publicID)
	if session.SessionID != expectedPrivateID {
		t.Errorf("セッションIDが一致しません: got %v, want %v", session.SessionID, expectedPrivateID)
	}
}

// TestGetSessionByID_NonExistent は存在しないセッションIDでエラーが返ることをテスト
func TestGetSessionByID_NonExistent(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSessionRepository(queries)

	// 存在しないセッションIDでGetSessionByIDを実行
	session, err := repo.GetSessionByID(context.Background(), "non_existent_session_id")

	// エラーが返ることを確認
	if err == nil {
		t.Error("存在しないセッションIDでエラーが返されませんでした")
	}

	// セッションがnilであることを確認
	if session != nil {
		t.Error("存在しないセッションIDでセッションが返されました")
	}
}

// TestGetUserByID_Success はユーザーを正常に取得できることをテスト
func TestGetUserByID_Success(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSessionRepository(queries)

	// テスト用ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_user").
		WithEmail("test@example.com").
		Build()

	// GetUserByIDを実行
	user, err := repo.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserByIDに失敗: %v", err)
	}

	// ユーザーが取得できたことを確認
	if user == nil {
		t.Fatal("ユーザーがnilです")
	}

	// ユーザーIDが一致することを確認
	if user.ID != userID {
		t.Errorf("ユーザーIDが一致しません: got %v, want %v", user.ID, userID)
	}
}

// TestGetUserByID_NonExistent は存在しないユーザーIDでエラーが返ることをテスト
func TestGetUserByID_NonExistent(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSessionRepository(queries)

	// 存在しないユーザーIDでGetUserByIDを実行
	user, err := repo.GetUserByID(context.Background(), 999999)

	// エラーが返ることを確認
	if err == nil {
		t.Error("存在しないユーザーIDでエラーが返されませんでした")
	}

	// ユーザーがnilであることを確認
	if user != nil {
		t.Error("存在しないユーザーIDでユーザーが返されました")
	}
}

// TestUpdateSession_Success はセッションを正常に更新できることをテスト
func TestUpdateSession_Success(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSessionRepository(queries)

	// テスト用セッションを作成
	publicID := testutil.NewSessionBuilder(t, tx).
		WithSessionID("test_update_session_id").
		WithUserID(1).
		Build()

	// セッションデータを更新
	newData := []byte(`{"key": "updated_value"}`)
	err := repo.UpdateSession(context.Background(), publicID, newData)
	if err != nil {
		t.Fatalf("UpdateSessionに失敗: %v", err)
	}

	// セッションが更新されたことを確認
	privateID := generatePrivateID(publicID)
	session, err := queries.GetSessionByID(context.Background(), privateID)
	if err != nil {
		t.Fatalf("更新後のセッション取得に失敗: %v", err)
	}

	// データが更新されていることを確認
	if string(session.Data) != string(newData) {
		t.Errorf("セッションデータが更新されていません: got %v, want %v", string(session.Data), string(newData))
	}
}

// TestCreateSession_Success はセッションを正常に作成できることをテスト
func TestCreateSession_Success(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSessionRepository(queries)

	// セッションを作成
	publicID := "test_create_session_id"
	sessionData := []byte(`{"key": "value"}`)

	session, err := repo.CreateSession(context.Background(), publicID, sessionData)
	if err != nil {
		t.Fatalf("CreateSessionに失敗: %v", err)
	}

	// セッションが作成されたことを確認
	privateID := generatePrivateID(publicID)
	if session.SessionID != privateID {
		t.Errorf("セッションIDが一致しません: got %v, want %v", session.SessionID, privateID)
	}

	// データが正しいことを確認
	if string(session.Data) != string(sessionData) {
		t.Errorf("セッションデータが一致しません: got %v, want %v", string(session.Data), string(sessionData))
	}

	// DBから取得して確認
	fetchedSession, err := queries.GetSessionByID(context.Background(), privateID)
	if err != nil {
		t.Fatalf("作成後のセッション取得に失敗: %v", err)
	}

	if fetchedSession.SessionID != privateID {
		t.Errorf("DBのセッションIDが一致しません: got %v, want %v", fetchedSession.SessionID, privateID)
	}
}
