# frozen_string_literal: true

module Api
  module V1
    module Me
      class FollowingActivitiesController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(index)

        def index
          following_ids = current_user.followings.without_deleted.pluck(:id)
          following_ids << current_user.id
          activities = Activity.
            where(user_id: following_ids).
            includes(:recipient, :trackable, user: :profile)
          service = Api::V1::Me::FollowingActivityIndexService.new(activities, @params)
          service.user = current_user
          @activities = service.result
        end
      end
    end
  end
end
