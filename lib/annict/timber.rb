# frozen_string_literal: true

module Annict
  class Timber
    ACTIONS = {
      EPISODE_RECORD_CREATE: "episode_record.create",
      GRAPHQL_API_REQUEST: "graphql_api.request",
      REST_API_REQUEST: "rest_api.request",
      STATUS_CREATE: "status.create",
      WORK_RECORD_CREATE: "work_record.create"
    }.freeze

    attr_writer :page_category

    def initialize(request, user)
      @request = request
      @user = user
    end

    def log(type, action, metadata = {})
      if Rails.env.production?
        Rails.logger.send(type, "Send log to Timber: #{action}", data: build_metadata(action, metadata))
      else
        puts "Send log to Timber: #{build_metadata(action, metadata).merge(type: type)}"
      end
    end

    private

    def build_metadata(action, metadata = {})
      metadata.merge(
        action: ACTIONS.fetch(action),
        page_category: @page_category,
        request: {
          path: @request.path,
          params: @request.params.slice(:work_id, :episode_id)
        },
        viewer: {
          type: @user.present? ? "user" : "guest"
        },
        user: {
          id: @user&.id,
          username: @user&.username
        }
      )
    end
  end
end
