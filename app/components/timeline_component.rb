# frozen_string_literal: true

class TimelineComponent < ApplicationComponent
  def initialize(activity_groups:, pagination:)
    @activity_groups = activity_groups
    @pagination = pagination
  end

  private

  attr_reader :activity_groups, :pagination
end
