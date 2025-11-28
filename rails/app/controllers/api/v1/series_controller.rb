# typed: false
# frozen_string_literal: true

module Api
  module V1
    class SeriesController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @series_list = Series.only_kept
        @series_list = Deprecated::Api::V1::SeriesIndexService.new(@series_list, @params).result
      end
    end
  end
end
