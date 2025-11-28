// Package session はセッション管理機能を提供します
package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
)

// SessionKey はRailsのセッションクッキー名
const SessionKey = "_annict_session_v201904"

// Manager はセッション管理を行う
type Manager struct {
	sessionRepo *repository.SessionRepository
	cfg         *config.Config
	CookieName  string
}

// NewManager は新しいManagerを作成
func NewManager(sessionRepo *repository.SessionRepository, cfg *config.Config) *Manager {
	return &Manager{
		sessionRepo: sessionRepo,
		cfg:         cfg,
		CookieName:  SessionKey,
	}
}

// SessionData はセッションに保存されているデータ
type SessionData struct {
	UserID    *int64         `json:"warden.user.user.key"`
	CSRFToken string         `json:"_csrf_token"`
	ExtraData map[string]any `json:"-"`
}

// GetSessionID はクッキーからセッションIDを取得
func (m *Manager) GetSessionID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionKey)
	if err != nil {
		if err == http.ErrNoCookie {
			return "", nil // クッキーがない場合はnilを返す
		}
		return "", err
	}

	// ActiveRecordストアの場合、セッションIDは直接格納されている
	sessionID := cookie.Value

	return sessionID, nil
}

// GetSession はセッションIDからセッションデータを取得
func (m *Manager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	if sessionID == "" {
		return nil, nil
	}

	// SessionRepositoryがpublic IDからprivate IDへの変換を行う
	session, err := m.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // セッションが見つからない
		}
		return nil, fmt.Errorf("セッション取得エラー: %w", err)
	}

	// JSONBデータをパース
	var rawData map[string]any
	if err := json.Unmarshal(session.Data, &rawData); err != nil {
		return nil, fmt.Errorf("セッションデータのパースエラー: %w", err)
	}

	// SessionData構造体に変換
	sessionData := &SessionData{
		ExtraData: rawData,
	}

	// warden.user.user.keyからユーザーIDを抽出
	// Railsのwardenは以下の形式でユーザー情報を保存:
	// "warden.user.user.key": [[user_id], "authenticatable_salt"]
	if wardenKey, ok := rawData["warden.user.user.key"]; ok {
		if wardenArray, ok := wardenKey.([]any); ok && len(wardenArray) >= 1 {
			if userArray, ok := wardenArray[0].([]any); ok && len(userArray) >= 1 {
				if userIDFloat, ok := userArray[0].(float64); ok {
					userID := int64(userIDFloat)
					sessionData.UserID = &userID
				}
			}
		}
	}

	// CSRFトークンを取得
	if csrfToken, ok := rawData["_csrf_token"].(string); ok {
		sessionData.CSRFToken = csrfToken
	}

	return sessionData, nil
}

// GetCurrentUser は現在のログインユーザーを取得
func (m *Manager) GetCurrentUser(ctx context.Context, r *http.Request) (*query.GetUserByIDRow, error) {
	sessionID, err := m.GetSessionID(r)
	if err != nil {
		return nil, err
	}

	if sessionID == "" {
		return nil, nil // セッションがない
	}

	sessionData, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if sessionData == nil || sessionData.UserID == nil {
		return nil, nil // ログインしていない
	}

	// ユーザー情報を取得
	user, err := m.sessionRepo.GetUserByID(ctx, *sessionData.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // ユーザーが見つからない
		}
		return nil, fmt.Errorf("ユーザー取得エラー: %w", err)
	}

	return user, nil
}

// CreateSession は新規セッションを作成してユーザーIDとCSRFトークンを保存
// Rails互換のセッションデータを作成し、Cookieを設定する
func (m *Manager) CreateSession(ctx context.Context, w http.ResponseWriter, r *http.Request, userID int64) error {
	// 1. Public IDを生成
	publicID, err := generatePublicID()
	if err != nil {
		return fmt.Errorf("public IDの生成に失敗しました: %w", err)
	}

	// 2. CSRFトークンを生成
	csrfToken, err := GenerateCSRFToken()
	if err != nil {
		return fmt.Errorf("CSRFトークンの生成に失敗しました: %w", err)
	}

	// 3. セッションデータを作成
	// Railsのwardenは以下の形式でユーザー情報を保存:
	// "warden.user.user.key": [[user_id], "authenticatable_salt"]
	sessionData := map[string]any{
		"warden.user.user.key": []any{[]any{float64(userID)}, "authenticatable_salt"},
		"_csrf_token":          csrfToken,
	}

	// 4. JSONにエンコード
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("セッションデータのJSON化に失敗しました: %w", err)
	}

	// 5. DBに保存
	_, err = m.sessionRepo.CreateSession(ctx, publicID, jsonData)
	if err != nil {
		return fmt.Errorf("セッションの作成に失敗しました: %w", err)
	}

	// 6. Cookieを設定
	m.setSessionCookie(w, r, publicID)

	return nil
}

