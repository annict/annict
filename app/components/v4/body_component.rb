# frozen_string_literal: true

module Deprecated
  class BodyComponent < Deprecated::ApplicationComponent
    def initialize(height: nil, format: :simple, class_name: "")
      @height = height
      @format = format
      @class_name = class_name
    end

    private

    attr_reader :class_name, :height

    def render_content
      case @format
      when :simple
        simple_format(content)
      when :markdown
        helpers.render_markdown(content)
      when :html
        content
      end
    end
  end
end
