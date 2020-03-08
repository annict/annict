# frozen_string_literal: true

module API
  module V1
    class WorksController < API::V1::ApplicationController
      before_action :prepare_params!, only: [:index]

      def index
        @works = Work.without_deleted
        @works = API::V1::WorkIndexService.new(@works, @params).result
      end
    end
  end
end
