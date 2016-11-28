# frozen_string_literal: true

module ControllerCommon
  extend ActiveSupport::Concern

  included do
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

    def load_i18n_into_gon(keys)
      gon.I18n = {}
      keys.each do |k, v|
        key = v.present? && browser.device.mobile? && v.key?(:mobile) ? v[:mobile] : k
        gon.I18n[k] = I18n.t(key)
      end
    end

    def ga_client
      @ga_client ||= Annict::Analytics::Client.new(request, current_user)
    end

    def keen_client
      @keen_client ||= Annict::Keen::Client.new(request)
    end

    def store_client_uuid
      return if cookies[:ann_client_uuid].present?

      cookies[:ann_client_uuid] = {
        value: request.uuid,
        expires: 2.year.from_now,
        domain: ".#{ENV.fetch('ANNICT_DOMAIN')}"
      }

      cookies[:ann_client_uuid]
    end

    def client_uuid
      cookies[:ann_client_uuid].presence || store_client_uuid
    end

    def render_jb(path, assigns)
      ApplicationController.render("#{path}.jb", assigns: assigns)
    end
  end
end
