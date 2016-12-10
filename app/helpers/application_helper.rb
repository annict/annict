# frozen_string_literal: true

module ApplicationHelper
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
    locale == :ja
  end

  def local_time_ago_in_words(from_time, options = {})
    days = (Time.zone.now.to_date - from_time.to_date).to_i
    return from_time.strftime("%Y/%m/%d") if days > 3
    spacer = I18n.locale == :en ? " " : ""
    "#{time_ago_in_words(from_time, options)}#{spacer}#{I18n.t('words.ago')}"
  end

  def local_datetime(datetime)
    if user_signed_in? && current_user.time_zone.present?
      return datetime&.in_time_zone(current_user.time_zone)&.strftime("%Y-%m-%d %H:%M")
    end
    datetime&.strftime("%Y-%m-%d %H:%M %:z")
  end
end
