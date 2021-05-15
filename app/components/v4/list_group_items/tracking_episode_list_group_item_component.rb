# frozen_string_literal: true

module ListGroupItems
  class TrackingEpisodeListGroupItemComponent < ApplicationComponent
    def initialize(episode:)
      @episode = episode
    end
  end
end
