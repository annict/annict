# frozen_string_literal: true

module API
  module V1
    class ReviewsController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @work_records = WorkRecord.includes(:work).all
        @work_records = API::V1::WorkRecordIndexService.new(@work_records, @params).result
      end
    end
  end
end
