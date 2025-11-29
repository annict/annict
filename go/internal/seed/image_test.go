package seed

import (
	"bytes"
	"encoding/json"
	"image/png"
	"testing"
)

func TestGenerateShrineImageData(t *testing.T) {
	// テスト用の画像データを作成（Shrineのpretty_location形式）
	img := &RandomImage{
		Data:     []byte("dummy data"),
		Filename: "master-test-uuid.png",
		Path:     "shrine/workimage/789/image/master-test-uuid.png",
		Size:     12345,
		Width:    800,
		Height:   600,
	}

	jsonStr, err := GenerateShrineImageData(img)
	if err != nil {
		t.Fatalf("Shrine JSON生成に失敗: %v", err)
	}

	// JSONとしてパースできることを確認
	var data ShrineImageData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("生成されたJSONがパースできません: %v", err)
	}

	// 各フィールドの値を確認
	// IDは "shrine/" プレフィックスを削除した値（workimage/789/image/master-test-uuid.png）
	expectedID := "workimage/789/image/master-test-uuid.png"
	if data.Master.ID != expectedID {
		t.Errorf("Master.ID が不正: got %q, want %q", data.Master.ID, expectedID)
	}

	if data.Master.Storage != "store" {
		t.Errorf("Master.Storage が不正: got %q, want %q", data.Master.Storage, "store")
	}

	if data.Master.Metadata.Filename != img.Filename {
		t.Errorf("Master.Metadata.Filename が不正: got %q, want %q", data.Master.Metadata.Filename, img.Filename)
	}

	if data.Master.Metadata.Size != img.Size {
		t.Errorf("Master.Metadata.Size が不正: got %d, want %d", data.Master.Metadata.Size, img.Size)
	}

	if data.Master.Metadata.MimeType != "image/png" {
		t.Errorf("Master.Metadata.MimeType が不正: got %q, want %q", data.Master.Metadata.MimeType, "image/png")
	}

	if data.Master.Metadata.Width != img.Width {
		t.Errorf("Master.Metadata.Width が不正: got %d, want %d", data.Master.Metadata.Width, img.Width)
	}

	if data.Master.Metadata.Height != img.Height {
		t.Errorf("Master.Metadata.Height が不正: got %d, want %d", data.Master.Metadata.Height, img.Height)
	}
}

func TestGenerateRandomWorkImage(t *testing.T) {
	testWorkID := int64(123)
	img, err := GenerateRandomWorkImage(testWorkID)
	if err != nil {
		t.Fatalf("画像生成に失敗: %v", err)
	}

	// 画像データが存在することを確認
	if len(img.Data) == 0 {
		t.Error("画像データが空です")
	}

	// ファイル名が生成されていることを確認（master-{hash}.pngの形式）
	if img.Filename == "" {
		t.Error("ファイル名が生成されていません")
	}
	if len(img.Filename) < len("master-") || img.Filename[:7] != "master-" {
		t.Errorf("ファイル名の形式が不正: got %q, want master-{hash}.png", img.Filename)
	}

	// パスがShrineのpretty_location形式であることを確認（shrine/workimage/{workID}/image/master-{hash}.png）
	expectedPathPrefix := "shrine/workimage/123/image/master-"
	if len(img.Path) <= len(expectedPathPrefix) || img.Path[:len(expectedPathPrefix)] != expectedPathPrefix {
		t.Errorf("パスの形式が不正: got %q, want prefix %q", img.Path, expectedPathPrefix)
	}

	// サイズが正の値であることを確認
	if img.Size <= 0 {
		t.Errorf("サイズが不正: got %d, want > 0", img.Size)
	}

	// 画像の幅と高さを確認（600x800px）
	if img.Width != WorkImageWidth {
		t.Errorf("画像の幅が不正: got %d, want %d", img.Width, WorkImageWidth)
	}
	if img.Height != WorkImageHeight {
		t.Errorf("画像の高さが不正: got %d, want %d", img.Height, WorkImageHeight)
	}

	// PNGとしてデコードできることを確認
	_, err = png.Decode(bytes.NewReader(img.Data))
	if err != nil {
		t.Fatalf("生成された画像がPNGとしてデコードできません: %v", err)
	}
}

func TestGenerateRandomProfileImage(t *testing.T) {
	testProfileID := int64(456)
	img, err := GenerateRandomProfileImage(testProfileID)
	if err != nil {
		t.Fatalf("画像生成に失敗: %v", err)
	}

	// 画像データが存在することを確認
	if len(img.Data) == 0 {
		t.Error("画像データが空です")
	}

	// ファイル名が生成されていることを確認（master-{hash}.pngの形式）
	if img.Filename == "" {
		t.Error("ファイル名が生成されていません")
	}
	if len(img.Filename) < len("master-") || img.Filename[:7] != "master-" {
		t.Errorf("ファイル名の形式が不正: got %q, want master-{hash}.png", img.Filename)
	}

	// パスがShrineのpretty_location形式であることを確認（shrine/profile/{profileID}/image/master-{hash}.png）
	expectedPathPrefix := "shrine/profile/456/image/master-"
	if len(img.Path) <= len(expectedPathPrefix) || img.Path[:len(expectedPathPrefix)] != expectedPathPrefix {
		t.Errorf("パスの形式が不正: got %q, want prefix %q", img.Path, expectedPathPrefix)
	}

	// サイズが正の値であることを確認
	if img.Size <= 0 {
		t.Errorf("サイズが不正: got %d, want > 0", img.Size)
	}

	// 画像の幅と高さを確認（400x400px）
	if img.Width != ProfileImageWidth {
		t.Errorf("画像の幅が不正: got %d, want %d", img.Width, ProfileImageWidth)
	}
	if img.Height != ProfileImageHeight {
		t.Errorf("画像の高さが不正: got %d, want %d", img.Height, ProfileImageHeight)
	}

	// PNGとしてデコードできることを確認
	_, err = png.Decode(bytes.NewReader(img.Data))
	if err != nil {
		t.Fatalf("生成された画像がPNGとしてデコードできません: %v", err)
	}
}

func TestRandomColor(t *testing.T) {
	// ランダムなカラーを複数回生成して、アルファ値をチェック
	for i := 0; i < 10; i++ {
		c := randomColor()

		// アルファ値が255であることを確認
		if c.A != 255 {
			t.Errorf("A値が不正: got %d, want 255", c.A)
		}
	}
}
