package viewmodel

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// TestPublishingStatus_MirrorsDomainConstantsは各ドメインのstatus定数が、同名の
// PublishingStatus定数へ射影されることを検証する。Presentation層は導出した状態を直接変換
// する (PublishingStatus(work.DerivedStatus()) / PublishingStatus(episode.DerivedStatus()))
// ため、この共有型から値がずれたドメイン定数は、components.StatusLabelが描画しない値に
// 黙って着地してしまう。
func TestPublishingStatus_MirrorsDomainConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  PublishingStatus
		want PublishingStatus
	}{
		{name: "作品が公開中", got: PublishingStatus(model.WorkStatusPublished), want: PublishingStatusPublished},
		{name: "作品が非公開", got: PublishingStatus(model.WorkStatusArchived), want: PublishingStatusArchived},
		{name: "作品が削除済み", got: PublishingStatus(model.WorkStatusDeleted), want: PublishingStatusDeleted},
		{name: "エピソードが公開中", got: PublishingStatus(model.EpisodeStatusPublished), want: PublishingStatusPublished},
		{name: "エピソードが非公開", got: PublishingStatus(model.EpisodeStatusArchived), want: PublishingStatusArchived},
		{name: "エピソードが削除済み", got: PublishingStatus(model.EpisodeStatusDeleted), want: PublishingStatusDeleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.got != tt.want {
				t.Errorf("PublishingStatus = %q、期待値 = %q", tt.got, tt.want)
			}
		})
	}
}
