# frozen_string_literal: true

module ApplicationHelper
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
    locale == :ja
  end

  def local_time_ago_in_words(from_time, options = {})
    spacer = I18n.locale == :en ? " " : ""
    "#{time_ago_in_words(from_time, options)}#{spacer}#{I18n.t('words.ago')}"
  end
end
