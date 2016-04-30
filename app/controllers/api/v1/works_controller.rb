# frozen_string_literal: true

module Api
  module V1
    class WorksController < Api::V1::ApplicationController
      before_action :prepare_params!, only: [:index]

      def index
        @works = Work.limit(2)
      end
    end
  end
end
