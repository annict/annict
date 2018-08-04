# frozen_string_literal: true

class HomeController < ApplicationController
  layout "v3/application"

  def show
    return show_for_member if user_signed_in?

    render :show_for_guest
  end

  private

  def show_for_member
    @tips = current_user.tips.unfinished.with_locale(current_user.locale)
    tips_data = render_jb("home/_tips", tips: @tips.limit(3))

    latest_statuses = TrackableService.new(current_user).latest_statuses
    latest_status_data = render_jb "api/internal/latest_statuses/index",
      user: current_user,
      latest_statuses: latest_statuses

    unless browser.device.mobile?
      @forum_posts = ForumPost.
        joins(:forum_category).
        merge(ForumCategory.with_slug(:site_news))
      @forum_posts = localable_resources(@forum_posts).order(created_at: :desc).limit(5)

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
    end

    gon.push(
      tipsData: tips_data,
      latestStatusData: latest_status_data,
      activityData: activity_data
    )

    render :index
  end
end
