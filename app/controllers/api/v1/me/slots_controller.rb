# frozen_string_literal: true

module Api
  module V1
    module Me
      class SlotsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i[index]

        def index
          slots = Deprecated::UserSlotsQuery.new(
            current_user,
            Slot.only_kept.with_works(current_user.works_on(:wanna_watch, :watching).only_kept)
          ).call
          service = Deprecated::Api::V1::Me::SlotIndexService.new(slots, @params)
          service.user = current_user
          @slots = service.result
        end
      end
    end
  end
end
