# frozen_string_literal: true

class StatusActivityComponent < ApplicationComponent
  def initialize(activity_group_entity:, page_category:)
    @activity_group_entity = activity_group_entity
    @page_category = page_category
  end

  private

  attr_reader :activity_group_entity, :page_category

  def status_entities
    activity_group_entity.itemables
  end
end
