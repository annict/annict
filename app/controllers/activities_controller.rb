# frozen_string_literal: true

class ActivitiesController < ApplicationController
  before_action :authenticate_user!

  def index
    return redirect_to root_path unless browser.device.mobile?

    activities = current_user.
      following_activities.
      order(id: :desc).
      includes(user: :profile).
      page(1)

    activity_data = render_jb("api/internal/activities/index",
      user: current_user,
      activities: activities)

    works = Work.where(id: activities.pluck(:work_id))
    work_list_data = render_jb "works/_list",
      user: current_user,
      works: works

    gon.push(
      activityData: activity_data,
      workListData: work_list_data
    )
  end
end
