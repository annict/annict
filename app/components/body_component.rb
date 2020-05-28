# frozen_string_literal: true

class BodyComponent < ApplicationComponent
  def initialize(height: nil, markdown: false, class_name: "")
    @height = height
    @markdown = markdown
    @class_name = class_name
  end

  private

  attr_reader :class_name, :height, :markdown

  def render_content
    return simple_format(content) unless markdown

    helpers.render_markdown(content)
  end
end
