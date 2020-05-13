# frozen_string_literal: true

class EpisodeRecordActivityComponent < ApplicationComponent
  # @param activity [ActivityEntity]
  def initialize(activity:)
    @activity = activity
  end

  private

  attr_reader :activity
end
