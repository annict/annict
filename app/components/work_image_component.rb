# frozen_string_literal: true

class WorkImageComponent < ApplicationComponent
  def initialize(image_url_1x:, image_url_2x:, bg_color: "#FFE1F4")
    @image_url_1x = image_url_1x
    @image_url_2x = image_url_2x
    @bg_color = bg_color
  end

  def call
    Htmlrb.build do |el|
      el.div(
        class: "c-work-image img-fluid js-lazy rounded",
        data_bg: image_url_1x,
        data_bg_hidpi: image_url_2x,
        style: "background-color: #{bg_color};"
      ) do
        el.img src: dummy_src
      end
    end.html_safe
  end

  private

  attr_reader :image_url_1x, :image_url_2x, :bg_color

  def dummy_src
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 3 4'%3E%3C/svg%3E"
  end
end
