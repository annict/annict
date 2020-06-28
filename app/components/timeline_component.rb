# frozen_string_literal: true

class TimelineComponent < ApplicationComponent
  def initialize(username:, page_category:, activity_group_entities:, page_info_entity:)
    @username = username
    @page_category = page_category
    @activity_group_entities = activity_group_entities
    @page_info_entity = page_info_entity
  end

  private

  attr_reader :activity_group_entities, :page_category, :page_info_entity, :username
end
