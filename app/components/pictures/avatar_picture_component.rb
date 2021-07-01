# frozen_string_literal: true

module Pictures
  class AvatarPictureComponent < ApplicationV6Component
    def initialize(view_context, user:, width:, alt: "", class_name: "")
      super view_context
      @user = user
      @width = width
      @height = @width
      @alt = alt.presence || "@#{user.username}"
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :picture, class: "c-avatar-picture" do
          h.tag :source, type: "image/webp", srcset: srcset(@height, @width, "webp")
          h.tag :source, type: "image/jpeg", srcset: srcset(@height, @width, "jpg")
          h.tag :img, {
            alt: @alt,
            class: "img-thumbnail rounded-circle #{@class_name}",
            height: @height,
            loading: "lazy",
            src: ann_image_url(@user.profile, :image, width: @width, format: "jpg"),
            width: @width
          }
        end
      end
    end

    private

    def srcset(height, width, format)
      [
        "#{ann_image_url(@user.profile, :image, width: width, format: format)} 1x",
        "#{ann_image_url(@user.profile, :image, width: width * 2, format: format)} 2x"
      ].join(", ")
    end
  end
end
