# frozen_string_literal: true

module Api
  module Internal
    class StatusChartDataController < Api::Internal::ApplicationController
      def show
        @work = Work.only_kept.find(params[:work_id])
        render json: @work.status_chart_dataset
      end
    end
  end
end
