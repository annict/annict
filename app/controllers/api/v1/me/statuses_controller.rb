# frozen_string_literal: true

module Api
  module V1
    module Me
      class StatusesController < Api::V1::ApplicationController
        before_action :prepare_params!, only: [:create]

        def create
          work = Work.only_kept.find(@params.work_id)

          ChangeStatusService.new(user: current_user, work: work).call(status_kind: @params.kind)

          head 204
        end
      end
    end
  end
end
