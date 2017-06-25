# frozen_string_literal: true

module Api
  module V1
    class ReviewsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @reviews = Review.includes(:work).all
        @reviews = Api::V1::ReviewIndexService.new(@reviews, @params).result
      end
    end
  end
end
