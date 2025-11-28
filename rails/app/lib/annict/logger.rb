# typed: false
# frozen_string_literal: true

module Annict
  class Logger
    ACTIONS = {
      GRAPHQL_API_REQUEST: "graphql_api.request",
      REST_API_REQUEST: "rest_api.request"
    }.freeze

    attr_writer :page_category

    def initialize(request, user)
      @request = request
      @user = user
    end

    def log(type, action, metadata = {})
      Rails.logger.send(type, build_metadata(action, metadata).to_json)
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
