# frozen_string_literal: true

module Api
  module Internal
    class ActivitiesController < Api::Internal::ApplicationController
      def index(username: nil, page: nil)
        return head(404) if username.blank? && !user_signed_in?

        activities = if username.blank?
          current_user.following_activities
        else
          User.where(username: username).first&.activities.presence || Activity.none
        end

        @activities = activities.
          order(id: :desc).
          includes(:recipient, trackable: :user, user: :profile).
          page(page)
      end
    end
  end
end
