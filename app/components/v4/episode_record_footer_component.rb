# frozen_string_literal: true

module V4
  class EpisodeRecordFooterComponent < V4::ApplicationComponent
    def initialize(viewer:, user_entity:, record_entity:, episode_record_entity:)
      @viewer = viewer
      @user_entity = user_entity
      @record_entity = record_entity
      @episode_record_entity = episode_record_entity
    end
  end
end
