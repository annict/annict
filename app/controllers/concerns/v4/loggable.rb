# frozen_string_literal: true

module V4
  module Loggable
    extend ActiveSupport::Concern

    def annict_logger
      @annict_logger ||= Annict::Logger.new(request, current_user)
    end

    def set_client_uuid
      return if cookies[:ann_client_uuid].present?

      cookies[:ann_client_uuid] = {
        value: request.uuid,
        expires: 2.year.from_now,
        domain: ".#{request.domain}"
      }
    end

    def client_uuid
      @client_uuid ||= begin
        cookies[:ann_client_uuid].presence || set_client_uuid
      end
    end

    # Add custom variables to the event payload
    # https://github.com/roidrage/lograge
    def append_info_to_payload(payload)
      super
      payload[:request_id] = request.uuid
      payload[:client_uuid] = client_uuid
      payload[:user_id] = current_user&.id
    end
  end
end
