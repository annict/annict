# frozen_string_literal: true

module Pictures
  class AnimePictureComponent < ApplicationV6Component
    def initialize(view_context, anime:, width:, alt: "", class_name: "")
      super view_context
      @anime = anime
      @width = width
      @height = @anime.anime_image_height(@width)
      @alt = alt.presence || @anime.local_title
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :picture, class: "c-anime-picture" do
          h.tag :source, type: "image/webp", srcset: srcset("webp")
          h.tag :source, type: "image/jpeg", srcset: srcset("jpg")
          h.tag :img, {
            alt: @alt,
            class: "img-thumbnail #{@class_name}",
            height: @height,
            loading: "lazy",
            src: ann_image_url(@anime.anime_image, :image, width: @width, height: @height, format: "jpg"),
            width: @width
          }
        end
      end
    end

    private

    def srcset(format)
      [
        "#{ann_image_url(@anime.anime_image, :image, width: @width, height: @height, format: format)} 1x",
        "#{ann_image_url(@anime.anime_image, :image, width: @width * 2, height: @height * 2, format: format)} 2x"
      ].join(", ")
    end
  end
end
