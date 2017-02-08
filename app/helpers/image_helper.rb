# frozen_string_literal: true

module ImageHelper
  def ann_image_url(record, field, options = {})
    path = record&.send(field)&.path(:master).presence || "/no-image.jpg"

    msize = options[:msize]
    size = browser.device.mobile? && msize.present? ? msize : options[:size]
    width, height = size.split("x").map do |s|
      s.present? ? (s.to_i * 2) : nil
    end

    ix_options = {
      auto: "format"
    }
    ix_options[:w] = width if width.present?
    ix_options[:h] = height if height.present?

    ix_image_url(path, ix_options)
  end

  def ann_image_tag(record, field, options = {})
    url = ann_image_url(record, field, options)

    msize = options[:msize]
    options[:size] = msize if browser.device.mobile? && msize.present?
    options.delete(:msize) if options.key?(:msize)

    image_tag(url, options)
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
end
