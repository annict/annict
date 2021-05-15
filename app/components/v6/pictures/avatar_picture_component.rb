# frozen_string_literal: true

module V6::Pictures
  class AvatarPictureComponent < V6::ApplicationComponent
    def initialize(view_context, user:, width:, mb_width:, alt: "", class_name: "")
      super view_context
      @user = user
      @width = width
      @height = @width
      @mb_width = mb_width
      @mb_height = @mb_width
      @alt = alt.presence || "@#{user.username}"
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :picture, class: "c-avatar-picture" do
          h.tag :source, media: "(min-width: 576px)", type: "image/webp", srcset: srcset(@height, @width, "webp")
          h.tag :source, media: "(min-width: 576px)", type: "image/jpeg", srcset: srcset(@height, @width, "jpg")
          h.tag :source, type: "image/webp", srcset: srcset(@mb_height, @mb_width, "webp")
          h.tag :source, type: "image/jpeg", srcset: srcset(@mb_height, @mb_width, "jpg")
          h.tag :img, {
            alt: @alt,
            class: "img-thumbnail rounded-circle #{@class_name}",
            height: @height,
            loading: "lazy",
            src: view_context.v6_ann_image_url(@user.profile, :image, height: @height, width: @width, format: "jpg"),
            width: @width
          }
        end
      end
    end

    private

    def srcset(height, width, format)
      [
        "#{view_context.v6_ann_image_url(@user.profile, :image, height: height, width: width, format: format)} 1x",
        "#{view_context.v6_ann_image_url(@user.profile, :image, height: height * 2, width: width * 2, format: format)} 2x"
      ].join(", ")
    end
  end
end
