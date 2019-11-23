# frozen_string_literal: true

module Api
  module Internal
    class ActivitiesController < Api::Internal::ApplicationController
      def index
        return head(404) if params[:username].blank? && !user_signed_in?

        activities = if params[:username].blank?
          current_user.following_activities
        else
          User.where(username: params[:username]).first&.activities.presence || Activity.none
        end

        @user = current_user
        @activities = UserActivitiesQuery.new.call(
          activities: activities,
          user: current_user,
          page: params[:page]
        )

        @works = Work.without_deleted.where(id: @activities.all.map(&:work_id))
      end
    end
  end
end
