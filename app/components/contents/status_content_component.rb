# frozen_string_literal: true

module Contents
  class StatusContentComponent < ApplicationComponent
    def initialize(view_context, status:, page_category: "")
      super view_context
      @status = status
      @page_category = page_category
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-status-content" do
          h.html Cards::StatusCardComponent.new(view_context, status: @status).render

          h.tag :div, class: "mt-2" do
            h.html Footers::StatusFooterComponent.new(view_context, status: @status, page_category: @page_category).render
          end
        end
      end
    end
  end
end
