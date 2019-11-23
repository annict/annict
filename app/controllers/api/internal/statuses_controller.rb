# frozen_string_literal: true

module Api
  module Internal
    class StatusesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def select
        @work = Work.without_deleted.find(params[:work_id])
        page_category = params[:page_category]
        ga_client.page_category = page_category
        status = StatusService.new(current_user, @work)
        status.ga_client = ga_client
        status.via = "internal_api"
        status.page_category = page_category
        status.change!(params[:status_kind])
        head(200)
      end
    end
  end
end
