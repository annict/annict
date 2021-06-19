# frozen_string_literal: true

module ImageV6Helper
  def ann_image_url(record, field, width:, blur: 0, format: "jpg")
    height = height(record, field, width)
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

  private

  def height(record, field, width)
    case [record.class, field]
    when [AnimeImage, :image] then ((4 * width) / 3).ceil
    when [Profile, :image] then width
    when [Trailer, :image] then ((9 * width) / 16).ceil
    end
  end
end
