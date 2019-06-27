# frozen_string_literal: true

module Api
  module Internal
    module V3
      module Me
        class StatusesController < Api::Internal::V3::ApplicationController
          before_action :authenticate_user!

          def show
            data = ::V3::FetchStatusService.new(user: current_user, work_id: params[:work_id]).call.
              to_h.dig("data", "searchWorks", "nodes").first

            render json: {
              status: data["viewerStatusState"]
            }
          end

          def update
            ::V3::UpdateStatusService.new(user: current_user, work_id: params[:work_id], status_kind: params[:status_kind]).call
            head 204
          end
        end
      end
    end
  end
end
