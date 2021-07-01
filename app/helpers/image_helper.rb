# frozen_string_literal: true

module ImageHelper
  def v4_ann_image_url(record, field, options = {})
    path = image_path(record, field)
    size = options[:size]

    width, = size.split("x").map do |s|
      s.present? ? (s.to_i * (options[:size_rate].presence || 1)) : nil
    end

    ix_options = {
      auto: "format"
    }

    if width
      ix_options[:w] = width
    end

    if options[:crop]
      ix_options[:fit] = "crop"
    end

    if options[:ratio]
      ix_options[:fit] = "crop"
      ix_options[:ar] = options[:ratio]
    end

    if options[:blur]
      ix_options[:blur] = options[:blur]
    end

    ix_image_url(path, ix_options)
  end

  def ann_image_tag(record, field, options = {})
    url2x = v4_ann_image_url(record, field, options.merge(size_rate: 2))

    image_tag(url2x, options)
  end

  def profile_background_image_url(profile, options)
    background_image = profile.background_image
    field = background_image ? :background_image : :image
    image = profile.send(field)

    if background_image.present? && profile.background_image_animated?
      return "#{ENV.fetch("ANNICT_FILE_STORAGE_URL")}/shrine/#{image[:original].id}"
    end

    v4_ann_image_url(profile, field, options)
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

    v4_ann_image_url(profile, :image, size: size, ratio: "1:1")
  end

  private

  def image_path(record, field)
    id = record&.uploaded_file(field)&.id
    path = id ? "shrine/#{id}" : ""
    path.presence || "no-image.jpg"
  end
end
