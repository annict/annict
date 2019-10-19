# frozen_string_literal: true

module Api
  module V1
    module Me
      class SlotsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(index)

        def index
          slots = current_user.
            slots.
            all.
            work_published.
            episode_published
          service = Api::V1::Me::SlotIndexService.new(slots, @params)
          service.user = current_user
          @slots = service.result
        end
      end
    end
  end
end
