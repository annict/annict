# frozen_string_literal: true

module Pictures
  class AvatarPictureComponent < ApplicationComponent
    def initialize(user:, width:, alt: "", class_name: "")
      @user = user
      @width = width
      @height = @width
      @alt = alt.presence || "@#{user.username}"
      @class_name = class_name
    end

    private

    def source_srcset(format)
      [
        "#{helpers.ann_avatar_image_url(@user, width: @width, format: format)} 1x",
        "#{helpers.ann_avatar_image_url(@user, width: @width * 2, format: format)} 2x"
      ].join(", ")
    end
  end
end
