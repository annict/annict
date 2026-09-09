package viewmodel

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// TestPublishingStatus_MirrorsDomainConstants verifies that every domain status constant
// projects onto the PublishingStatus constant with the matching name. The Presentation layer
// converts a derived status directly (PublishingStatus(work.DerivedStatus()),
// PublishingStatus(episode.DerivedStatus())), so a domain constant whose value drifted from
// this shared type would silently land on a value components.StatusLabel does not render.
//
// [Ja] TestPublishingStatus_MirrorsDomainConstants は各ドメインの status 定数が、同名の
// PublishingStatus 定数へ射影されることを検証する。Presentation 層は導出した状態を直接変換
// する (PublishingStatus(work.DerivedStatus()) / PublishingStatus(episode.DerivedStatus()))
// ため、この共有型から値がずれたドメイン定数は、components.StatusLabel が描画しない値に
// 黙って着地してしまう。
func TestPublishingStatus_MirrorsDomainConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  PublishingStatus
		want PublishingStatus
	}{
		{name: "work published", got: PublishingStatus(model.WorkStatusPublished), want: PublishingStatusPublished},
		{name: "work archived", got: PublishingStatus(model.WorkStatusArchived), want: PublishingStatusArchived},
		{name: "work deleted", got: PublishingStatus(model.WorkStatusDeleted), want: PublishingStatusDeleted},
		{name: "episode published", got: PublishingStatus(model.EpisodeStatusPublished), want: PublishingStatusPublished},
		{name: "episode archived", got: PublishingStatus(model.EpisodeStatusArchived), want: PublishingStatusArchived},
		{name: "episode deleted", got: PublishingStatus(model.EpisodeStatusDeleted), want: PublishingStatusDeleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.got != tt.want {
				t.Errorf("PublishingStatus = %q, want %q", tt.got, tt.want)
			}
		})
	}
}
