# frozen_string_literal: true

module Api
  module V1
    class WorksController < Api::V1::ApplicationController
      before_action :prepare_params!, only: [:index]

      def index
        @works = Work.only_kept
        @works = Api::V1::WorkIndexService.new(@works, @params).result
      end
    end
  end
end
