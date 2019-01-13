# frozen_string_literal: true

module Gonable
  extend ActiveSupport::Concern

  included do
    def load_i18n_into_gon(keys)
      gon.I18n = gon.I18n.presence || {}

      keys.each do |k, v|
        key = v.present? && !device_pc? && v.key?(:mobile) ? v[:mobile] : k
        gon.I18n[k] = I18n.t(key)
      end
    end
  end

  private

  def store_data_into_gon
    data = {
      user: {
        clientUUID: client_uuid,
        device: device_pc? ? "pc" : "mobile",
        locale: locale,
        userId: user_signed_in? ? current_user.encoded_id : nil,
        isSignedIn: user_signed_in?
      },
      I18n: default_i18n_data,
      annict: {
        domain: locale == :ja ? ENV.fetch("ANNICT_JP_DOMAIN") : ENV.fetch("ANNICT_DOMAIN"),
        url: locale == :ja ? ENV.fetch("ANNICT_JP_URL") : ENV.fetch("ANNICT_URL")
      },
      facebook: {
        appId: ENV.fetch("FACEBOOK_APP_ID")
      },
      rails: {
        env: Rails.env
      },
      page: {}
    }

    if user_signed_in?
      data[:user].merge!(
        shareRecordToTwitter: current_user.setting.share_record_to_twitter?,
        sharableToTwitter: current_user.authorized_to?(:twitter, shareable: true)
      )
    end

    gon.push(data)
  end

  def default_i18n_data
    username_preview_val = if !device_pc?
      I18n.t("messages.registrations.new.username_preview_mobile")
    else
      I18n.t("messages.registrations.new.username_preview")
    end

    {
      "messages.registrations.new.username_preview": username_preview_val,
      "messages._common.updated": I18n.t("messages._common.updated")
    }
  end
end
