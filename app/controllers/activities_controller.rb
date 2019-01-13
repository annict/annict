# frozen_string_literal: true

class ActivitiesController < ApplicationController
  before_action :authenticate_user!

  def index
    return redirect_to root_path if device_pc?

    activities = current_user.
      following_activities.
      order(id: :desc).
      includes(:work, user: :profile).
      page(1)
    works = Work.where(id: activities.pluck(:work_id))

    activity_data = render_jb("api/internal/activities/index",
      user: current_user,
      activities: activities,
      works: works)

    gon.push(activityData: activity_data)
  end
end
