# frozen_string_literal: true

module API
  module V1
    class CastsController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @casts = Cast.without_deleted
        @casts = API::V1::CastIndexService.new(@casts, @params).result
      end
    end
  end
end
