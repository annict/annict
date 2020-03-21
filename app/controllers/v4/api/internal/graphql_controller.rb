# frozen_string_literal: true

module V4
  module Api
    module Internal
      class GraphqlController < ActionController::Base
        include V4::Localizable
        include V4::Loggable
        include V4::RavenContext

        skip_before_action :verify_authenticity_token
        before_action :set_raven_context
        around_action :set_locale

        def execute
          render json: execute(current_user)
        end

        def execute_local
          render json: execute(User.find_by(username: params[:username]))
        end

        private

        def execute(user)
          context = {
            writable: true,
            admin: true,
            viewer: user
          }

          ::Canary::AnnictSchema.execute(params[:query], variables: variables, context: context)
        end

        def variables
          @variables ||= ensure_hash(params[:variables])
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
end
