# frozen_string_literal: true

class EpisodeRecordContentComponent2 < ApplicationComponent
  def initialize(current_user:, record:, show_card: true)
    @current_user = current_user
    @record = record
    @episode_record = @record.episode_record
    @show_card = show_card
  end
end
