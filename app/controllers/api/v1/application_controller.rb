# frozen_string_literal: true

module Api
  module V1
    class ApplicationController < ActionController::Base
      include Analyzable
      include SentryLoadable

      rescue_from ActiveRecord::RecordNotFound, with: :not_found

      attr_reader :current_user

      before_action :set_sentry_context
      before_action -> { doorkeeper_authorize! :read }, only: %i[index show]
      before_action only: %i[create update destroy] do
        doorkeeper_authorize! :write
      end
      before_action :logging_request
      skip_before_action :verify_authenticity_token

      def not_found
        error = {
          type: "unknown_route",
          message: "リクエストに失敗しました",
          developer_message: "404 Not Found"
        }
        render json: {errors: [error]}, status: 404
      end

      private

      def current_user
        return nil if doorkeeper_token.blank?
        @current_user ||= User.find(doorkeeper_token.resource_owner_id)
      end

      def prepare_params!
        class_name = self.class.name
        class_name = class_name.sub("Controller", "#{params[:action].classify}Params")
        @params = Object.const_get(class_name).new(params)
        return response_params_error unless @params.valid?
      end

      def response_params_error
        errors = @params.errors.full_messages.map { |message|
          {
            type: "invalid_params",
            message: "リクエストに失敗しました",
            developer_message: message
          }
        }

        render json: {errors: errors}, status: 400
      end

      def logging_request
        annict_logger.log(:info, :REST_API_REQUEST, oauth_access_token_id: doorkeeper_token&.id)
      end
    end
  end
end
