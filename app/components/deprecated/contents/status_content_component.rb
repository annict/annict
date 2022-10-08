# frozen_string_literal: true

module Deprecated::Contents
  class StatusContentComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, status:)
      super view_context
      @status = status
      @work = @status.work
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-status-content" do
          h.tag :span, class: "badge rounded-pill u-bg-#{@status.kind_v3.to_s.dasherize}" do
            h.text t("enumerize.status.kind_v3.#{@status.kind_v3}")
          end

          h.tag :hr

          h.html Deprecated::Boxes::WorkBoxComponent.new(view_context, work: @work).render

          h.tag :hr

          h.tag :div, class: "mt-1" do
            h.html Deprecated::Footers::StatusFooterComponent.new(view_context, status: @status, page_category: @page_category).render
          end
        end
      end
    end
  end
end
