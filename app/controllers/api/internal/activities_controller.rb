# frozen_string_literal: true

module Api
  module Internal
    class ActivitiesController < Api::Internal::ApplicationController
      def index(username: nil, page: nil)
        return render(status: 404, nothing: true) if username.blank? && !user_signed_in?

        activities = if username.blank?
          current_user.following_activities
        else
          User.find_by(username: username).activities
        end

        @activities = activities.
          order(id: :desc).
          includes(:recipient, trackable: :user, user: :profile).
          page(page)
      end
    end
  end
end
