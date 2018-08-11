# frozen_string_literal: true

module Api
  class GraphqlController < ActionController::Base
    include Analyzable
    include LogrageSetting
    include RavenContext

    skip_before_action :verify_authenticity_token

    rescue_from ActionController::InvalidAuthenticityToken, with: :bad_credentials

    def execute
      variables = ensure_hash(params[:variables])
      query = params[:query]
      context = {
        access_token: access_token,
        oauth_application: access_token.application,
        viewer: current_user,
        ga_client: ga_client
      }
      result = AnnictSchema.execute(query, variables: variables, context: context)
      render json: result
    end

    private

    def access_token
      doorkeeper_token.presence || guest_access_token
    end

    def guest_access_token
      @guest_access_token ||= GuestAccessToken.new
    end

    def current_user
      @current_user ||= access_token.owner
    end

    def bad_credentials
      json = {
        "message": "Bad credentials"
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
