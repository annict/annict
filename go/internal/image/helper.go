// Package imageは画像URL生成機能を提供します
package image

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/annict/annict/go/internal/config"
)

// ImageDataはwork_imagesテーブルのimage_dataカラムの構造
type ImageData struct {
	Original ImageFile `json:"original"`
	Master   ImageFile `json:"master"`
}

// ImageFileは画像ファイルの情報
type ImageFile struct {
	ID      string `json:"id"`      // S3互換ストレージのオブジェクトキー (例: "workimage/2349/image/master-xxx.jpg")
	Storage string `json:"storage"` // ストレージタイプ (例: "store")
}

// Helperは画像URL生成のヘルパー構造体
type Helper struct {
	config *config.Config
}

// NewHelperは新しい画像ヘルパーを作成します
func NewHelper(cfg *config.Config) *Helper {
	return &Helper{
		config: cfg,
	}
}

// GetWorkImageURLは作品画像のURLを生成します
func (h *Helper) GetWorkImageURL(imageDataJSON string, width int, format string) string {
	// image_dataがある場合は、JSONから画像URLを取得
	if imageDataJSON != "" {
		var imageData ImageData
		if err := json.Unmarshal([]byte(imageDataJSON), &imageData); err == nil {
			// masterがあれば優先的に使用 (最適化済みのJPEG)
			var objectKey string
			if imageData.Master.ID != "" {
				objectKey = imageData.Master.ID
			} else if imageData.Original.ID != "" {
				objectKey = imageData.Original.ID
			}

			if objectKey != "" {
				// imgproxyはS3プロトコルを使用 (imgproxy設定で対応済み)
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

// WorkImageHeightは指定幅に対する3:4の作品画像表示枠の高さを返す。呼び出し側は
// width / height属性やプレースホルダーの枠に使い、すべてのサムネイル領域を同じ大きさで
// 確保する。GenerateImgproxyURLは切り抜かずに元画像をこの枠へ収める。
func WorkImageHeight(width int) int {
	return width * 4 / 3
}

// GenerateImgproxyURLは作品画像を3:4の枠へ収める署名付きimgproxy URLを生成する。
func (h *Helper) GenerateImgproxyURL(originalURL string, width int, format string) string {
	if originalURL == "" {
		return ""
	}

	// 固定3:4表示枠の高さを計算する。
	height := WorkImageHeight(width)

	// 元画像のアスペクト比を保ち切り抜きが起きないよう "fit" でリサイズする。登録される
	// 作品画像は3:4とは限らず、横長の画像は "fill" では上下が切れてしまう。imgproxyは枠より
	// 片方向が小さい画像を返すので、呼び出し側が固定の3:4の枠内で中央寄せする。Rails版でも
	// ann_work_image_urlはresizing_typeを渡さずimgproxy既定の "fit" になっており
	// (1:1のアバターにのみfill-downを指定)、それに合わせている。
	processingOptions := fmt.Sprintf("resize:fit:%d:%d:0", width, height)
	if format != "jpg" {
		processingOptions = fmt.Sprintf("%s/format:%s", processingOptions, format)
	}

	// 元URLをエンコードする。
	encodedURL := base64.RawURLEncoding.EncodeToString([]byte(originalURL))

	// imgproxyのパスを組み立てる。
	path := fmt.Sprintf("/%s/%s.%s", processingOptions, encodedURL, format)

	// パスへ署名する。
	key, _ := hex.DecodeString(h.config.ImgproxyKey)
	salt, _ := hex.DecodeString(h.config.ImgproxySalt)

	mac := hmac.New(sha256.New, key)
	mac.Write(salt)
	mac.Write([]byte(path))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	// 署名付きURLを "{endpoint}/{signature}{path}" 形式で組み立てる。
	return fmt.Sprintf("%s/%s%s", h.config.ImgproxyEndpoint, signature, path)
}

// GetSrcSetは1xと2xの画像URLセットを生成します
func (h *Helper) GetSrcSet(originalURL string, width int, format string) string {
	if originalURL == "" {
		return ""
	}

	// 1xと2xのURLを生成 (それぞれ署名付き)
	url1x := h.GenerateImgproxyURL(originalURL, width, format)
	url2x := h.GenerateImgproxyURL(originalURL, width*2, format)

	return fmt.Sprintf("%s 1x, %s 2x", url1x, url2x)
}

// ExtractImageURLはimage_dataから画像URLを取得します
func (h *Helper) ExtractImageURL(imageDataJSON string) string {
	if imageDataJSON != "" {
		var imageData ImageData
		if err := json.Unmarshal([]byte(imageDataJSON), &imageData); err == nil {
			// masterがあれば優先的に使用 (最適化済みのJPEG)
			var objectKey string
			if imageData.Master.ID != "" {
				objectKey = imageData.Master.ID
			} else if imageData.Original.ID != "" {
				objectKey = imageData.Original.ID
			}

			if objectKey != "" {
				// S3プロトコルのURL (開発/本番環境: Cloudflare R2)
				// Shrineプレフィックス付き
				return fmt.Sprintf("s3://%s/shrine/%s", h.config.S3BucketName, objectKey)
			}
		}
	}
	return ""
}

// GetAvatarImageURLはアバター画像のURLを生成します (1:1比率)
func (h *Helper) GetAvatarImageURL(imageDataJSON string, width int, format string) string {
	// image_dataから元画像URLを取得
	originalURL := h.ExtractImageURL(imageDataJSON)
	if originalURL == "" {
		return ""
	}

	return h.generateSquareImgproxyURL(originalURL, width, format)
}

// generateSquareImgproxyURLは正方形 (1:1) の画像URL生成します
func (h *Helper) generateSquareImgproxyURL(originalURL string, width int, format string) string {
	if originalURL == "" {
		return ""
	}

	// 1:1比率なので高さ＝幅
	height := width

	// Processing options (fill-downでアスペクト比を維持しつつ指定サイズに収める)
	processingOptions := fmt.Sprintf("resize:fill-down:%d:%d:0/gravity:ce", width, height)
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
	return fmt.Sprintf("%s/%s%s", h.config.ImgproxyEndpoint, signature, path)
}
