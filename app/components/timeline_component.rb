# frozen_string_literal: true

class TimelineComponent < ApplicationComponent
  def initialize(activity_groups:, page_info:)
    @activity_groups = activity_groups
    @page_info = page_info
  end

  private

  attr_reader :activity_groups, :page_info
end
