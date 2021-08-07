# frozen_string_literal: true

module ImageV6Helper
  def ann_image_url(record, field, width:, format: "jpg")
    path = record ? record.uploaded_file_path(field) : "no-image.jpg"
    height = case [record.class, field]
      when [AnimeImage, :image] then ((4 * width) / 3).ceil
      when [Profile, :image] then width
      when [Trailer, :image] then ((9 * width) / 16).ceil
    end
    fit = width == height ? "crop" : "fill"

    ix_image_url(path, {
      fill: "solid",
      fit: fit,
      fm: format,
      height: height,
      w: width
    })
  end
end
