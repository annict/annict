# typed: false
# frozen_string_literal: true

class Deprecated::EmptyV6Component < Deprecated::ApplicationV6Component
  def initialize(view_context, text:)
    super view_context
    @text = text
  end

  def render
    build_html do |h|
      h.tag :div, class: "c-empty p-4 text-center" do
        h.tag :i, class: "fa-regular fa-face-meh display-4"

        h.tag :div, class: "h2 mt-3" do
          h.text @text
        end

        if block_given?
          yield h
        end
      end
    end
  end
end
