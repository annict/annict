# frozen_string_literal: true

class EpisodeRecordContentComponent < ApplicationComponent
  def initialize(user_entity:, work_entity:, episode_entity:, record_entity:, episode_record_entity:, show_card: true)
    @user_entity = user_entity
    @work_entity = work_entity
    @episode_entity = episode_entity
    @record_entity = record_entity
    @episode_record_entity = episode_record_entity
    @show_card = show_card
  end

  private

  attr_reader :episode_entity, :episode_record_entity, :record_entity, :user_entity, :work_entity
end
