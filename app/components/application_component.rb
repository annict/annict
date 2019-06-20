# frozen_string_literal: true

class ApplicationComponent < ActionView::Base
  def self.compile
    @compiled ||= nil
    return if @compiled

    method = [
      "def rendered_template; @output_buffer = ActionView::OutputBuffer.new;",
      ActionView::Template::Handlers::ERB.erb_implementation.new(template, trim: true).src,
      ";",
      "end"
    ]
    class_eval method.join

    @compiled = true
  end

  def render_in(view_context, &block)
    self.class.compile
    @view_context = view_context
    @content = view_context.capture(&block) if block_given?
    rendered_template
  end

  private

  attr_reader :view_context, :content
end
