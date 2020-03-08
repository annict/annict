# frozen_string_literal: true

module API
  module V1
    module Me
      class StatusesController < API::V1::ApplicationController
        before_action :prepare_params!, only: [:create]

        def create
          work = Work.without_deleted.find(@params.work_id)
          status = StatusService.new(current_user, work)
          status.app = doorkeeper_token.application
          status.ga_client = ga_client
          status.via = "rest_api"

          status.change!(@params.kind)

          head 204
        end
      end
    end
  end
end
