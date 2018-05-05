# frozen_string_literal: true

module Api
  module V1
    class ReviewsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @records = Record.reviews.includes(:work).all
        @records = Api::V1::ReviewIndexService.new(@records, @params).result
      end
    end
  end
end
