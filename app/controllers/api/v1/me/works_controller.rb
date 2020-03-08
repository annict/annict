# frozen_string_literal: true

module API
  module V1
    module Me
      class WorksController < API::V1::ApplicationController
        before_action :prepare_params!, only: %i(index)

        def index
          works = current_user.works.all.without_deleted
          service = API::V1::Me::WorkIndexService.new(works, @params)
          service.user = current_user
          @works = service.result
        end
      end
    end
  end
end
