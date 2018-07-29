# frozen_string_literal: true

module Api
  module Internal
    class GraphqlController < ActionController::Base
      include GraphqlMethods
      include Analyzable
      include LogrageSetting
      include RavenContext

      skip_before_action :verify_authenticity_token

      def execute
        query = params[:query]
        variables = ensure_hash(params[:variables])
        context = {
          viewer: current_user,
          internal: true,
          ga_client: ga_client
        }
        result = AnnictSchema.execute(query, variables: variables, context: context)
        render json: result
      end
    end
  end
end
