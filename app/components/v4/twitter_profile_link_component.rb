# frozen_string_literal: true

module V4
  class TwitterProfileLinkComponent < V4::ApplicationComponent
    def initialize(entity:, title: nil)
      @entity = entity
      @title = title
    end

    def call
      return "-" if entity.twitter_username.blank?

      link_to link_title, entity.twitter_profile_url, target: "_blank", rel: "noopener"
    end

    private

    attr_reader :title, :entity

    def link_title
      title.presence || "@#{entity.twitter_username}"
    end
  end
end
