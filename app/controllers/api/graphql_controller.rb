# frozen_string_literal: true

module Api
  class GraphqlController < ActionController::Base
    include GraphqlMethods
    include Analyzable
    include LogrageSetting
    include RavenContext

    before_action :doorkeeper_authorize!
    skip_before_action :verify_authenticity_token

    rescue_from ActionController::InvalidAuthenticityToken, with: :bad_credentials

    def execute
      variables = ensure_hash(params[:variables])
      query = params[:query]
      context = {
        access_token: doorkeeper_token,
        oauth_application: doorkeeper_token.application,
        viewer: current_user,
        ga_client: ga_client
      }
      result = AnnictSchema.execute(query, variables: variables, context: context)
      render json: result
    end

    private

    def current_user
      return if doorkeeper_token.nil?
      @current_user ||= doorkeeper_token.owner
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
  end
end
