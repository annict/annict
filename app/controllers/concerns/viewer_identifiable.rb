# frozen_string_literal: true

module ViewerIdentifiable
  extend ActiveSupport::Concern

  included do
    def viewer_uuid
      cookies[:ann_client_uuid].presence || cookies[:ann_viewer_uuid].presence || store_client_uuid
    end

    def current_viewer
      return current_user if user_signed_in?

      @current_viewer ||= Guest.where(uuid: viewer_uuid).first_or_create(
        user_agent: request.user_agent,
        remote_ip: request.remote_ip,
        time_zone: cookies["ann_time_zone"].presence || "Asia/Tokyo",
        locale: locale
      )
    end

    private

    def store_viewer_uuid
      return if !cookies[:ann_client_uuid] && !cookies[:ann_viewer_uuid]

      cookies[:ann_viewer_uuid] = {
        value: request.uuid,
        expires: 2.year.from_now,
        domain: ".#{request.domain}"
      }

      cookies[:ann_viewer_uuid]
    end
  end
end
