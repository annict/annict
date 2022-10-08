# frozen_string_literal: true

module Api
  module V1
    module Me
      class FollowingActivitiesController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i[index]

        def index
          following_ids = current_user.followings.only_kept.pluck(:id)
          following_ids << current_user.id
          activities = Activity
            .where(user_id: following_ids)
            .includes(:itemable, user: :profile)
          service = Deprecated::Api::V1::Me::FollowingActivityIndexService.new(activities, @params)
          service.user = current_user
          @activities = service.result
        end
      end
    end
  end
end
