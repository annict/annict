# frozen_string_literal: true

class StatusActivityComponent < ApplicationComponent
  include TimeZoneHelper

  # @param activity [ActivityEntity]
  def initialize(activity:)
    @activity = activity
  end

  private

  attr_reader :activity

  def statuses
    activity.resources
  end
end
