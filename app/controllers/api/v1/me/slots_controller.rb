# frozen_string_literal: true

module Api
  module V1
    module Me
      class SlotsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(index)

        def index
          slots = UserSlotsQuery.new(
            current_user,
            Slot.without_deleted,
            status_kinds: %i(wanna_watch watching)
          ).call
          service = Api::V1::Me::SlotIndexService.new(slots, @params)
          service.user = current_user
          @slots = service.result
        end
      end
    end
  end
end
