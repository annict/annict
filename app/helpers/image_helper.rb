# frozen_string_literal: true

module ImageHelper
  def ann_image_url(record, field, width:, format: :webp, blur: nil)
    return unless record

    proxy_options = {
      format: format,
    }

    if record.image_aspect_ratio(field) == "1:1"
      proxy_options[:crop] = {
        width: width,
        height: width
      }
    else
      proxy_options = proxy_options.merge(
        width: width,
        height: record.image_height(field, width)
      )
    end

    if blur
      proxy_options[:blur] = blur
    end

    Imgproxy.url_for(record.origin_image_url(field), **proxy_options)
  end

  def ann_image_tag(record, field, options = {})
    url2x = v4_ann_image_url(record, field, options.merge(size_rate: 2))

    image_tag(url2x, options)
  end

  def ann_api_assets_url(record, field)
    "#{ENV.fetch("ANNICT_API_ASSETS_URL")}/#{record.uploaded_file_path(field)}"
  end

  def ann_api_assets_background_image_url(profile)
    background_image = profile.background_image
    field = background_image ? :background_image : :image
    image = profile.send(field)

    if background_image.present? && profile.background_image_animated?
      return "#{ENV.fetch("ANNICT_API_ASSETS_URL")}/shrine/#{image[:original].id}"
    end

    ann_api_assets_url(profile, field)
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

    ann_image_url(profile, :image, width: width)
  end
end
