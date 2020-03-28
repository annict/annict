# frozen_string_literal: true

module V4
  module Db
    class SeriesController < V4::Db::ApplicationController
      def index
        @series_list = Series.without_deleted.order(id: :desc).page(params[:page]).per(100)
      end
    end
  end
end
