# frozen_string_literal: true

module Api
  module V1
    class ReviewsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @work_records = WorkRecord
          .eager_load(:record)
          .preload(record: [:work, user: :profile])
          .merge(Record.only_kept)
          .all
        @work_records = Api::V1::WorkRecordIndexService.new(@work_records, @params).result
      end
    end
  end
end
