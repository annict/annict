# frozen_string_literal: true

module API
  module Internal
    class GraphQLController < ActionController::Base
      include RavenContext

      skip_before_action :verify_authenticity_token
      before_action :set_raven_context

      def execute
        variables = ensure_hash(params[:variables])
        query = params[:query]
        context = {
          writable: true,
          admin: true,
          viewer: current_user
        }
        result = ::Canary::AnnictSchema.execute(query, variables: variables, context: context)

        render json: result
      end

      private

      def current_user
        User.find_by(id: params[:user_id])
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
