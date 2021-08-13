# frozen_string_literal: true

module Api
  module V1
    module Me
      class WorksController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i[index]

        def index
          works = current_user.works.only_kept
          service = Api::V1::Me::WorkIndexService.new(works, @params)
          service.user = current_user
          @works = service.result
        end
      end
    end
  end
end
