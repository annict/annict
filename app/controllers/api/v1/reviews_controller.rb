# typed: false
# frozen_string_literal: true

module Api
  module V1
    class ReviewsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @work_records = WorkRecord.includes(:work).all
        @work_records = Deprecated::Api::V1::WorkRecordIndexService.new(@work_records, @params).result
      end
    end
  end
end
