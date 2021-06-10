# frozen_string_literal: true

module Contents
  class StatusContentComponent < ApplicationV6Component
    def initialize(view_context, status:)
      super view_context
      @status = status
      @anime = @status.anime
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-status-content" do
          h.tag :span, class: "badge u-bg-#{@status.kind_v3}" do
            h.text t("enumerize.status.kind_v3.#{@status.kind_v3}")
          end

          h.tag :hr

          h.html Boxes::AnimeBoxComponent.new(view_context, anime: @anime).render

          h.tag :hr

          h.tag :div, class: "mt-1" do
            h.html V6::Footers::StatusFooterComponent.new(view_context, status: @status, page_category: @page_category).render
          end
        end
      end
    end
  end
end
