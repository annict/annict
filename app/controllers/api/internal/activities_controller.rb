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
        @activities = activities.
          order(id: :desc).
          includes(:work, user: :profile).
          merge(Work.published).
          page(params[:page])

        @works = Work.published.where(id: @activities.pluck(:work_id))
      end
    end
  end
end
