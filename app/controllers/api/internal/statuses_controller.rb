# frozen_string_literal: true

module Api
  module Internal
    class StatusesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def select
        @work = Work.only_kept.find(params[:work_id])

        ChangeStatusService.new(user: current_user, work: @work).call(status_kind: params[:status_kind])

        head 200
      end
    end
  end
end
