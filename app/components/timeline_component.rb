# frozen_string_literal: true

class TimelineComponent < ApplicationComponent
  def initialize(username:, page_category:, activity_groups:, page_info:)
    @username = username
    @page_category = page_category
    @activity_groups = activity_groups
    @page_info = page_info
  end

  private

  attr_reader :activity_groups, :page_category, :page_info, :username
end
