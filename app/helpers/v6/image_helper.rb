# frozen_string_literal: true

module V6::ImageHelper
  def v6_ann_image_url(record, field, height:, width:, blur: 0, format: "jpg")
    path = record ? record.uploaded_file_path(field) : "no-image.jpg"

    ix_image_url(path, {
      blur: blur,
      fill: "solid",
      fit: "fill",
      fm: format,
      height: height,
      w: width
    })
  end
end
