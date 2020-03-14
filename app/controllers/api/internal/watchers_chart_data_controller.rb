# frozen_string_literal: true

module Api
  module Internal
    class WatchersChartDataController < Api::Internal::ApplicationController
      def show
        @work = Work.without_deleted.find(params[:work_id])
        render json: @work.watchers_chart_dataset
      end
    end
  end
end
