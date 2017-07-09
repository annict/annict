# frozen_string_literal: true

module Gonable
  extend ActiveSupport::Concern

  included do
    def load_i18n_into_gon(keys)
      gon.I18n = gon.I18n.presence || {}

      keys.each do |k, v|
        key = v.present? && browser.device.mobile? && v.key?(:mobile) ? v[:mobile] : k
        gon.I18n[k] = I18n.t(key)
      end
    end
  end

  private

  def load_data_into_gon
    data = {
      user: {
        clientUUID: client_uuid,
        device: browser.device.mobile? ? "mobile" : "pc",
        locale: locale,
        userId: user_signed_in? ? current_user.encoded_id : nil,
        isSignedIn: user_signed_in?
      },
      I18n: default_i18n_data,
      facebook: {
        appId: ENV.fetch("FACEBOOK_APP_ID")
      }
    }

    if user_signed_in?
      data[:user].merge!(
        shareRecordToTwitter: current_user.setting.share_record_to_twitter?,
        shareRecordToFacebook: current_user.setting.share_record_to_facebook?,
        sharableToTwitter: current_user.authorized_to?(:twitter, shareable: true),
        sharableToFacebook: current_user.authorized_to?(:facebook, shareable: true)
      )
    end

    gon.push(data)
  end

  def default_i18n_data
    username_preview_val = if browser.device.mobile?
      I18n.t("messages.registrations.new.username_preview_mobile")
    else
      I18n.t("messages.registrations.new.username_preview")
    end

    {
      "messages.registrations.new.username_preview": username_preview_val
    }
  end
end
