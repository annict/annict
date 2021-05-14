# frozen_string_literal: true

module Deprecated
  class WorkImageComponent < Deprecated::ApplicationComponent
    def initialize(image_url_1x:, image_url_2x:, alt: "Work Image", bg_color: "#F1F1F1", class_name: "")
      @image_url_1x = image_url_1x
      @image_url_2x = image_url_2x
      @alt = alt
      @bg_color = bg_color
      @class_name = class_name
    end

    private

    attr_reader :alt, :class_name, :image_url_1x, :image_url_2x, :bg_color

    def dummy_src
      "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 3 4'%3E%3C/svg%3E"
    end
  end
end
