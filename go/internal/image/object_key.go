package image

import (
	"encoding/json"
	"fmt"
)

// shrineStoragePrefixesは、Rails版のShrineがストレージの種類ごとにオブジェクトキーへ付ける接頭辞。
var shrineStoragePrefixes = map[string]string{
	"store": "shrine/",
	"cache": "shrine/cache/",
}

// ObjectKeysは、Shrineの添付データ (original・masterの2つの版を持つimage_data) が指す
// オブジェクトストレージのキーを返す。種類の分からないストレージの版は含めない。
func ObjectKeys(imageDataJSON string) ([]string, error) {
	var imageData ImageData
	if err := json.Unmarshal([]byte(imageDataJSON), &imageData); err != nil {
		return nil, fmt.Errorf("画像の添付データの解析に失敗: %w", err)
	}

	var keys []string
	for _, file := range []ImageFile{imageData.Original, imageData.Master} {
		prefix, ok := shrineStoragePrefixes[file.Storage]
		if !ok || file.ID == "" {
			continue
		}
		keys = append(keys, prefix+file.ID)
	}
	return keys, nil
}
