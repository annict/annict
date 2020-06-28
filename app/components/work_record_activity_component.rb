# frozen_string_literal: true

class WorkRecordActivityComponent < ApplicationComponent
  def initialize(activity_group_entity:, page_category:)
    @activity_group_entity = activity_group_entity
    @page_category = page_category
  end

  private

  attr_reader :activity_group_entity, :page_category

  def user_entity
    activity_group_entity.user
  end

  def work_record_entities
    activity_group_entity.itemables
  end
end
