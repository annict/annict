# frozen_string_literal: true

class RecordActivityComponent < ApplicationComponent
  def initialize(viewer:, activity_group_entity:, page_category:)
    @viewer = viewer
    @activity_group_entity = activity_group_entity
    @page_category = page_category
  end

  private

  def user_entity
    @activity_group_entity.user
  end

  def record_entities
    @activity_group_entity.itemables
  end
end
