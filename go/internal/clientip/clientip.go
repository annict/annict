// Package clientip はクライアントIPアドレスの取得機能を提供します
package clientip

import (
	"net/http"
	"strings"
)

// GetClientIP はHTTPリクエストからクライアントのIPアドレスを取得します
//
// 優先順位:
// 1. CF-Connecting-IP (Cloudflareが設定する実際のクライアントIP)
// 2. X-Forwarded-For (プロキシチェーンの最初のIP)
// 3. X-Real-IP
// 4. RemoteAddr (直接接続の場合)
func GetClientIP(r *http.Request) string {
	// 1. CF-Connecting-IP（Cloudflareが設定する実際のクライアントIP）
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// 2. X-Forwarded-Forの最初のIP（プロキシチェーンの最初）
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// カンマ区切りの最初のIPを取得
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// 3. X-Real-IPヘッダーをチェック
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 4. RemoteAddrを使用（直接接続の場合、ポート番号を除去）
	clientIP := r.RemoteAddr
	if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
		clientIP = clientIP[:idx]
	}
	return clientIP
}
