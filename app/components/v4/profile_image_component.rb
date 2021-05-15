# frozen_string_literal: true

module V4
  class ProfileImageComponent < ApplicationComponent
    def initialize(image_url_1x:, alt:, bg_color: "#F1F1F1", class_name: "", img_options: {})
      @image_url_1x = image_url_1x
      @alt = alt
      @bg_color = bg_color
      @class_name = class_name
      @img_options = img_options
    end

    def call
      helpers.image_tag(image_url_1x, {
        alt: alt,
        class: img_class_name
      }.merge(img_options))
    end

    private

    attr_reader :alt, :bg_color, :class_name, :image_url_1x, :img_options

    def img_class_name
      classes = %w[c-profile-image img-fluid img-thumbnail rounded-circle]
      classes += class_name.split(" ")
      classes.join(" ")
    end

    def dummy_src
      "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1 1'%3E%3C/svg%3E"
    end
  end
end
