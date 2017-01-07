# frozen_string_literal: true

module Gonable
  extend ActiveSupport::Concern

  included do
    before_action :load_data_into_gon

    def load_i18n_into_gon(keys)
      gon.I18n = {}
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
        userId: user_signed_in? ? current_user.encoded_id : nil
      },
      keen: {
        projectId: ENV.fetch("KEEN_PROJECT_ID"),
        writeKey: ENV.fetch("KEEN_WRITE_KEY")
      }
    }

    if user_signed_in?
      data[:user].merge!(
        shareRecordToTwitter: current_user.setting.share_record_to_twitter?,
        shareRecordToFacebook: current_user.setting.share_record_to_facebook?,
        sharableToTwitter: current_user.shareable_to?(:twitter),
        sharableToFacebook: current_user.shareable_to?(:facebook)
      )
    end

    gon.push(data)
  end
end
