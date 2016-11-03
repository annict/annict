# frozen_string_literal: true

module ApplicationHelper
  def annict_image_url(record, field, options = {})
    path = record.try!(:send, field).try!(:path, :master).presence || "no-image.jpg"
    path = path.sub(%r{\A.*paperclip/}, "paperclip/") unless Rails.env.production?

    msize = options[:msize]
    size = browser.device.mobile? && msize.present? ? msize : options[:size]
    width, height = size.split("x").map { |s| s.to_i * 2 }

    blur = options[:blur].presence || 0

    ix_image_url(path, w: width, h: height, fit: "crop", auto: "format", blur: blur)
  end

  def annict_image_tag(record, field, options = {})
    url = annict_image_url(record, field, options)

    msize = options[:msize]
    options[:size] = msize if browser.device.mobile? && msize.present?
    options.delete(:msize) if options.key?(:msize)

    image_tag(url, options)
  end

  def custom_time_ago_in_words(datetime)
    days = (Time.zone.now.to_date - datetime.to_date).to_i

    if days > 3
      datetime.strftime("%Y/%m/%d")
    else
      "#{time_ago_in_words(datetime)}#{t('words.ago')}"
    end
  end

  def body_classes
    controller_name = controller.controller_path.tr("/", "-")
    basic_body_classes = [
      "p-#{controller_name}",
      "p-#{controller_name}-#{controller.action_name}"
    ].join(" ")

    if content_for?(:extra_body_classes)
      [basic_body_classes, content_for(:extra_body_classes)].join(" ")
    else
      basic_body_classes
    end
  end

  def locale_ja?
    locale == :ja || (user_signed_in? && current_user.role.admin?)
  end

  def local_time_ago_in_words(from_time, options = {})
    spacer = I18n.locale == :en ? " " : ""
    "#{time_ago_in_words(from_time, options)}#{spacer}#{I18n.t('words.ago')}"
  end
end
