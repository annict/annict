# frozen_string_literal: true

class WorkImageComponent < ApplicationComponent
  def initialize(image_url_1x:, image_url_2x:, alt: "Work Image", bg_color: "#F1F1F1")
    @image_url_1x = image_url_1x
    @image_url_2x = image_url_2x
    @alt = alt
    @bg_color = bg_color
  end

  def call
    Htmlrb.build do |el|
      el.div(
        class: "c-work-image img-fluid js-lazy",
        data_bg: image_url_1x,
        data_bg_hidpi: image_url_2x,
        style: "background-color: #{bg_color};"
      ) do
        el.img alt: alt, src: dummy_src
      end
    end.html_safe
  end

  private

  attr_reader :alt, :image_url_1x, :image_url_2x, :bg_color

  def dummy_src
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 3 4'%3E%3C/svg%3E"
  end
end
