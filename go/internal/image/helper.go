// Package image は画像URL生成機能を提供します
package image

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/annict/annict/internal/config"
)

// ImageData はwork_imagesテーブルのimage_dataカラムの構造
type ImageData struct {
	Original ImageFile `json:"original"`
	Master   ImageFile `json:"master"`
}

// ImageFile は画像ファイルの情報
type ImageFile struct {
	ID      string `json:"id"`      // S3互換ストレージのオブジェクトキー (例: "workimage/2349/image/master-xxx.jpg")
	Storage string `json:"storage"` // ストレージタイプ (例: "store")
}

// Helper は画像URL生成のヘルパー構造体
type Helper struct {
	config *config.Config
}

// NewHelper は新しい画像ヘルパーを作成します
func NewHelper(cfg *config.Config) *Helper {
	return &Helper{
		config: cfg,
	}
}

// GetWorkImageURL は作品画像のURLを生成します
func (h *Helper) GetWorkImageURL(imageDataJSON string, width int, format string) string {
	// image_dataがある場合は、JSONから画像URLを取得
	if imageDataJSON != "" {
		var imageData ImageData
		if err := json.Unmarshal([]byte(imageDataJSON), &imageData); err == nil {
			// masterがあれば優先的に使用（最適化済みのJPEG）
			var objectKey string
			if imageData.Master.ID != "" {
				objectKey = imageData.Master.ID
			} else if imageData.Original.ID != "" {
				objectKey = imageData.Original.ID
			}

			if objectKey != "" {
				// imgproxyはS3プロトコルを使用（imgproxy設定で対応済み）
				// 開発/本番環境: Cloudflare R2
				// Shrineを使用しているため、shrine/プレフィックスが必要
				s3URL := fmt.Sprintf("s3://%s/shrine/%s", h.config.S3BucketName, objectKey)
				return h.GenerateImgproxyURL(s3URL, width, format)
			}
		}
	}

	// 画像がない場合は空文字列を返す
	return ""
}

// GenerateImgproxyURL はimgproxyのURLを生成します
func (h *Helper) GenerateImgproxyURL(originalURL string, width int, format string) string {
	if originalURL == "" {
		return ""
	}

	// 画像の高さを4:3の比率で計算
	height := width * 3 / 4

	// Processing options
	processingOptions := fmt.Sprintf("resize:fill:%d:%d:0/gravity:ce", width, height)
	if format != "jpg" {
		processingOptions = fmt.Sprintf("%s/format:%s", processingOptions, format)
	}

	// URLをエンコード
	encodedURL := base64.RawURLEncoding.EncodeToString([]byte(originalURL))

	// パスを構築
	path := fmt.Sprintf("/%s/%s.%s", processingOptions, encodedURL, format)

	// 署名を生成
	key, _ := hex.DecodeString(h.config.ImgproxyKey)
	salt, _ := hex.DecodeString(h.config.ImgproxySalt)

	mac := hmac.New(sha256.New, key)
	mac.Write(salt)
	mac.Write([]byte(path))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	// 署名付きURLを構築
	// フォーマット: /{signature}{path}
	return fmt.Sprintf("%s/%s%s", h.config.ImgproxyEndpoint, signature, path)
}

// GetSrcSet は1xと2xの画像URLセットを生成します
func (h *Helper) GetSrcSet(originalURL string, width int, format string) string {
	if originalURL == "" {
		return ""
	}

	// 1xと2xのURLを生成（それぞれ署名付き）
	url1x := h.GenerateImgproxyURL(originalURL, width, format)
	url2x := h.GenerateImgproxyURL(originalURL, width*2, format)

	return fmt.Sprintf("%s 1x, %s 2x", url1x, url2x)
}

// ExtractImageURL はimage_dataから画像URLを取得します
func (h *Helper) ExtractImageURL(imageDataJSON string) string {
	if imageDataJSON != "" {
		var imageData ImageData
		if err := json.Unmarshal([]byte(imageDataJSON), &imageData); err == nil {
			// masterがあれば優先的に使用（最適化済みのJPEG）
			var objectKey string
			if imageData.Master.ID != "" {
				objectKey = imageData.Master.ID
			} else if imageData.Original.ID != "" {
				objectKey = imageData.Original.ID
			}

			if objectKey != "" {
				// S3プロトコルのURL（開発/本番環境: Cloudflare R2）
				// Shrineプレフィックス付き
				return fmt.Sprintf("s3://%s/shrine/%s", h.config.S3BucketName, objectKey)
			}
		}
	}
	return ""
}
