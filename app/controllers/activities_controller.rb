# frozen_string_literal: true

class ActivitiesController < ApplicationController
  before_action :authenticate_user!

  def index
    return redirect_to root_path unless browser.device.mobile?

    activities = current_user.
      following_activities.
      order(id: :desc).
      includes(:recipient, trackable: :user, user: :profile).
      page(1)

    activity_data = render_jb("api/internal/activities/index",
      user: current_user,
      activities: activities)

    gon.push(activityData: activity_data)
  end
end
