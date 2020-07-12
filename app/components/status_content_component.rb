# frozen_string_literal: true

class StatusContentComponent < ApplicationComponent
  def initialize(status_entity:, page_category:)
    @status_entity = status_entity
    @page_category = page_category
  end

  private

  attr_reader :page_category, :status_entity
end
