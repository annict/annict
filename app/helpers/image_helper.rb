# frozen_string_literal: true

module ImageHelper
  def ann_image_url(record, field, format:, height:, width:)
    path = record ? record.uploaded_file_path(field) : "no-image.jpg"

    ix_image_url(path, {
      fill: "solid",
      fit: "fill",
      fm: format,
      height: height,
      w: width,
    })
  end

  def ann_image_tag(record, field, options = {})
    url = ann_image_url(record, field, options)
    url2x = ann_image_url(record, field, options.merge(size_rate: 2))

    options["data-src"] = url
    options[:class] = if options[:class].present?
      options[:class].split(" ").push("js-lazy").join(" ")
    else
      "js-lazy"
    end
    options["data-srcset"] = "#{url} 320w, #{url2x} 640w"

    image_tag("", options)
  end

  def profile_background_image_url(profile, options)
    background_image = profile.background_image
    field = background_image ? :background_image : :image
    image = profile.send(field)

    if background_image.present? && profile.background_image_animated?
      return "#{ENV.fetch("ANNICT_FILE_STORAGE_URL")}/shrine/#{image[:original].id}"
    end

    ann_image_url(profile, field, options)
  end

  def ann_api_assets_url(record, field)
    path = image_path(record, field)
    "#{ENV.fetch("ANNICT_API_ASSETS_URL")}/#{path}"
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
    size = case size
    when "size50" then "50x50"
    when "size100" then "100x100"
    when "size150" then "150x150"
    when "size200" then "200x200"
    else
      "200x200"
    end

    ann_image_url(profile, :image, size: size, ratio: "1:1")
  end

  private

  def image_path(record, field)
    id = record&.uploaded_file(field)&.id
    path = id ? "shrine/#{id}" : ""
    path.presence || "no-image.jpg"
  end
end
