# frozen_string_literal: true

module Api
  module Canary
    class GraphqlController < ActionController::Base
      include Analyzable
      include SentryLoadable

      before_action :doorkeeper_authorize!
      skip_before_action :verify_authenticity_token

      rescue_from ActionController::InvalidAuthenticityToken, with: :bad_credentials

      def execute
        variables = ensure_hash(params[:variables])
        query = params[:query]
        context = {
          writable: doorkeeper_token.writable?,
          application: doorkeeper_token.application,
          admin: doorkeeper_token.application&.owner&.role&.admin?,
          viewer: current_user
        }
        result = ::Canary::AnnictSchema.execute(query, variables: variables, context: context)
        annict_logger.log(
          :info,
          :GRAPHQL_API_REQUEST,
          oauth_access_token_id: doorkeeper_token.id,
          query: query,
          variables: variables
        )
        render json: result
      end

      private

      def current_user
        return nil if doorkeeper_token.blank?

        @current_user ||= User.only_kept.find(doorkeeper_token.resource_owner_id)
      end

      def bad_credentials
        json = {
          message: "Bad credentials"
        }
        render json: json, status: 401
      end

      def doorkeeper_unauthorized_render_options(_error)
        {
          json: {
            message: "Not authorized"
          }
        }
      end

      # Handle form data, JSON body, or a blank value
      def ensure_hash(ambiguous_param)
        case ambiguous_param
        when String
          if ambiguous_param.present?
            ensure_hash(JSON.parse(ambiguous_param))
          else
            {}
          end
        when Hash, ActionController::Parameters
          ambiguous_param
        when nil
          {}
        else
          raise ArgumentError, "Unexpected parameter: #{ambiguous_param}"
        end
      end
    end
  end
end
