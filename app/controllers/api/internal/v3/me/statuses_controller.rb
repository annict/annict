# frozen_string_literal: true

module Api
  module Internal
    module V3
      module Me
        class StatusesController < Api::Internal::V3::ApplicationController
          before_action :authenticate_user!

          def show
            work = ::V3::FetchStatusQuery.new(
              user: current_user,
              gql_work_id: params[:gql_work_id]
            ).call

            render json: {
              status: work.viewer_status_state
            }
          end

          def update
            ::V3::UpdateStatusQuery.new(
              user: current_user,
              gql_work_id: params[:gql_work_id],
              status_kind: params[:status_kind]
            ).call

            head 204
          end
        end
      end
    end
  end
end
