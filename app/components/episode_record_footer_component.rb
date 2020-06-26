# frozen_string_literal: true

class EpisodeRecordFooterComponent < ApplicationComponent
  def initialize(user_entity:, record_entity:, episode_record_entity:)
    @user_entity = user_entity
    @record_entity = record_entity
    @episode_record_entity = episode_record_entity
  end

  private

  attr_reader :episode_record_entity, :record_entity, :user_entity
end
