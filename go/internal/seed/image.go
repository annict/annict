package seed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	// WorkImageWidth は作品画像の幅（3:4の縦長）
	WorkImageWidth = 600
	// WorkImageHeight は作品画像の高さ（3:4の縦長）
	WorkImageHeight = 800

	// ProfileImageWidth はプロフィール画像の幅（1:1の正方形）
	ProfileImageWidth = 400
	// ProfileImageHeight はプロフィール画像の高さ（1:1の正方形）
	ProfileImageHeight = 400

	// ShrinePathPrefix はShrine gemの画像パスプレフィックス
	ShrinePathPrefix = "shrine/workimage"
	// ShrineProfilePathPrefix はプロフィール画像のShrine gemのパスプレフィックス
	ShrineProfilePathPrefix = "shrine/profile"
)

// RandomImage はランダムに生成された画像とそのメタデータを保持する
type RandomImage struct {
	Data     []byte // PNG画像のバイトデータ
	Filename string // ファイル名（UUID.png）
	Path     string // S3パス（shrine/workimage/UUID.png）
	Size     int    // ファイルサイズ（バイト）
	Width    int    // 画像の幅
	Height   int    // 画像の高さ
}

// ShrineImageData はShrine gemのimage_dataカラムに格納するJSON構造
// 実際のShrineではバージョン管理のため、トップレベルに "master" キーがある
type ShrineImageData struct {
	Master ShrineImageVersion `json:"master"`
}

// ShrineImageVersion はShrineの画像バージョン情報
type ShrineImageVersion struct {
	ID       string                  `json:"id"`
	Storage  string                  `json:"storage"`
	Metadata ShrineImageDataMetadata `json:"metadata"`
}

// ShrineImageDataMetadata はShrine gemのメタデータ構造
type ShrineImageDataMetadata struct {
	Filename string `json:"filename"`
	Size     int    `json:"size"`
	MimeType string `json:"mime_type"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// GenerateRandomWorkImage はランダムな単色画像を生成します（作品画像用: 600x800px）
// workID: 作品ID（Shrineのpretty_locationプラグインの仕様に合わせるため）
func GenerateRandomWorkImage(workID int64) (*RandomImage, error) {
	// Shrineのpretty_locationプラグインの仕様に合わせたパスを生成
	// 形式: shrine/workimage/{work_id}/image/master-{hash}.png
	hash := uuid.New().String()
	filename := fmt.Sprintf("master-%s.png", hash)
	path := fmt.Sprintf("%s/%d/image/%s", ShrinePathPrefix, workID, filename)

	return generateImage(WorkImageWidth, WorkImageHeight, filename, path)
}

// GenerateRandomProfileImage はランダムな単色画像を生成します（プロフィール画像用: 400x400px）
// profileID: プロフィールID（Shrineのpretty_locationプラグインの仕様に合わせるため）
func GenerateRandomProfileImage(profileID int64) (*RandomImage, error) {
	// Shrineのpretty_locationプラグインの仕様に合わせたパスを生成
	// 形式: shrine/profile/{profile_id}/image/master-{hash}.png
	hash := uuid.New().String()
	filename := fmt.Sprintf("master-%s.png", hash)
	path := fmt.Sprintf("%s/%d/image/%s", ShrineProfilePathPrefix, profileID, filename)

	return generateImage(ProfileImageWidth, ProfileImageHeight, filename, path)
}

// generateImage は指定されたサイズのランダムな単色画像を生成します（内部ヘルパー関数）
func generateImage(width, height int, filename, path string) (*RandomImage, error) {
	// 画像を作成
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// ランダムな単色を生成
	fillColor := randomColor()

	// 画像全体を単色で塗りつぶす
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fillColor)
		}
	}

	// PNG形式でエンコード（圧縮レベルをBestSpeedに設定して高速化）
	// テストデータなので画質より速度を優先
	var buf bytes.Buffer
	encoder := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}
	if err := encoder.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("png画像のエンコードに失敗: %w", err)
	}

	return &RandomImage{
		Data:     buf.Bytes(),
		Filename: filename,
		Path:     path,
		Size:     buf.Len(),
		Width:    width,
		Height:   height,
	}, nil
}

// randomColor はランダムなRGBカラーを生成します
// シード画像生成用のため、暗号学的に安全な乱数は不要
func randomColor() color.RGBA {
	return color.RGBA{
		R: uint8(rand.Intn(256)), // #nosec G404,G115
		G: uint8(rand.Intn(256)), // #nosec G404,G115
		B: uint8(rand.Intn(256)), // #nosec G404,G115
		A: 255,
	}
}

// UploadToR2 は画像をCloudflare R2にアップロードします
func UploadToR2(ctx context.Context, img *RandomImage, endpoint, accessKeyID, secretAccessKey, region, bucketName string) error {
	// S3クライアントを作成（Cloudflare R2はS3互換API）
	cfg := aws.Config{
		Region: region,
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		),
		BaseEndpoint: aws.String(endpoint),
	}
	client := s3.NewFromConfig(cfg)

	// R2にアップロード
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(img.Path),
		Body:        bytes.NewReader(img.Data),
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return fmt.Errorf("r2へのアップロードに失敗: %w", err)
	}

	return nil
}

// GenerateShrineImageData はShrine形式のimage_data JSONを生成します
func GenerateShrineImageData(img *RandomImage) (string, error) {
	// Shrineのpretty_locationプラグインの仕様では、
	// IDには "shrine/" プレフィックスを含めない（storage側で管理される）
	// 例: "workimage/40001/image/master-192a17d3-6832-490a-899f-16dfe7bab1b2.png"
	idWithoutPrefix := img.Path
	if len(img.Path) > 7 && img.Path[:7] == "shrine/" {
		idWithoutPrefix = img.Path[7:]
	}

	data := ShrineImageData{
		Master: ShrineImageVersion{
			ID:      idWithoutPrefix,
			Storage: "store",
			Metadata: ShrineImageDataMetadata{
				Filename: img.Filename,
				Size:     img.Size,
				MimeType: "image/png",
				Width:    img.Width,
				Height:   img.Height,
			},
		},
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("shrine JSONの生成に失敗: %w", err)
	}

	return string(jsonBytes), nil
}
