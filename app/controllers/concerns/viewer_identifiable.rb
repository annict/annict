# frozen_string_literal: true

module ViewerIdentifiable
  extend ActiveSupport::Concern

  included do
    helper_method :viewer_uuid

    def viewer_uuid
      cookies[:ann_client_uuid].presence || cookies[:ann_viewer_uuid].presence || store_client_uuid
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
