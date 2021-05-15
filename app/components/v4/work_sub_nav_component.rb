# frozen_string_literal: true

module V4
  class WorkSubNavComponent < V4::ApplicationComponent
    def initialize(work_entity:, page_category:)
      @work_entity = work_entity
      @page_category = page_category.to_s
    end

    private

    attr_reader :page_category, :work_entity
  end
end
