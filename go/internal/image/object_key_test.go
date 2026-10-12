package image

import (
	"slices"
	"testing"
)

func TestObjectKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		imageDataJSON string
		want          []string
		wantErr       bool
	}{
		{
			name:          "正常系: originalとmasterのキーを返す",
			imageDataJSON: `{"original":{"id":"profile/1/image/original-a.png","storage":"store"},"master":{"id":"profile/1/image/master-b.jpg","storage":"store"}}`,
			want:          []string{"shrine/profile/1/image/original-a.png", "shrine/profile/1/image/master-b.jpg"},
		},
		{
			name:          "正常系: cacheの版はcacheの接頭辞を付ける",
			imageDataJSON: `{"original":{"id":"a.png","storage":"cache"}}`,
			want:          []string{"shrine/cache/a.png"},
		},
		{
			name:          "正常系: 種類の分からないストレージとIDの無い版は含めない",
			imageDataJSON: `{"original":{"id":"a.png","storage":"unknown"},"master":{"id":"","storage":"store"}}`,
			want:          nil,
		},
		{
			name:          "異常系: JSONとして解析できない",
			imageDataJSON: `{`,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ObjectKeys(tt.imageDataJSON)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ObjectKeys()のエラー = %v、エラーの期待 = %v", err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("ObjectKeys() = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}
