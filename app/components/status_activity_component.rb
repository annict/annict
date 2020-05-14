# frozen_string_literal: true

class StatusActivityComponent < ApplicationComponent
  include TimeZoneHelper

  # @param activity_group [ActivityGroupEntity]
  def initialize(activity_group:)
    @activity_group = activity_group
  end

  private

  attr_reader :activity_group

  def statuses
    activity_group.resources
  end
end
