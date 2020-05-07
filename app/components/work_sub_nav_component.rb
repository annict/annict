# frozen_string_literal: true

class WorkSubNavComponent < ApplicationComponent
  def initialize(work:, page_category:)
    @work = work
    @page_category = page_category
  end

  private

  attr_reader :work, :page_category
end