// SetValue セッションに任意の値を保存
// 新規セッション作成時は自動的にCSRFトークンを生成・保存する
func (m *Manager) SetValue(ctx context.Context, w http.ResponseWriter, r *http.Request, key, value string) error {
	// 既存のセッションIDを取得または新規作成
	sessionID, err := m.GetSessionID(r)
	if err != nil {
		return fmt.Errorf("セッションIDの取得に失敗しました: %w", err)
	}

	var sessionData map[string]any
	var sessionExists bool

	if sessionID != "" {
		// 既存セッションを取得
		session, err := m.sessionRepo.GetSessionByID(ctx, sessionID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("セッションの取得に失敗しました: %w", err)
		}

		if err == nil {
			// 既存のセッションデータをパース（重要: 既存データを保持）
			if err := json.Unmarshal(session.Data, &sessionData); err != nil {
				return fmt.Errorf("セッションデータのパースに失敗しました: %w", err)
			}
			sessionExists = true
		} else {
			// セッションが見つからない場合は新規作成
			sessionData = make(map[string]any)
			sessionExists = false
		}
	} else {
		// 新規セッションを作成
		publicID, err := generatePublicID()
		if err != nil {
			return fmt.Errorf("public IDの生成に失敗しました: %w", err)
		}
		sessionID = publicID
		sessionData = make(map[string]any)
		sessionExists = false

		// 新規セッション作成時にCSRFトークンを自動生成
		csrfToken, err := GenerateCSRFToken()
		if err != nil {
			return fmt.Errorf("CSRFトークンの生成に失敗しました: %w", err)
		}
		sessionData["_csrf_token"] = csrfToken

		// Cookieを設定
		m.setSessionCookie(w, r, publicID)
	}

	// 値を設定（既存のデータに追加）
	sessionData[key] = value

	// JSONにエンコード
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("セッションデータのJSON化に失敗しました: %w", err)
	}

	// DBに保存（UPDATE or INSERT）
	if sessionExists {
		// UPDATE
		err = m.sessionRepo.UpdateSession(ctx, sessionID, jsonData)
	} else {
		// INSERT
		_, err = m.sessionRepo.CreateSession(ctx, sessionID, jsonData)
	}

	if err != nil {
		return fmt.Errorf("セッションの保存に失敗しました: %w", err)
	}

	return nil
}

// GetValue セッションから値を取得（削除しない）
func (m *Manager) GetValue(ctx context.Context, r *http.Request, key string) (string, error) {
	// セッションIDを取得
	sessionID, err := m.GetSessionID(r)
	if err != nil || sessionID == "" {
		return "", err
	}

	// セッションを取得
	session, err := m.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("セッションの取得に失敗しました: %w", err)
	}

	// セッションデータをパース
	var sessionData map[string]any
	if err := json.Unmarshal(session.Data, &sessionData); err != nil {
		return "", fmt.Errorf("セッションデータのパースに失敗しました: %w", err)
	}

	// 値を取得
	value, exists := sessionData[key]
	if !exists {
		return "", nil
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("値が文字列ではありません")
	}

	return valueStr, nil
}

// DeleteValue セッションから値を削除
func (m *Manager) DeleteValue(ctx context.Context, r *http.Request, key string) error {
	// セッションIDを取得
	sessionID, err := m.GetSessionID(r)
	if err != nil || sessionID == "" {
		return err
	}

	// セッションを取得
	session, err := m.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // セッションがない場合は正常終了
		}
		return fmt.Errorf("セッションの取得に失敗しました: %w", err)
	}

	// セッションデータをパース
	var sessionData map[string]any
	if err := json.Unmarshal(session.Data, &sessionData); err != nil {
		return fmt.Errorf("セッションデータのパースに失敗しました: %w", err)
	}

	// 値を削除
	delete(sessionData, key)

	// JSONにエンコード
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("セッションデータのJSON化に失敗しました: %w", err)
	}

	// DBに保存
	err = m.sessionRepo.UpdateSession(ctx, sessionID, jsonData)
	if err != nil {
		return fmt.Errorf("セッションの更新に失敗しました: %w", err)
	}

	return nil
}

