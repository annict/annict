# frozen_string_literal: true

module Api
  module V1
    module Me
      class StatusesController < Api::V1::ApplicationController
        before_action :prepare_params!, only: [:create]

        def create
          work = Work.published.find(@params.work_id)
          status = StatusService.new(current_user, work, keen_client, ga_client)
          status.app = doorkeeper_token.application

          head(204) if status.change(@params.kind)
        end
      end
    end
  end
end
