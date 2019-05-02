# frozen_string_literal: true

module ImageHelper
  def ann_image_url(record, field, options = {})
    path = image_path(record, field)

    width, _height = options[:size].split("x").map do |s|
      s.present? ? (s.to_i * (options[:size_rate].presence || 1)) : nil
    end

    ix_options = {
      auto: "format"
    }
    ix_options[:w] = width if width.present?

    ix_image_url(path, ix_options)
  end

  def ann_email_image_url(record, field, options = {})
    path = image_path(record, field)
    width, height = options[:size]&.split("x")

    ix_options = {
      auto: "format"
    }
    ix_options[:w] = width if width.present?
    ix_options[:h] = height if height.present?

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
    background_image = profile.tombo_background_image
    field = background_image.present? ? :tombo_background_image : :tombo_avatar
    image = profile.send(field)

    if background_image.present? && profile.background_image_animated?
      path = image.path(:original).sub(%r{\A.*paperclip/}, "paperclip/")
      return "#{ENV.fetch('ANNICT_FILE_STORAGE_URL')}/#{path}"
    end

    ann_image_url(profile, field, options)
  end

  def ann_api_assets_url(record, field)
    path = image_path(record, field)
    "#{ENV.fetch('ANNICT_API_ASSETS_URL')}/#{path}"
  end

  def ann_api_assets_background_image_url(record, field)
    background_image = record&.send(field)
    field = background_image.present? ? :tombo_background_image : :tombo_avatar
    image = record&.send(field)

    if background_image.present? && record.background_image_animated?
      path = image.path(:original).sub(%r{\A.*paperclip/}, "paperclip/")
      return "#{ENV.fetch('ANNICT_API_ASSETS_URL')}/#{path}"
    end

    ann_api_assets_url(record, field)
  end

  private

  def image_path(record, field)
    path = if Rails.env.test?
      record&.send(field)&.url(:master)
    else
      id = record&.send(field)&.fetch(:master, nil)&.id
      id ? "shrine/#{id}" : ""
    end

    path.presence || "no-image.jpg"
  end
end