// getAndDeleteSessionValue セッションから値を取得して削除（flashメッセージ用の内部メソッド）
func (m *Manager) getAndDeleteSessionValue(ctx context.Context, r *http.Request, key string) (string, error) {
	// セッションIDを取得
	sessionID, err := m.GetSessionID(r)
	if err != nil || sessionID == "" {
		return "", err
	}

	// セッションを取得
	session, err := m.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("セッションの取得に失敗しました: %w", err)
	}

	// セッションデータをパース
	var sessionData map[string]any
	if err := json.Unmarshal(session.Data, &sessionData); err != nil {
		return "", fmt.Errorf("セッションデータのパースに失敗しました: %w", err)
	}

	// 値を取得
	value, exists := sessionData[key]
	if !exists {
		return "", nil
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("値が文字列ではありません")
	}

	// 値を削除
	delete(sessionData, key)

	// JSONにエンコード
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return "", fmt.Errorf("セッションデータのJSON化に失敗しました: %w", err)
	}

	// DBに保存
	err = m.sessionRepo.UpdateSession(ctx, sessionID, jsonData)
	if err != nil {
		return "", fmt.Errorf("セッションの更新に失敗しました: %w", err)
	}

	return valueStr, nil
}

// SetSessionCookieByPublicID はセッションCookieを設定する
// ハンドラーからusecase呼び出し後にCookieを設定するために使用
func (m *Manager) SetSessionCookieByPublicID(w http.ResponseWriter, r *http.Request, publicID string) {
	m.setSessionCookie(w, r, publicID)
}

// setSessionCookie はセッションCookieを設定する内部ヘルパー
// X-Forwarded-Protoヘッダーを考慮してSecure属性を判定する
func (m *Manager) setSessionCookie(w http.ResponseWriter, r *http.Request, publicID string) {
	secure := m.cfg.SessionSecure == "true"
	// リバースプロキシ経由のHTTPS接続を検出
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		secure = true
	}

	cookie := &http.Cookie{
		Name:     SessionKey,
		Value:    publicID,
		Path:     "/",
		Domain:   m.cfg.CookieDomain,
		Secure:   secure,
		HttpOnly: m.cfg.SessionHTTPOnly == "true",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60, // 30日
	}
	http.SetCookie(w, cookie)
}

// EnsureCSRFToken はセッションが存在しない場合に新規作成し、CSRFトークンを返す
// ログインフォームなど、セッションがまだ存在しない可能性があるページで使用
func (m *Manager) EnsureCSRFToken(ctx context.Context, w http.ResponseWriter, r *http.Request) (string, error) {
	// 既存のセッションIDを取得
	sessionID, err := m.GetSessionID(r)
	if err != nil {
		return "", fmt.Errorf("セッションIDの取得に失敗しました: %w", err)
	}

	// セッションが存在する場合は既存のCSRFトークンを返す
	if sessionID != "" {
		sessionData, err := m.GetSession(ctx, sessionID)
		if err != nil {
			return "", fmt.Errorf("セッションデータの取得に失敗しました: %w", err)
		}
		if sessionData != nil && sessionData.CSRFToken != "" {
			return sessionData.CSRFToken, nil
		}
	}

	// セッションが存在しないか、CSRFトークンがない場合は新規セッションを作成
	publicID, err := generatePublicID()
	if err != nil {
		return "", fmt.Errorf("public IDの生成に失敗しました: %w", err)
	}

	// CSRFトークンを生成
	csrfToken, err := GenerateCSRFToken()
	if err != nil {
		return "", fmt.Errorf("CSRFトークンの生成に失敗しました: %w", err)
	}

	// セッションデータを作成
	sessionData := map[string]any{
		"_csrf_token": csrfToken,
	}

	// JSONにエンコード
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return "", fmt.Errorf("セッションデータのJSON化に失敗しました: %w", err)
	}

	// DBに保存
	_, err = m.sessionRepo.CreateSession(ctx, publicID, jsonData)
	if err != nil {
		return "", fmt.Errorf("セッションの作成に失敗しました: %w", err)
	}

	// Cookieを設定
	m.setSessionCookie(w, r, publicID)

	return csrfToken, nil
}

// generatePublicID ランダムなpublic IDを生成
func generatePublicID() (string, error) {
	// 32バイトのランダムデータを生成
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
