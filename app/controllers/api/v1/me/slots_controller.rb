# frozen_string_literal: true

module API
  module V1
    module Me
      class SlotsController < API::V1::ApplicationController
        before_action :prepare_params!, only: %i(index)

        def index
          slots = UserSlotsQuery.new(
            current_user,
            Slot.without_deleted.with_works(current_user.works_on(:wanna_watch, :watching).without_deleted)
          ).call
          service = API::V1::Me::SlotIndexService.new(slots, @params)
          service.user = current_user
          @slots = service.result
        end
      end
    end
  end
end
