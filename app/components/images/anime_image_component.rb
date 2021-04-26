# frozen_string_literal: true

module Images
  class AnimeImageComponent < ApplicationComponent2
    def initialize(view_context, image_url_1x:, image_url_2x:, alt: "Work Image", bg_color: "#F1F1F1", class_name: "")
      super view_context
      @image_url_1x = image_url_1x
      @image_url_2x = image_url_2x
      @alt = alt
      @bg_color = bg_color
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div,
          class: "c-work-image img-fluid js-lazy rounded-sm #{@class_name}",
          data_bg: @image_url_1x,
          data_bg_hidpi: @image_url_2x,
          style: "background-color: #{@bg_color};" do
            h.html image_tag(dummy_src, alt: @alt)
          end
      end
    end

    private

    def dummy_src
      "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 3 4'%3E%3C/svg%3E"
    end
  end
end
