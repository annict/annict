# frozen_string_literal: true

class ProfileImageComponent < ApplicationComponent
  def initialize(image_url_1x:, alt:, bg_color: "#F1F1F1", img_options: {})
    @image_url_1x = image_url_1x
    @alt = alt
    @bg_color = bg_color
    @img_options = img_options
  end

  def call
    helpers.image_tag(dummy_src, {
      alt: alt,
      class: "img-fluid img-thumbnail js-lazy rounded-circle",
      "data-src": image_url_1x,
      style: "background-color: #{bg_color};"
    }.merge(img_options))
  end

  private

  attr_reader :alt, :bg_color, :image_url_1x, :img_options

  def dummy_src
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1 1'%3E%3C/svg%3E"
  end
end
