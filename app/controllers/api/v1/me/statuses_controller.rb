# frozen_string_literal: true

module Api
  module V1
    module Me
      class StatusesController < Api::V1::ApplicationController
        before_action :prepare_params!, only: [:create]

        def create
          work = Work.published.find(@params.work_id)
          status = StatusService.new(current_user, work, ga_client)
          status.app = doorkeeper_token.application

          render(status: 204, nothing: true) if status.change(@params.kind)
        end
      end
    end
  end
end
