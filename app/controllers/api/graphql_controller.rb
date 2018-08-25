# frozen_string_literal: true

module Api
  class GraphqlController < ActionController::Base
    include ViewerIdentifiable
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
        viewer: doorkeeper_token.owner,
        ga_client: ga_client
      }
      result = AnnictSchema.execute(query, variables: variables, context: context)
      render json: result
    end

    private

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
