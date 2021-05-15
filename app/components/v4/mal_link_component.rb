# frozen_string_literal: true

module V4
  class MalLinkComponent < V4::ApplicationComponent
    def initialize(work_entity:, title: nil)
      @work_entity = work_entity
      @title = title
    end

    def call
      return "-" if work_entity.mal_anime_id.blank?

      link_to link_title, work_entity.mal_anime_url, target: "_blank", rel: "noopener"
    end

    private

    attr_reader :title, :work_entity

    def link_title
      title.presence || work_entity.mal_anime_id
    end
  end
end
