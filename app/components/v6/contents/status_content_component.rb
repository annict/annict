# frozen_string_literal: true

module Contents
  class StatusContentComponent < ApplicationComponent
    def initialize(view_context, status:, page_category:)
      super view_context
      @status = status
      @page_category = page_category
      @anime = @status.anime
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-status-content" do
          h.tag :span, class: "badge bg-#{@status.kind_v3}" do
            h.text t("enumerize.status.kind_v3.#{@status.kind_v3}")
          end

          h.tag :hr

          h.html Boxes::AnimeBoxComponent.new(view_context,
            anime: @anime,
            page_category: @page_category).render

          h.tag :hr

          h.tag :div, class: "mt-1" do
            h.html Footers::StatusFooterComponent.new(view_context, status: @status, page_category: @page_category).render
          end
        end
      end
    end
  end
end
