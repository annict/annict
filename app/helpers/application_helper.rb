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

  def local_time_ago_in_words(from_time, options = {})
    days = (Time.zone.now.to_date - from_time.to_date).to_i
    return display_time(from_time) if days > 3
    spacer = I18n.locale == :en ? " " : ""
    "#{time_ago_in_words(from_time, options)}#{spacer}#{I18n.t('noun.ago')}"
  end

  def local_datetime(datetime)
    if user_signed_in? && current_user.time_zone.present?
      return datetime&.in_time_zone(current_user.time_zone)&.strftime("%Y-%m-%d %H:%M")
    end
    datetime&.strftime("%Y-%m-%d %H:%M %:z")
  end

  def twitter_username
    I18n.locale == :ja ? "@AnnictJP" : "@anannict"
  end

  def show_privacy_policy_modal?
    user_signed_in? && !current_user.setting.privacy_policy_agreed?
  end

  def annict_config
    config = {
      currentUrlJa: local_current_url(locale: :ja),
      currentUrlEn: local_current_url(locale: :en),
      facebook: {
        appId: ENV.fetch("FACEBOOK_APP_ID")
      },
      images: {
        logoUrl: asset_bundle_url("images/logos/color-mizuho.png")
      },
      isLocaleJa: locale_ja?,
      localUrl: local_url,
      season: {
        current: ENV.fetch("ANNICT_CURRENT_SEASON"),
        next: ENV.fetch("ANNICT_NEXT_SEASON"),
        prev: ENV.fetch("ANNICT_PREVIOUS_SEASON")
      },
      twitter: {
        username: twitter_username
      }
    }

    javascript_tag "window.annConfig = #{config.to_json.html_safe};"
  end
end
