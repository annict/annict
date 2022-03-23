# frozen_string_literal: true

module ImageUploadable
  extend ActiveSupport::Concern

  def uploaded_file(field, size: :master)
    send(field).instance_of?(Hash) ? send(field)[size] : send(field)
  end

  def uploaded_file_path(field)
    id = uploaded_file(field)&.id

    id ? "shrine/#{id}" : nil
  end

  def origin_image_url(field)
    image_path = uploaded_file_path(field).presence || "no-image.jpg"

    "s3://#{ENV.fetch('S3_BUCKET_NAME')}/#{image_path}"
  end

  def image_aspect_ratio(field)
    raise NotImplementedError
  end

  def image_height(field, width)
    ratio_w, ratio_h = image_aspect_ratio(field).split(":")

    ((ratio_w.to_i * width.to_i) / ratio_h.to_i).ceil
  end
end
