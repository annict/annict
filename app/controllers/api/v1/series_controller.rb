# frozen_string_literal: true

module API
  module V1
    class SeriesController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @series_list = Series.without_deleted
        @series_list = API::V1::SeriesIndexService.new(@series_list, @params).result
      end
    end
  end
end
