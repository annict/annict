# frozen_string_literal: true

class WorkSubNavComponent < ApplicationComponent
  def initialize(work_entity:, page_category:)
    @work_entity = work_entity
    @page_category = page_category.to_s
  end

  private

  attr_reader :page_category, :work_entity
end
