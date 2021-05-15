# frozen_string_literal: true

module V4
  class EpisodeRecordContentComponent < ApplicationComponent
    def initialize(viewer:, user_entity:, work_entity:, episode_entity:, record_entity:, episode_record_entity:, show_card: true)
      @viewer = viewer
      @user_entity = user_entity
      @work_entity = work_entity
      @episode_entity = episode_entity
      @record_entity = record_entity
      @episode_record_entity = episode_record_entity
      @show_card = show_card
    end
  end
end
