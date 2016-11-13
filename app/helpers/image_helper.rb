# frozen_string_literal: true

module ImageHelper
  def ann_image_url(record, field, options = {})
    path = record&.send(field)&.url(:master).presence || "/no-image.jpg"

    msize = options[:msize]
    size = browser.device.mobile? && msize.present? ? msize : options[:size]
    width, height = size.split("x").map do |s|
      return nil if s.blank?
      s.to_i * 2
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
end
