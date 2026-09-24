// Package turnstileはturnstile機能を提供します
package turnstile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// CloudflareのSiteverify APIエンドポイント
	siteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	// タイムアウト時間 (10秒)
	requestTimeout = 10 * time.Second
)

// VerifierはTurnstileトークンを検証するインターフェース
type Verifier interface {
	Verify(ctx context.Context, token string) (bool, error)
}

// ClientはTurnstile APIクライアント
type Client struct {
	siteKey    string
	secretKey  string
	httpClient *http.Client
}

// VerifyResponseはCloudflare Siteverify APIのレスポンス
type VerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

// verifyRequestはCloudflare Siteverify APIへのリクエストボディ
type verifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
}

// NewClientは新しいTurnstile APIクライアントを作成する
func NewClient(siteKey, secretKey string) *Client {
	return &Client{
		siteKey:   siteKey,
		secretKey: secretKey,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// VerifyはTurnstileトークンを検証する
// トークンが有効な場合はtrue、無効な場合はfalseを返す
func (c *Client) Verify(ctx context.Context, token string) (bool, error) {
	// テスト環境用: SecretKeyが空の場合は常に検証成功を返す
	if c.secretKey == "" {
		return true, nil
	}

	// トークンが空なのは、ユーザーがTurnstileを完了せずにフォームを送信した
	// ケース (ウィジェット未解決・JavaScriptのブロック・Botによる直接POSTなど)。
	// これはシステムエラーではなく想定内の検証失敗なので、(false, nil) を返して
	// 呼び出し側でwarnレベルのログに寄せる。ここでerrorを返すと、空送信のたびに
	// Sentryにエラーイベントとして送られてしまう。
	if token == "" {
		return false, nil
	}

	// リクエストボディを作成
	reqBody := verifyRequest{
		Secret:   c.secretKey,
		Response: token,
	}

	// Turnstile siteverify APIは仕様上secret_keyの送信が必須のため、
	// reqBody.Secretを含む構造体のシリアライズに対するgosec G117はfalse positiveとして抑制する。
	//nolint:gosec // G117
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("リクエストボディのJSONエンコードに失敗しました: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, siteverifyURL, bytes.NewReader(jsonBody))
	if err != nil {
		return false, fmt.Errorf("HTTPリクエストの作成に失敗しました: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("HTTPリクエストの送信に失敗しました: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// レスポンスボディを読み込む
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("レスポンスボディの読み込みに失敗しました: %w", err)
	}

	// ステータスコードが200でない場合はエラー
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("siteverify APIがエラーを返しました (ステータスコード: %d): %s", resp.StatusCode, string(body))
	}

	// JSONデコード
	var verifyResp VerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return false, fmt.Errorf("レスポンスのJSONデコードに失敗しました: %w", err)
	}

	// 検証結果を返す
	if !verifyResp.Success {
		// error-codesがある場合はログに記録できるように返す
		if len(verifyResp.ErrorCodes) > 0 {
			return false, fmt.Errorf("turnstile検証に失敗しました (エラーコード: %v)", verifyResp.ErrorCodes)
		}
		return false, fmt.Errorf("turnstile検証に失敗しました")
	}

	return true, nil
}
