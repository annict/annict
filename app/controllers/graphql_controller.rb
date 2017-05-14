# frozen_string_literal: true

class GraphqlController < ActionController::Base
  before_action :doorkeeper_authorize!

  rescue_from ActionController::InvalidAuthenticityToken, with: :bad_credentials

  def execute
    variables = ensure_hash(params[:variables])
    query = params[:query]
    context = {
      doorkeeper_token: doorkeeper_token
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
