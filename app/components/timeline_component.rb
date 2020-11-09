# frozen_string_literal: true

class TimelineComponent < ApplicationComponent
  def initialize(viewer:, page_category:, activity_group_entities:, page_info_entity:)
    @viewer = viewer
    @page_category = page_category
    @activity_group_entities = activity_group_entities
    @page_info_entity = page_info_entity
  end
end
