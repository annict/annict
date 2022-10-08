# frozen_string_literal: true

module Deprecated::Badges
  class WorkSeasonBadgeComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, work:, class_name: "")
      super view_context
      @work = work
      @class_name = class_name
      @season = @work.season
    end

    def render
      return "" if @season.blank?

      build_html do |h|
        h.tag :a, class: "badge rounded-pill u-bg-season-#{bg_class_name} #{@class_name}", href: view_context.seasonal_work_list_path(@season.slug) do
          h.text @season.local_name
        end
      end
    end

    private

    def bg_class_name
      return @season.name unless @season.all?

      "unknown"
    end
  end
end
