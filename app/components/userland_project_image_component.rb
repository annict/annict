# frozen_string_literal: true

class UserlandProjectImageComponent < ApplicationComponent
  def initialize(image_url_1x:, alt:)
    @image_url_1x = image_url_1x
    @alt = alt
  end

  def call
    helpers.image_tag(dummy_src, {
      alt: alt,
      class: "img-fluid img-thumbnail js-lazy rounded",
      "data-src": image_url_1x
    })
  end

  private

  attr_reader :alt, :image_url_1x

  def dummy_src
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1 1'%3E%3C/svg%3E"
  end
end
