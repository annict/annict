# frozen_string_literal: true

# TODO: RecordFooterComponent に置き換える
class EpisodeRecordFooterComponent < ApplicationComponent
  def initialize(viewer:, user_entity:, record_entity:, episode_record_entity:)
    @viewer = viewer
    @user_entity = user_entity
    @record_entity = record_entity
    @episode_record_entity = episode_record_entity
  end
end
