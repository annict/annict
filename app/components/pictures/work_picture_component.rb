# frozen_string_literal: true

module Pictures
  class WorkPictureComponent < ApplicationComponent
    def initialize(work:, width:, alt: "", class_name: "")
      @work = work
      @work_image = work.work_image
      @width = width
      @height = @work_image&.image_height(:image, @width)
      @alt = alt.presence || work.local_title
      @class_name = class_name
    end

    private

    def source_srcset(format)
      [
        "#{helpers.ann_image_url(@work_image, :image, width: @width, format: format)} 1x",
        "#{helpers.ann_image_url(@work_image, :image, width: @width * 2, format: format)} 2x"
      ].join(", ")
    end
  end
end
