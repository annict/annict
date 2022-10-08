# frozen_string_literal: true

module Deprecated::Badges
  class WorkMediaBadgeComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, work:)
      super view_context
      @work = work
    end

    def render
      build_html do |h|
        h.tag :span, class: "badge rounded-pill u-bg-work" do
          h.text @work.media.text
        end
      end
    end
  end
end
