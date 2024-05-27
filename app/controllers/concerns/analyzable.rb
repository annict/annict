# typed: false
# frozen_string_literal: true

module Analyzable
  extend ActiveSupport::Concern

  included do
    helper_method :client_uuid

    def ga_client
      @ga_client ||= Annict::Analytics::Client.new(request, current_user)
    end

    def annict_logger
      @annict_logger ||= Annict::Logger.new(request, current_user)
    end

    def store_client_uuid
      return if cookies[:ann_client_uuid].present?

      cookies[:ann_client_uuid] = {
        value: request.uuid,
        expires: 2.year.from_now,
        domain: ".#{request.domain}"
      }

      cookies[:ann_client_uuid]
    end

    def client_uuid
      cookies[:ann_client_uuid].presence || store_client_uuid
    end
  end
end
