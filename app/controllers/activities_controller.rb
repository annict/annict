# frozen_string_literal: true

class ActivitiesController < ApplicationController
  before_action :authenticate_user!

  def index
    return redirect_to root_path if device_pc?

    activities = UserActivitiesQuery.new.call(
      activities: current_user.following_activities,
      user: current_user,
      page: 1
    )
    works = Work.where(id: activities.all.map(&:work_id))

    activity_data = render_jb("api/internal/activities/index",
      user: current_user,
      activities: activities,
      works: works)

    gon.push(activityData: activity_data)
  end
end
