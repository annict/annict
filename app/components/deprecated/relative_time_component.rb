# frozen_string_literal: true

class Deprecated::RelativeTimeComponent < Deprecated::ApplicationV6Component
  def initialize(view_context, time:, class_name: "")
    super view_context
    @time = time
    @class_name = class_name
  end

  def render
    build_html do |h|
      h.tag :span,
        class: @class_name,
        data_controller: "relative-time",
        data_relative_time_time: @time
    end
  end
end
