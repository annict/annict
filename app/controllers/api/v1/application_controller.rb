# frozen_string_literal: true

module Api
  module V1
    class ApplicationController < ActionController::Base
      include ViewerIdentifiable
      include Analyzable
      include LogrageSetting
      include RavenContext
      include PageCategoryMethods

      rescue_from ActiveRecord::RecordNotFound, with: :not_found

      attr_reader :current_user

      before_action :store_page_category
      before_action -> { doorkeeper_authorize! :read }, only: %i(index show)
      before_action only: %i(create update destroy) do
        doorkeeper_authorize! :write
      end
      skip_before_action :verify_authenticity_token

      def not_found
        error = {
          type: "unknown_route",
          message: "リクエストに失敗しました",
          developer_message: "404 Not Found"
        }
        render json: { errors: [error] }, status: 404
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
        errors = @params.errors.full_messages.map do |message|
          {
            type: "invalid_params",
            message: "リクエストに失敗しました",
            developer_message: message
          }
        end

        render json: { errors: errors }, status: 400
      end
    end
  end
end
