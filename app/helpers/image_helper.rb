# frozen_string_literal: true

module ImageHelper
  def ann_image_url(record, field, options = {})
    path = image_path(record, field)
    size = options[:size]
    ratio = options[:ratio].presence || "1:1"

    width, height = size.split("x").map do |s|
      s.present? ? (s.to_i * (options[:size_rate].presence || 1)) : nil
    end

    ix_options = {
      auto: "format"
    }

    if width
      ix_options[:w] = width
    end

    unless height
      ix_options[:h] = (width * ratio.split(":")[1].to_i) / ratio.split(":")[0].to_i
    end

    if record.nil? || (record.instance_of?(WorkImage) && ratio == "3:4")
      ix_options[:fit] = "fillmax"
      ix_options[:fill] = "blur"
    end

    if options[:blur]
      ix_options[:blur] = options[:blur]
    end

    ix_image_url(path, ix_options)
  end

  def ann_image_tag(record, field, options = {})
    url = ann_image_url(record, field, options)
    url2x = ann_image_url(record, field, options.merge(size_rate: 2))

    options["v-lazy"] = "{ src: '#{url}' }"
    options[:class] = if options[:class].present?
      options[:class].split(" ").push("c-vue-lazyload").join(" ")
    else
      "c-vue-lazyload"
    end
    options["data-srcset"] = "#{url} 320w, #{url2x} 640w"

    image_tag("", options)
  end

  def profile_background_image_url(profile, options)
    background_image = profile.background_image
    field = background_image ? :background_image : :image
    image = profile.send(field)

    if background_image.present? && profile.background_image_animated?
      return "#{ENV.fetch('ANNICT_FILE_STORAGE_URL')}/shrine/#{image[:original].id}"
    end

    ann_image_url(profile, field, options)
  end

  def ann_api_assets_url(record, field)
    path = image_path(record, field)
    "#{ENV.fetch('ANNICT_API_ASSETS_URL')}/#{path}"
  end

  def ann_api_assets_background_image_url(profile)
    background_image = profile.background_image
    field = background_image ? :background_image : :image
    image = profile.send(field)

    if background_image.present? && record.background_image_animated?
      return "#{ENV.fetch('ANNICT_API_ASSETS_URL')}/shrine/#{image[:original].id}"
    end

    ann_api_assets_url(profile, field)
  end

  private

  def image_path(record, field)
    path = if Rails.env.test?
      record&.send(field)&.url(:master)
    else
      id = record&.uploaded_file(field)&.id
      id ? "shrine/#{id}" : ""
    end

    path.presence || "no-image.jpg"
  end
end
