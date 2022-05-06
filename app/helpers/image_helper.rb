# frozen_string_literal: true

module ImageHelper
  def ann_image_url(record, field, width:, ratio:, format: :webp, blur: nil)
    proxy_options = {
      width: width,
      height: image_height(width, ratio)
    }

    if format != :jpg
      proxy_options[:format] = format
    end

    if ratio == "1:1"
      proxy_options[:resizing_type] = "fill-down"
    end

    if blur
      proxy_options[:blur] = blur
    end

    Imgproxy.url_for(origin_image_url(record, field), **proxy_options)
  end

  def ann_work_image_url(work, width:, format: :webp, blur: nil)
    ann_image_url(work.work_image, :image, width: width, ratio: "4:3", format: format, blur: blur)
  end

  def ann_video_image_url(trailer, width:, format: :webp, blur: nil)
    ann_image_url(trailer, :image, width: width, ratio: "16:9", format: format, blur: blur)
  end

  def ann_avatar_image_url(user, width:, format: :webp, blur: nil)
    ann_image_url(user.profile, :image, width: width, ratio: "1:1", format: format, blur: blur)
  end

  def ann_project_image_url(project, width:, format: :webp, blur: nil)
    ann_image_url(project, :image, width: width, ratio: "1:1", format: format, blur: blur)
  end

  def api_user_avatar_url(profile, size)
    width = case size
    when "size50" then 50
    when "size100" then 100
    when "size150" then 150
    when "size200" then 200
    else
      200
    end

    ann_image_url(profile, :image, width: width, ratio: "1:1", format: :jpg)
  end

  def image_path(record, field)
    record&.uploaded_file_path(field).presence || "no-image.jpg"
  end

  def origin_image_url(record, field)
    "s3://#{ENV.fetch("S3_BUCKET_NAME")}/#{image_path(record, field)}"
  end

  def image_height(width, ratio)
    return width if ratio == "1:1"

    ratio_w, ratio_h = ratio.split(":")

    ((ratio_w.to_i * width.to_i) / ratio_h.to_i).ceil
  end
end
