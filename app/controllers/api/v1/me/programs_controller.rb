# frozen_string_literal: true

module Api
  module V1
    module Me
      class ProgramsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(index)

        def index
          programs = current_user.
            programs.
            all.
            work_published.
            episode_published
          service = Api::V1::Me::ProgramIndexService.new(programs, @params)
          service.user = current_user
          @programs = service.result
        end
      end
    end
  end
end
