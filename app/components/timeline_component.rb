# frozen_string_literal: true

class TimelineComponent < ApplicationComponent
  def initialize(activities:, pagination:)
    @activities = activities
    @pagination = pagination
  end

  private

  attr_reader :activities, :pagination
end
