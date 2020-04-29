# frozen_string_literal: true

class EmptyComponent < ApplicationComponent
  def initialize(text:)
    @text = text
  end

  def call
    Htmlrb.build do |el|
      el.div class: "c-empty p-4 text-center" do
        el.i(class: "far fa-meh") {}

        el.div class: " h2 mt-3" do
          text
        end

        el.div do
          content
        end
      end
    end.html_safe
  end

  private

  attr_reader :text
end
