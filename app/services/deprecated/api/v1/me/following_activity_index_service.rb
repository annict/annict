# frozen_string_literal: true

module Deprecated::Api
  module V1
    module Me
      class FollowingActivityIndexService < Deprecated::Api::V1::BaseService
        attr_writer :user

        def result
          @collection = filter_actions
          @collection = filter_muted
          @collection = sort_id
          @collection = @collection.page(@params.page).per(@params.per_page)
          @collection
        end

        private

        def filter_actions
          return @collection if @params.filter_actions.blank?
          @collection.where(action: @params.latest_filter_actions)
        end

        def filter_muted
          return @collection if @params.filter_muted == "false"
          mute_user_ids = @user.mute_users.pluck(:muted_user_id)
          @collection.where.not(user_id: mute_user_ids)
        end
      end
    end
  end
end
