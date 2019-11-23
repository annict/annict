# frozen_string_literal: true

module Api
  module V1
    class CastsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @casts = Cast.without_deleted
        @casts = Api::V1::CastIndexService.new(@casts, @params).result
      end
    end
  end
end
