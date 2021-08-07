# frozen_string_literal: true

module V4
  module LocalApi
    class GraphqlController < ActionController::Base
      include Localizable

      around_action :set_locale_with_params
      skip_before_action :verify_authenticity_token

      def execute
        render json: ::Canary::AnnictSchema.execute(params[:query], variables: variables, context: context)
      end

      private

      def variables
        @variables ||= ensure_hash(params[:variables])
      end

      def context
        @context ||= {
          writable: true,
          admin: true,
          viewer: User.find_by(username: params[:username])
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
